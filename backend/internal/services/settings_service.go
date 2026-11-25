package services

import (
	"context"
	"regexp"
	"strings"

	"ecommerce-backend/internal/apperr"
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/repository"
	"ecommerce-backend/internal/validators"
)

// SettingsService: the store name, logo and theme colors.
type SettingsService struct {
	settings *repository.SettingsRepository
}

func NewSettingsService(s *repository.SettingsRepository) *SettingsService {
	return &SettingsService{settings: s}
}

// MaxSettingsBody is the request limit for saving settings: a 1 MB logo file
// grows by a third when embedded as base64.
const MaxSettingsBody = 2 << 20

const maxLogoLength = 1_500_000 // characters of the data URL

var (
	logoDataURL = regexp.MustCompile(`^data:image/(png|jpeg|jpg|webp|gif|svg\+xml);base64,[A-Za-z0-9+/]+={0,2}$`)
	colorToken  = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9]{0,39}$`)
)

func (s *SettingsService) Get(ctx context.Context) (models.StoreSettings, error) {
	return s.settings.Get(ctx)
}

// Save validates and stores the settings, then returns what was stored.
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
	if err := c.Err(); err != nil {
		return models.StoreSettings{}, err
	}
	if in.Colors == nil {
		in.Colors = map[string]map[string]string{}
	}
	if err := s.settings.Save(ctx, in); err != nil {
		return models.StoreSettings{}, apperr.FromDB(err, "settings not found")
	}
	return in, nil
}
