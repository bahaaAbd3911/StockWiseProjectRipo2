package validators

import (
	"regexp"
	"strings"

	"ecommerce-backend/internal/models"

	"github.com/shopspring/decimal"
)

// Codes are what customers type: letters, digits, - and _, 3 to 40 long.
var couponCodePattern = regexp.MustCompile(`^[A-Z0-9_-]{3,40}$`)

// NormalizeCouponCode is how a typed code is compared: trimmed, upper-case.
func NormalizeCouponCode(s string) string { return strings.ToUpper(strings.TrimSpace(s)) }

// ValidateCoupon normalises and checks an admin coupon body.
func ValidateCoupon(in *models.CouponInput) error {
	in.Code = NormalizeCouponCode(in.Code)
	Trim(&in.Description)
	Trim(&in.DescriptionAr)
	Trim(&in.DiscountType)
	in.Scope = strings.ToLower(strings.TrimSpace(in.Scope))
	if in.Scope == "" {
		in.Scope = models.CouponScopeAll
	}
	TrimOptional(&in.Category)
	if !in.PerCustomerLimitSet {
		one := 1
		in.PerCustomerLimit = &one
	}

	c := New()
	c.Add(couponCodePattern.MatchString(in.Code), "code", "must be 3-40 letters, digits, - or _ (e.g. SUMMER25)")
	c.MaxLen("description", in.Description, 500)
	c.MaxLen("description_ar", in.DescriptionAr, 500)

	switch in.DiscountType {
	case models.OfferPercentage:
		c.Add(in.DiscountValue.IsPositive() && in.DiscountValue.LessThanOrEqual(decimal.NewFromInt(100)),
			"discount_value", "must be more than 0 and at most 100 for a percentage coupon")
	case models.OfferFixed:
		c.Add(in.DiscountValue.IsPositive(), "discount_value", "must be more than 0")
		c.Money("discount_value", in.DiscountValue)
		c.Add(in.MaxDiscount == nil, "max_discount", "is only used with percentage coupons")
	default:
		c.Add(false, "discount_type", "must be percentage or fixed")
	}
	c.Add(in.DiscountValue.Equal(in.DiscountValue.Round(2)), "discount_value", "can have at most 2 decimal places")

	switch in.Scope {
	case models.CouponScopeAll:
		c.Add(len(in.ProductIDs) == 0, "product_ids", "must be empty when scope is all")
		c.Add(in.Category == nil, "category", "must be empty when scope is all")
	case models.CouponScopeProducts:
		c.Add(len(in.ProductIDs) > 0, "product_ids", "choose at least one product")
		c.Add(len(in.ProductIDs) <= MaxOfferProducts, "product_ids", "too many products (maximum 500)")
		c.Add(in.Category == nil, "category", "must be empty when scope is products")
		for _, id := range in.ProductIDs {
			c.Add(id > 0, "product_ids", "must contain valid product IDs")
		}
		in.ProductIDs = dedupe(in.ProductIDs)
	case models.CouponScopeCategory:
		c.Add(in.Category != nil, "category", "is required when scope is category")
		c.OptionalMaxLen("category", in.Category, 100)
		c.Add(len(in.ProductIDs) == 0, "product_ids", "must be empty when scope is category")
	default:
		c.Add(false, "scope", "must be all, products or category")
	}

	c.Add(!in.MinOrderAmount.IsNegative(), "min_order_amount", "cannot be negative")
	c.Money("min_order_amount", in.MinOrderAmount)
	if in.MaxDiscount != nil {
		c.Add(in.MaxDiscount.IsPositive(), "max_discount", "must be more than 0 (or null for no cap)")
		c.Money("max_discount", *in.MaxDiscount)
	}
	if in.StartsAt != nil && in.EndsAt != nil {
		c.Add(in.EndsAt.After(*in.StartsAt), "ends_at", "must be after starts_at")
	}
	if in.UsageLimit != nil {
		c.Add(*in.UsageLimit > 0, "usage_limit", "must be at least 1 (or null for unlimited)")
	}
	if in.PerCustomerLimit != nil {
		c.Add(*in.PerCustomerLimit > 0, "per_customer_limit", "must be at least 1 (or null for unlimited)")
	}
	return c.Err()
}
