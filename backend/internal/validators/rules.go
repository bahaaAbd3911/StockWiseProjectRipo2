package validators

import (
	"ecommerce-backend/internal/models"
)

// Limits shared with the services.
const (
	MinPasswordLength = 8
	MaxPasswordBytes  = 72 // bcrypt ignores everything after 72 bytes, so we reject longer passwords
	MaxQuantity       = 1000
	MaxReviewLength   = 5000
	MaxStock          = 1_000_000
)

// ---------- users ----------

func ValidateRegister(req *models.RegisterRequest) error {
	Trim(&req.FirstName)
	Trim(&req.LastName)
	Trim(&req.Phone)
	req.Email = NormalizeEmail(req.Email)
	TrimOptional(&req.ProfileImageURL)

	c := New()
	checkName(c, req.FirstName, req.LastName)
	c.Required("email", req.Email)
	c.Add(req.Email == "" || IsEmail(req.Email), "email", "must be a valid email address")
	c.MaxLen("email", req.Email, 255)
	checkPassword(c, "password", req.Password)
	checkPhone(c, "phone", req.Phone)
	checkOptionalURL(c, "profile_image_url", req.ProfileImageURL)
	return c.Err()
}

func ValidateLogin(req *models.LoginRequest) error {
	req.Email = NormalizeEmail(req.Email)
	c := New()
	c.Required("email", req.Email)
	c.Add(req.Password != "", "password", "is required")
	return c.Err()
}

func ValidateUpdateProfile(req *models.UpdateProfileRequest) error {
	Trim(&req.FirstName)
	Trim(&req.LastName)
	Trim(&req.Phone)
	TrimOptional(&req.ProfileImageURL)

	c := New()
	checkName(c, req.FirstName, req.LastName)
	checkPhone(c, "phone", req.Phone)
	checkOptionalURL(c, "profile_image_url", req.ProfileImageURL)
	return c.Err()
}

func ValidateChangePassword(req *models.ChangePasswordRequest) error {
	c := New()
	c.Add(req.CurrentPassword != "", "current_password", "is required")
	checkPassword(c, "new_password", req.NewPassword)
	c.Add(req.NewPassword == "" || req.NewPassword != req.CurrentPassword, "new_password", "must be different from the current password")
	return c.Err()
}

func checkName(c *Checker, first, last string) {
	c.Required("first_name", first)
	c.MaxLen("first_name", first, 100)
	c.Required("last_name", last)
	c.MaxLen("last_name", last, 100)
}

func checkPassword(c *Checker, field, pw string) {
	c.Add(len(pw) >= MinPasswordLength, field, "must be at least 8 characters")
	c.Add(len(pw) <= MaxPasswordBytes, field, "must be at most 72 bytes")
}

func checkPhone(c *Checker, field, phone string) {
	c.Required(field, phone)
	c.Add(phone == "" || IsPhone(phone), field, "must be a valid phone number")
	c.MaxLen(field, phone, 20)
}

func checkOptionalURL(c *Checker, field string, u *string) {
	if u != nil {
		c.Add(IsHTTPURL(*u), field, "must be an http(s) URL")
	}
}

// ---------- products & catalog ----------

func ValidateProduct(in *models.ProductInput) error {
	Trim(&in.Name)
	Trim(&in.NameAr)
	Trim(&in.Description)
	Trim(&in.DescriptionAr)
	Trim(&in.Category)
	Trim(&in.CategoryAr)

	c := New()
	c.Required("name", in.Name)
	c.MaxLen("name", in.Name, 255)
	c.Required("name_ar", in.NameAr)
	c.MaxLen("name_ar", in.NameAr, 255)
	c.MaxLen("category", in.Category, 100)
	c.MaxLen("category_ar", in.CategoryAr, 100)
	c.Money("price", in.Price)
	c.Percent("discount_value", in.DiscountValue)
	c.Add(in.Stock >= 0, "stock", "cannot be negative")
	c.Add(in.Stock <= MaxStock, "stock", "is too large")
	return c.Err()
}

func ValidateProductImage(in *models.ProductImageInput) error {
	Trim(&in.ImageURL)
	c := New()
	c.Required("image_url", in.ImageURL)
	c.Add(in.ImageURL == "" || IsHTTPURL(in.ImageURL), "image_url", "must be an http(s) URL")
	c.Add(in.DisplayOrder >= 0, "display_order", "cannot be negative")
	return c.Err()
}

