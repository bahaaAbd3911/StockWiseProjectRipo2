package services

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/shopspring/decimal"

	"ecommerce-backend/internal/apperr"
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/repository"
	"ecommerce-backend/internal/storage"
	"ecommerce-backend/internal/validators"
)

// SettingsService: a store's name, logo, theme colors and page content
// (banners, feature tiles, footer). Each store has its own settings; the
// same React components render every store from them.
type SettingsService struct {
	settings *repository.SettingsRepository
	files    storage.Storage // banner images (nil: uploads off)
}

func NewSettingsService(s *repository.SettingsRepository, files storage.Storage) *SettingsService {
	return &SettingsService{settings: s, files: files}
}

// MaxSettingsBody is the request limit for saving settings: a 1 MB logo file
// grows by a third when embedded as base64.
const MaxSettingsBody = 2 << 20

const maxLogoLength = 1_500_000 // characters of the data URL

var (
	logoDataURL = regexp.MustCompile(`^data:image/(png|jpeg|jpg|webp|gif|svg\+xml);base64,[A-Za-z0-9+/]+={0,2}$`)
	colorToken  = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9]{0,39}$`)
	// Font Awesome classes only: nothing else can end up in a class attribute.
	iconClass = regexp.MustCompile(`^fa-(solid|regular|brands) fa-[a-z0-9-]{1,40}$`)
)

// Social networks a footer may link to.
var socialNetworks = map[string]bool{
	"facebook": true, "instagram": true, "x": true, "youtube": true, "tiktok": true, "whatsapp": true, "snapchat": true, "linkedin": true,
}

func (s *SettingsService) Get(ctx context.Context) (models.StoreSettings, error) {
	return s.settings.Get(ctx)
}

// Save validates and stores the settings, then returns what was stored.
// Uploaded banner images that the new settings no longer use are deleted.
func (s *SettingsService) Save(ctx context.Context, in models.StoreSettings) (models.StoreSettings, error) {
	in.StoreName = strings.TrimSpace(in.StoreName)
	in.Logo = strings.TrimSpace(in.Logo)

	c := validators.New()
	c.MaxLen("storeName", in.StoreName, 100)
	switch {
	case in.Logo == "":
	case strings.HasPrefix(in.Logo, "data:"):
		c.Add(len(in.Logo) <= maxLogoLength, "logo", "must be under 1 MB")
		c.Add(logoDataURL.MatchString(in.Logo), "logo", "must be a PNG, JPG, WebP, GIF or SVG image")
	default:
		c.Add(validators.IsHTTPURL(in.Logo), "logo", "must be an image file or an http(s) URL")
	}
	for mode, tokens := range in.Colors {
		if mode != "light" && mode != "dark" {
			c.Add(false, "colors", "only light and dark modes are allowed")
			continue
		}
		c.Add(len(tokens) <= 40, "colors", "too many colors")
		for key, value := range tokens {
			c.Add(colorToken.MatchString(key), "colors", "color names may only contain letters and digits")
			c.Add(validators.IsHexColor(value), "colors", "every color must look like #RRGGBB")
		}
	}
	checkContent(c, in.Content)
	checkShipping(c, in.Shipping)
	if err := c.Err(); err != nil {
		return models.StoreSettings{}, err
	}
	if in.Colors == nil {
		in.Colors = map[string]map[string]string{}
	}

	old, err := s.settings.Get(ctx)
	if err != nil {
		return models.StoreSettings{}, err
	}
	if err := s.settings.Save(ctx, in); err != nil {
		return models.StoreSettings{}, apperr.FromDB(err, "settings not found")
	}
	s.deleteUnused(ctx, old.Content.Images(), in.Content.Images())
	return in, nil
}

// UploadImage stores a banner image and returns its URL; the admin then saves
// the settings with it. (Images are checked and stripped of metadata.)
func (s *SettingsService) UploadImage(ctx context.Context, data []byte) (string, error) {
	return saveImage(ctx, s.files, data)
}

func (s *SettingsService) deleteUnused(ctx context.Context, before, after []string) {
	if s.files == nil {
		return
	}
	keep := map[string]bool{}
	for _, u := range after {
		keep[u] = true
	}
	for _, u := range before {
		if !keep[u] {
			// Only this store's uploads are ever deleted (storage checks it).
			if err := s.files.Delete(ctx, u); err != nil {
				slog.Error("delete unused banner image", "error", err)
			}
		}
	}
}

// Limits for the shipping settings: a delivery fee is a small amount, and a
// store lists a handful of cities.
const (
	maxShippingCities = 50
	maxShippingFee    = 1000
)

// checkShipping validates (and tidies) the shipping settings: fees between 0
// and 1000 with 2 decimals, a positive "free over" amount, and each city at
// most once.
func checkShipping(c *validators.Checker, in *models.ShippingSettings) {
	if in == nil {
		return
	}
	maxFee := decimal.NewFromInt(maxShippingFee)
	fee := func(field string, v decimal.Decimal) {
		c.Money(field, v)
		c.Add(v.LessThanOrEqual(maxFee), field, "must be at most 1000")
	}
	fee("shipping.fee", in.Fee)
	if in.FreeOver != nil {
		c.Money("shipping.freeOver", *in.FreeOver)
		c.Add(in.FreeOver.IsPositive(), "shipping.freeOver", "must be more than 0 (or empty for never)")
	}
	c.Add(len(in.Cities) <= maxShippingCities, "shipping.cities", "too many cities (maximum 50)")
	seen := map[string]bool{}
	for i := range in.Cities {
		city := &in.Cities[i]
		city.City = strings.TrimSpace(city.City)
		c.Add(city.City != "", "shipping.cities", "every city needs a name")
		c.MaxLen("shipping.cities", city.City, 100)
		key := strings.ToLower(city.City)
		c.Add(!seen[key], "shipping.cities", city.City+" is listed twice")
		seen[key] = true
		fee("shipping.cities", city.Fee)
	}
}

func checkContent(c *validators.Checker, in *models.StoreContent) {
	if in == nil {
		return
	}
	for name, b := range map[string]*models.Banner{"content.hero": in.Hero, "content.productsBanner": in.ProductsBanner} {
		if b == nil {
			continue
		}
		for _, f := range []*string{&b.Title, &b.TitleAr, &b.Subtitle, &b.SubtitleAr, &b.Highlight, &b.HighlightAr,
			&b.ButtonText, &b.ButtonTextAr, &b.Image} {
			*f = strings.TrimSpace(*f)
		}
		c.MaxLen(name+".title", b.Title+b.TitleAr, 300)
		c.MaxLen(name+".subtitle", b.Subtitle+b.SubtitleAr+b.Highlight+b.HighlightAr, 800)
		c.MaxLen(name+".buttonText", b.ButtonText+b.ButtonTextAr, 80)
		c.Add(b.Image == "" || validators.IsHTTPURL(b.Image), name+".image", "must be an uploaded image or an http(s) URL")
	}

	c.Add(len(in.Features) <= 8, "content.features", "at most 8 tiles")
	for i := range in.Features {
		f := &in.Features[i]
		for _, v := range []*string{&f.Icon, &f.Name, &f.NameAr, &f.Comment, &f.CommentAr} {
			*v = strings.TrimSpace(*v)
		}
		field := fmt.Sprintf("content.features.%d", i)
		c.Add(iconClass.MatchString(f.Icon), field+".icon", "must be a Font Awesome icon such as fa-solid fa-truck")
		c.Add(f.Name != "" || f.NameAr != "", field+".name", "is required")
		c.MaxLen(field+".name", f.Name+f.NameAr, 120)
		c.MaxLen(field+".comment", f.Comment+f.CommentAr, 200)
	}

	if f := in.Footer; f != nil {
		for _, v := range []*string{&f.Tagline, &f.TaglineAr, &f.Copyright, &f.CopyrightAr} {
			*v = strings.TrimSpace(*v)
		}
		c.MaxLen("content.footer.tagline", f.Tagline+f.TaglineAr, 300)
		c.MaxLen("content.footer.copyright", f.Copyright+f.CopyrightAr, 300)
		for network, link := range f.Social {
			c.Add(socialNetworks[network], "content.footer.social", "unknown network "+network)
			c.Add(link == "" || validators.IsHTTPURL(link), "content.footer.social", "every link must be an http(s) URL")
		}
	}
}
