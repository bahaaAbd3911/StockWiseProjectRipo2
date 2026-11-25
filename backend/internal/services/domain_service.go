package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net"
	"regexp"
	"strings"

	"ecommerce-backend/internal/apperr"
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/repository"
	"ecommerce-backend/internal/tenant"
	"ecommerce-backend/internal/validators"
)

// LookupTXT finds DNS TXT records. A variable so tests can stand in for DNS.
var LookupTXT = net.DefaultResolver.LookupTXT

// TXTPrefix is the name of the ownership record: _stockwise-verify.<domain>.
const TXTPrefix = "_stockwise-verify."

var hostnamePattern = regexp.MustCompile(`^([a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,63}$`)

// DomainService: a store admin connects their own domain names.
//
//  1. add the domain -> a random token
//  2. publish TXT  _stockwise-verify.<domain>  "stockwise-verify=<token>"
//     and point the domain at the platform (CNAME to the store's address)
//  3. verify -> the API reads the TXT record; when it matches, the domain is
//     verified: visitors on it see this store, and HTTPS certificates may be
//     issued for it (Caddy asks /internal/tls-check first)
type DomainService struct {
	domains        *repository.DomainRepository
	resolver       *tenant.Resolver
	platformDomain string
	cnameTarget    func(tenant.Store) string // the store's platform address, e.g. ahmad.myplatform.com
}

func NewDomainService(domains *repository.DomainRepository, resolver *tenant.Resolver, platformDomain string,
	cnameTarget func(tenant.Store) string) *DomainService {
	return &DomainService{domains: domains, resolver: resolver, platformDomain: platformDomain, cnameTarget: cnameTarget}
}

func (s *DomainService) withDNS(ctx context.Context, d models.StoreDomain) models.StoreDomain {
	store, _ := tenant.FromContext(ctx)
	d.DNS = models.DomainDNS{
		TXTName:     TXTPrefix + d.Domain,
		TXTValue:    "stockwise-verify=" + d.VerifyToken,
		CNAMETarget: s.cnameTarget(store),
	}
	return d
}

func (s *DomainService) List(ctx context.Context) ([]models.StoreDomain, error) {
	list, err := s.domains.List(ctx)
	for i := range list {
		list[i] = s.withDNS(ctx, list[i])
	}
	return list, err
}

// Add connects a domain (not yet verified) and returns the DNS records to set.
func (s *DomainService) Add(ctx context.Context, req models.AddDomainRequest) (models.StoreDomain, error) {
	domain := normalizeDomain(req.Domain)
	if !hostnamePattern.MatchString(domain) || len(domain) > 253 {
		return models.StoreDomain{}, validators.FieldError("domain", "must be a domain name such as shop.example.com (no http:// or path)")
	}
	if s.platformDomain != "" && (domain == s.platformDomain || strings.HasSuffix(domain, "."+s.platformDomain)) {
		return models.StoreDomain{}, validators.FieldError("domain", "addresses on the platform's own domain cannot be added")
	}
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return models.StoreDomain{}, err
	}
	id, err := s.domains.Add(ctx, domain, hex.EncodeToString(b))
	if err != nil {
		return models.StoreDomain{}, err
	}
	d, err := s.domains.Get(ctx, id)
	return s.withDNS(ctx, d), err
}

// Verify checks the ownership TXT record. On success the domain starts
// serving this store at once (the host cache is cleared).
func (s *DomainService) Verify(ctx context.Context, id int64) (models.StoreDomain, error) {
	d, err := s.domains.Get(ctx, id)
	if err != nil {
		return d, err
	}
	if !d.Verified {
		want := "stockwise-verify=" + d.VerifyToken
		records, _ := LookupTXT(ctx, TXTPrefix+d.Domain)
		found := false
		for _, r := range records {
			if strings.TrimSpace(r) == want {
				found = true
			}
		}
		if !found {
			return models.StoreDomain{}, apperr.Conflict("the TXT record was not found yet; DNS changes can take up to a few hours - try again later")
		}
		if err := s.domains.MarkVerified(ctx, id); err != nil {
			return models.StoreDomain{}, err
		}
		s.resolver.Forget()
	}
	d, err = s.domains.Get(ctx, id)
	return s.withDNS(ctx, d), err
}

func (s *DomainService) SetPrimary(ctx context.Context, id int64) (models.StoreDomain, error) {
	if err := s.domains.SetPrimary(ctx, id); err != nil {
		return models.StoreDomain{}, err
	}
	s.resolver.Forget()
	d, err := s.domains.Get(ctx, id)
	return s.withDNS(ctx, d), err
}

func (s *DomainService) Delete(ctx context.Context, id int64) error {
	if err := s.domains.Delete(ctx, id); err != nil {
		return err
	}
	s.resolver.Forget()
	return nil
}

func normalizeDomain(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.TrimPrefix(strings.TrimPrefix(s, "https://"), "http://")
	if i := strings.IndexAny(s, "/?#"); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSuffix(s, ".")
}