func ValidateProductColor(in *models.ProductColorInput) error {
	Trim(&in.Name)
	Trim(&in.NameAr)
	Trim(&in.HexCode)
	c := New()
	c.Required("name", in.Name)
	c.MaxLen("name", in.Name, 50)
	c.Required("name_ar", in.NameAr)
	c.MaxLen("name_ar", in.NameAr, 50)
	c.Add(IsHexColor(in.HexCode), "hex_code", "must look like #RRGGBB")
	return c.Err()
}

// ---------- cart / waiting list ----------

func ValidateQuantity(c *Checker, q int) {
	c.Add(q >= 1 && q <= MaxQuantity, "quantity", "must be between 1 and 1000")
}

func ValidateAddCartItem(req *models.AddCartItemRequest) error {
	TrimOptional(&req.SelectedColor)
	c := New()
	c.Add(req.ProductID > 0, "product_id", "is required")
	ValidateQuantity(c, req.Quantity)
	c.OptionalMaxLen("selected_color", req.SelectedColor, 50)
	return c.Err()
}

func ValidateUpdateCartItem(req *models.UpdateCartItemRequest) error {
	TrimOptional(&req.SelectedColor)
	c := New()
	ValidateQuantity(c, req.Quantity)
	c.OptionalMaxLen("selected_color", req.SelectedColor, 50)
	return c.Err()
}

func ValidateWaitingList(req *models.CreateWaitingListRequest) error {
	TrimOptional(&req.SelectedColor)
	c := New()
	c.Add(req.ProductID > 0, "product_id", "is required")
	ValidateQuantity(c, req.Quantity)
	c.OptionalMaxLen("selected_color", req.SelectedColor, 50)
	return c.Err()
}

// ---------- addresses ----------

func ValidateAddress(in *models.AddressInput) error {
	TrimOptional(&in.AddressLabel)
	Trim(&in.RecipientName)
	Trim(&in.Phone)
	Trim(&in.City)
	Trim(&in.Area)
	Trim(&in.Street)
	TrimOptional(&in.BuildingNumber)
	TrimOptional(&in.ApartmentNumber)
	TrimOptional(&in.AdditionalInstructions)

	c := New()
	c.OptionalMaxLen("address_label", in.AddressLabel, 50)
	c.Required("recipient_name", in.RecipientName)
	c.MaxLen("recipient_name", in.RecipientName, 150)
	checkPhone(c, "phone", in.Phone)
	c.Required("city", in.City)
	c.MaxLen("city", in.City, 100)
	c.Required("area", in.Area)
	c.MaxLen("area", in.Area, 100)
	c.Required("street", in.Street)
	c.MaxLen("street", in.Street, 255)
	c.OptionalMaxLen("building_number", in.BuildingNumber, 50)
	c.OptionalMaxLen("apartment_number", in.ApartmentNumber, 50)
	c.OptionalMaxLen("additional_instructions", in.AdditionalInstructions, 1000)
	return c.Err()
}

// ---------- reviews ----------

func ValidateReview(in *models.ReviewInput) error {
	Trim(&in.Comment)
	TrimOptional(&in.CommentAr)
	c := New()
	c.Required("comment", in.Comment)
	c.MaxLen("comment", in.Comment, MaxReviewLength)
	c.OptionalMaxLen("comment_ar", in.CommentAr, MaxReviewLength)
	c.Add(in.Rating >= 1 && in.Rating <= 5, "rating", "must be between 1 and 5")
	return c.Err()
}

func ValidateForgotPassword(req *models.ForgotPasswordRequest) error {
	req.Email = NormalizeEmail(req.Email)
	c := New()
	c.Required("email", req.Email)
	c.Add(req.Email == "" || IsEmail(req.Email), "email", "must be a valid email address")
	c.MaxLen("email", req.Email, 255)
	return c.Err()
}

func ValidateResetPassword(req *models.ResetPasswordRequest) error {
	Trim(&req.Token)
	c := New()
	c.Add(req.Token != "", "token", "is required")
	c.Add(len(req.Token) <= 200, "token", "is invalid")
	checkPassword(c, "new_password", req.NewPassword)
	return c.Err()
}
