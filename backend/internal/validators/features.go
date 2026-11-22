package validators

import (
	"slices"

	"ecommerce-backend/internal/models"

	"github.com/shopspring/decimal"
)

const (
	MaxOfferProducts = 500
	MaxBundleItems   = 20
)

// ---------- offers ----------

func ValidateOffer(in *models.OfferInput) error {
	Trim(&in.Name)
	Trim(&in.NameAr)
	Trim(&in.Description)
	Trim(&in.DescriptionAr)
	Trim(&in.DiscountType)
	Trim(&in.Scope)
	TrimOptional(&in.Category)

	c := New()
	c.Required("name", in.Name)
	c.MaxLen("name", in.Name, 150)
	c.Required("name_ar", in.NameAr)
	c.MaxLen("name_ar", in.NameAr, 150)

	switch in.DiscountType {
	case models.OfferPercentage:
		c.Add(in.DiscountValue.IsPositive() && in.DiscountValue.LessThanOrEqual(decimal.NewFromInt(100)),
			"discount_value", "must be more than 0 and at most 100 for a percentage offer")
	case models.OfferFixed:
		c.Add(in.DiscountValue.IsPositive(), "discount_value", "must be more than 0")
		c.Money("discount_value", in.DiscountValue)
	default:
		c.Add(false, "discount_type", "must be percentage or fixed")
	}
	c.Add(in.DiscountValue.Equal(in.DiscountValue.Round(2)), "discount_value", "can have at most 2 decimal places")

	switch in.Scope {
	case models.OfferScopeProducts:
		c.Add(len(in.ProductIDs) > 0, "product_ids", "choose at least one product")
		c.Add(len(in.ProductIDs) <= MaxOfferProducts, "product_ids", "too many products (maximum 500)")
		c.Add(in.Category == nil, "category", "must be empty when scope is products")
		for _, id := range in.ProductIDs {
			c.Add(id > 0, "product_ids", "must contain valid product IDs")
		}
		in.ProductIDs = dedupe(in.ProductIDs)
	case models.OfferScopeCategory:
		c.Add(in.Category != nil, "category", "is required when scope is category")
		c.OptionalMaxLen("category", in.Category, 100)
		c.Add(len(in.ProductIDs) == 0, "product_ids", "must be empty when scope is category")
	default:
		c.Add(false, "scope", "must be products or category")
	}

	c.Add(!in.StartsAt.IsZero(), "starts_at", "is required (RFC 3339, e.g. 2026-12-01T00:00:00+03:00)")
	c.Add(!in.EndsAt.IsZero(), "ends_at", "is required (RFC 3339, e.g. 2026-12-31T23:59:59+03:00)")
	c.Add(in.StartsAt.IsZero() || in.EndsAt.IsZero() || in.EndsAt.After(in.StartsAt), "ends_at", "must be after starts_at")
	return c.Err()
}

// ---------- bundles ----------

func ValidateBundle(in *models.BundleInput) error {
	Trim(&in.Name)
	Trim(&in.NameAr)
	Trim(&in.Description)
	Trim(&in.DescriptionAr)
	TrimOptional(&in.ImageURL)

	c := New()
	c.Required("name", in.Name)
	c.MaxLen("name", in.Name, 255)
	c.Required("name_ar", in.NameAr)
	c.MaxLen("name_ar", in.NameAr, 255)
	checkOptionalURL(c, "image_url", in.ImageURL)
	c.Money("price", in.Price)
	c.Add(in.StartsAt == nil || in.EndsAt == nil || in.EndsAt.After(*in.StartsAt), "ends_at", "must be after starts_at")

	c.Add(len(in.Items) >= 2, "items", "a bundle needs at least 2 products (or one product with quantity 2+)")
	if len(in.Items) == 1 && in.Items[0].Quantity >= 2 {
		delete(c.fields, "items") // e.g. "3 for the price of 2" is a valid bundle
	}
	c.Add(len(in.Items) <= MaxBundleItems, "items", "too many products (maximum 20)")
	seen := map[int64]bool{}
	for i := range in.Items {
		it := &in.Items[i]
		TrimOptional(&it.SelectedColor)
		c.Add(it.ProductID > 0, "items", "every item needs a valid product_id")
		c.Add(it.Quantity >= 1 && it.Quantity <= 100, "items", "every item quantity must be between 1 and 100")
		c.OptionalMaxLen("items", it.SelectedColor, 50)
		c.Add(!seen[it.ProductID], "items", "a product can only appear once; raise its quantity instead")
		seen[it.ProductID] = true
	}
	return c.Err()
}

func ValidateAddCartBundle(req *models.AddCartBundleRequest) error {
	c := New()
	c.Add(req.BundleID > 0, "bundle_id", "is required")
	ValidateQuantity(c, req.Quantity)
	return c.Err()
}

// ---------- debts ----------

func ValidateCreditLimit(req *models.SetCreditLimitRequest) error {
	c := New()
	c.Money("credit_limit", req.CreditLimit)
	return c.Err()
}

func ValidatePayment(req *models.RecordPaymentRequest) error {
	TrimOptional(&req.Note)
	c := New()
	c.Add(req.Amount.IsPositive(), "amount", "must be more than 0")
	c.Money("amount", req.Amount)
	c.OptionalMaxLen("note", req.Note, 1000)
	c.Add(req.OrderID == nil || *req.OrderID > 0, "order_id", "must be a valid order ID")
	return c.Err()
}

func ValidateAdjustment(req *models.RecordAdjustmentRequest) error {
	Trim(&req.Note)
	c := New()
	c.Add(!req.Amount.IsZero(), "amount", "cannot be 0 (positive increases the debt, negative reduces it)")
	c.Money("amount", req.Amount.Abs())
	c.Required("note", req.Note)
	c.MaxLen("note", req.Note, 1000)
	c.Add(req.OrderID == nil || *req.OrderID > 0, "order_id", "must be a valid order ID")
	return c.Err()
}

func dedupe(ids []int64) []int64 {
	out := slices.Clone(ids)
	slices.Sort(out)
	return slices.Compact(out)
}
