// Package storage keeps uploaded files and returns the public URL of each.
//
//	STORAGE_DRIVER=local (default)  files in UPLOAD_DIR, served by the API at
//	                                PUBLIC_URL/uploads/<name>. Use a persistent
//	                                disk or Docker volume.
//	STORAGE_DRIVER=s3               any S3-compatible bucket: AWS S3,
//	                                Cloudflare R2, Backblaze B2, MinIO,
//	                                DigitalOcean Spaces. Needed on hosts whose
//	                                disk is wiped on every deploy.
//
// File names are random (32 hex characters + extension), never taken from the
// upload, so a client cannot overwrite another file or pick a dangerous name.
package storage

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Storage saves and deletes uploaded files.
type Storage interface {
	// Save stores data under a new random name and returns its public URL.
	Save(ctx context.Context, ext, contentType string, data []byte) (string, error)
	// Delete removes a file previously returned by Save. URLs that this
	// storage did not create (e.g. an image URL an admin typed in) are ignored.
	Delete(ctx context.Context, fileURL string) error
}

// Config selects and configures the storage.
type Config struct {
	Driver    string // local | s3
	Dir       string // local: UPLOAD_DIR
	PublicURL string // local: the API's public address (PUBLIC_URL)
	S3        S3Config
}

type S3Config struct {
	Endpoint        string // S3_ENDPOINT, host[:port], e.g. s3.eu-central-1.amazonaws.com or <account>.r2.cloudflarestorage.com
	Region          string // S3_REGION (R2: auto)
	Bucket          string // S3_BUCKET
	AccessKeyID     string // S3_ACCESS_KEY_ID
	SecretAccessKey string // S3_SECRET_ACCESS_KEY
	PublicURL       string // S3_PUBLIC_URL: where the bucket is readable (CDN, r2.dev, bucket website)
	Insecure        bool   // S3_INSECURE=true: plain HTTP (local MinIO only)
}

// New builds the storage described by cfg.
func New(cfg Config) (Storage, error) {
	switch cfg.Driver {
	case "", "local":
		if err := os.MkdirAll(cfg.Dir, 0o755); err != nil {
			return nil, fmt.Errorf("UPLOAD_DIR: %w", err)
		}
		return &Local{dir: cfg.Dir, base: strings.TrimRight(cfg.PublicURL, "/") + "/uploads/"}, nil
	case "s3":
		c := cfg.S3
		if c.Endpoint == "" || c.Bucket == "" || c.AccessKeyID == "" || c.SecretAccessKey == "" || c.PublicURL == "" {
			return nil, errors.New("STORAGE_DRIVER=s3 needs S3_ENDPOINT, S3_BUCKET, S3_ACCESS_KEY_ID, S3_SECRET_ACCESS_KEY and S3_PUBLIC_URL")
		}
		client, err := minio.New(c.Endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(c.AccessKeyID, c.SecretAccessKey, ""),
			Secure: !c.Insecure,
			Region: c.Region,
		})
		if err != nil {
			return nil, fmt.Errorf("S3: %w", err)
		}
		return &S3{client: client, bucket: c.Bucket, base: strings.TrimRight(c.PublicURL, "/") + "/"}, nil
	default:
		return nil, fmt.Errorf("unknown STORAGE_DRIVER %q (use local or s3)", cfg.Driver)
	}
}

// fileName matches the names Save creates; nothing else is ever served or deleted.
var fileName = regexp.MustCompile(`^[0-9a-f]{32}\.(jpg|png|webp)$`)

func newName(ext string) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b) + ext, nil
}

// nameIn returns the file name if fileURL was created under base.
func nameIn(base, fileURL string) (string, bool) {
	name, ok := strings.CutPrefix(fileURL, base)
	if !ok || !fileName.MatchString(name) {
		return "", false
	}
	return name, true
}

// ---------- local disk ----------

type Local struct {
	dir  string
	base string // PUBLIC_URL/uploads/
}

func (l *Local) Save(_ context.Context, ext, _ string, data []byte) (string, error) {
	name, err := newName(ext)
	if err != nil {
		return "", err
	}
	// Write to a temporary name first so a half-written file is never served.
	tmp := filepath.Join(l.dir, name+".tmp")
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return "", err
	}
	if err := os.Rename(tmp, filepath.Join(l.dir, name)); err != nil {
		_ = os.Remove(tmp)
		return "", err
	}
	return l.base + name, nil
}

func (l *Local) Delete(_ context.Context, fileURL string) error {
	name, ok := nameIn(l.base, fileURL)
	if !ok {
		return nil
	}
	if err := os.Remove(filepath.Join(l.dir, name)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

var contentTypes = map[string]string{".jpg": "image/jpeg", ".png": "image/png", ".webp": "image/webp"}

// Handler serves GET /uploads/{name}. Only names created by Save are served,
// never directory listings. Names are random and files never change, so
// browsers and CDNs may cache them for a year.
func (l *Local) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		if !fileName.MatchString(name) {
			http.NotFound(w, r)
			return
		}
		f, err := os.Open(filepath.Join(l.dir, name))
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer f.Close()
		info, err := f.Stat()
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", contentTypes[filepath.Ext(name)])
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		// Images may be shown by the frontend on another origin.
		w.Header().Set("Cross-Origin-Resource-Policy", "cross-origin")
		http.ServeContent(w, r, name, info.ModTime(), f)
	})
}

// ---------- S3-compatible ----------

type S3 struct {
	client *minio.Client
	bucket string
	base   string // S3_PUBLIC_URL/
}

const s3Prefix = "uploads/"

func (s *S3) Save(ctx context.Context, ext, contentType string, data []byte) (string, error) {
	name, err := newName(ext)
	if err != nil {
		return "", err
	}
	_, err = s.client.PutObject(ctx, s.bucket, s3Prefix+name, bytes.NewReader(data), int64(len(data)),
		minio.PutObjectOptions{ContentType: contentType, CacheControl: "public, max-age=31536000, immutable"})
	if err != nil {
		return "", fmt.Errorf("upload to S3: %w", err)
	}
	return s.base + s3Prefix + name, nil
}

func (s *S3) Delete(ctx context.Context, fileURL string) error {
	rest, ok := strings.CutPrefix(fileURL, s.base+s3Prefix)
	if !ok || !fileName.MatchString(rest) {
		return nil
	}
	return s.client.RemoveObject(ctx, s.bucket, s3Prefix+rest, minio.RemoveObjectOptions{})
}

// ValidPublicURL is used by the config checks.
func ValidPublicURL(s string) bool {
	u, err := url.Parse(s)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}
