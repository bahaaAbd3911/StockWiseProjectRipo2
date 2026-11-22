package validators

import (
	"strings"
	"testing"
	"time"

	"ecommerce-backend/internal/apperr"
	"ecommerce-backend/internal/models"

	"github.com/shopspring/decimal"
)

func fields(t *testing.T, err error) map[string]string {
	t.Helper()
	if err == nil {
		return nil
	}
	e, ok := apperr.As(err)
	if !ok || e.Kind != apperr.KindInvalid {
		t.Fatalf("expected a validation error, got %v", err)
	}
	return e.Fields
}

func TestValidateRegisterNormalisesAndChecks(t *testing.T) {
	req := models.RegisterRequest{
		FirstName: "  Sara ", LastName: "Khalil", Email: "  Sara@Example.COM ",
		Password: "password123", Phone: "+962 79 123 4567",
	}
	if err := ValidateRegister(&req); err != nil {
		t.Fatalf("valid registration rejected: %v", err)
	}
	if req.Email != "sara@example.com" || req.FirstName != "Sara" {
		t.Fatalf("input not normalised: %+v", req)
	}

	bad := models.RegisterRequest{Email: "not-an-email", Password: "short", Phone: "abc"}
	f := fields(t, ValidateRegister(&bad))
	for _, k := range []string{"first_name", "last_name", "email", "password", "phone"} {
		if f[k] == "" {
			t.Errorf("expected an error for %s, got %v", k, f)
		}
	}

	long := req
	long.Password = strings.Repeat("x", 73)
	if fields(t, ValidateRegister(&long))["password"] == "" {
		t.Error("passwords over 72 bytes must be rejected (bcrypt limit)")
	}
}

func TestValidateProduct(t *testing.T) {
	valid := func() models.ProductInput {
		return models.ProductInput{Name: "Laptop", NameAr: "حاسوب", Price: decimal.RequireFromString("999.99"),
			DiscountValue: decimal.RequireFromString("10"), Stock: 5}
	}
	in := valid()
	if err := ValidateProduct(&in); err != nil {
		t.Fatalf("valid product rejected: %v", err)
	}

	cases := map[string]func(*models.ProductInput){
		"name":           func(p *models.ProductInput) { p.Name = "  " },
		"name_ar":        func(p *models.ProductInput) { p.NameAr = "" },
		"price":          func(p *models.ProductInput) { p.Price = decimal.RequireFromString("-1") },
		"discount_value": func(p *models.ProductInput) { p.DiscountValue = decimal.RequireFromString("101") },
		"stock":          func(p *models.ProductInput) { p.Stock = -1 },
	}
	for field, mutate := range cases {
		p := valid()
		mutate(&p)
		if fields(t, ValidateProduct(&p))[field] == "" {
			t.Errorf("expected an error for %s", field)
		}
	}

	p := valid()
	p.Price = decimal.RequireFromString("10.999")
	if fields(t, ValidateProduct(&p))["price"] == "" {
		t.Error("prices with more than 2 decimals must be rejected")
	}
}

func TestMaxLenCountsCharactersNotBytes(t *testing.T) {
	c := New()
	c.MaxLen("name", strings.Repeat("ع", 50), 50) // 50 Arabic letters = 100 bytes
	if c.Err() != nil {
		t.Fatal("50 Arabic characters must fit in varchar(50)")
	}
}

func TestValidateColorAndAddress(t *testing.T) {
	color := models.ProductColorInput{Name: "Black", NameAr: "أسود", HexCode: "#00000G"}
	if fields(t, ValidateProductColor(&color))["hex_code"] == "" {
		t.Error("invalid hex code accepted")
	}

	empty := "   "
	addr := models.AddressInput{RecipientName: "Sara", Phone: "0791234567", City: "Amman", Area: "Abdoun",
		Street: "Main St", BuildingNumber: &empty}
	if err := ValidateAddress(&addr); err != nil {
		t.Fatalf("valid address rejected: %v", err)
	}
	if addr.BuildingNumber != nil {
		t.Error("blank optional field should become nil (NULL)")
	}
}

func TestValidateOffer(t *testing.T) {
	start := time.Now()
	valid := func() models.OfferInput {
		return models.OfferInput{Name: "Eid", NameAr: "عيد", DiscountType: models.OfferPercentage,
			DiscountValue: decimal.RequireFromString("20"), Scope: models.OfferScopeProducts,
			ProductIDs: []int64{3, 1, 3}, StartsAt: start, EndsAt: start.Add(time.Hour)}
	}
	in := valid()
	if err := ValidateOffer(&in); err != nil {
		t.Fatalf("valid offer rejected: %v", err)
	}
	if len(in.ProductIDs) != 2 {
		t.Errorf("duplicate product IDs should be removed: %v", in.ProductIDs)
	}
	cat := "Phones"
	cases := map[string]func(*models.OfferInput){
		"discount_value": func(o *models.OfferInput) { o.DiscountValue = decimal.RequireFromString("120") },
		"discount_type":  func(o *models.OfferInput) { o.DiscountType = "bogo" },
		"product_ids":    func(o *models.OfferInput) { o.ProductIDs = nil },
		"category":       func(o *models.OfferInput) { o.Category = &cat }, // category with products scope
		"ends_at":        func(o *models.OfferInput) { o.EndsAt = o.StartsAt.Add(-time.Minute) },
	}
	for field, mutate := range cases {
		o := valid()
		mutate(&o)
		if fields(t, ValidateOffer(&o))[field] == "" {
			t.Errorf("expected an error for %s", field)
		}
	}
}

func TestValidateBundle(t *testing.T) {
	in := models.BundleInput{Name: "Set", NameAr: "طقم", Price: decimal.RequireFromString("10"),
		Items: []models.BundleItemInput{{ProductID: 1, Quantity: 3}}}
	if err := ValidateBundle(&in); err != nil {
		t.Fatalf("one product x3 is a valid bundle: %v", err)
	}
	in.Items = []models.BundleItemInput{{ProductID: 1, Quantity: 1}}
	if fields(t, ValidateBundle(&in))["items"] == "" {
		t.Error("a single product x1 is not a bundle")
	}
	in.Items = []models.BundleItemInput{{ProductID: 1, Quantity: 1}, {ProductID: 1, Quantity: 2}}
	if fields(t, ValidateBundle(&in))["items"] == "" {
		t.Error("the same product twice must be rejected")
	}
}

func TestValidateDebtRequests(t *testing.T) {
	pay := models.RecordPaymentRequest{Amount: decimal.RequireFromString("-5")}
	if fields(t, ValidatePayment(&pay))["amount"] == "" {
		t.Error("negative payment accepted")
	}
	adj := models.RecordAdjustmentRequest{Amount: decimal.RequireFromString("-5"), Note: "  "}
	if fields(t, ValidateAdjustment(&adj))["note"] == "" {
		t.Error("adjustment without a note accepted")
	}
	adj.Note = "returned item"
	if err := ValidateAdjustment(&adj); err != nil {
		t.Errorf("negative adjustment with a note is valid: %v", err)
	}
}
