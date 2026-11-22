package validators

import (
	"strings"
	"testing"

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
