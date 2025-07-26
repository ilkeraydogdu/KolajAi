package validation

// FormSchema defines validation rules for a form
type FormSchema struct {
	Fields map[string]FieldRules
}

// FieldRules defines validation rules for a field
type FieldRules struct {
	Required  bool
	MinLength int
	MaxLength int
	Pattern   string
	Message   string
}

// RegisterSchemas creates and returns all form validation schemas
func RegisterSchemas() map[string]FormSchema {
	schemas := make(map[string]FormSchema)

	// Register formu için şema
	schemas["register"] = FormSchema{
		Fields: map[string]FieldRules{
			"name": {
				Required:  true,
				MinLength: 5,
				Message:   "Ad Soyad en az 5 karakter olmalıdır",
			},
			"email": {
				Required: true,
				Pattern:  `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`,
				Message:  "Geçerli bir e-posta adresi giriniz",
			},
			"phone": {
				Required: true,
				Pattern:  `^0[0-9 ]{10,14}$`,
				Message:  "Telefon numarası 0 ile başlamalıdır",
			},
			"captcha": {
				Required: true,
				Message:  "Güvenlik sorusunun cevabını giriniz",
			},
			"terms": {
				Required: true,
				Message:  "Kullanım koşullarını kabul etmelisiniz",
			},
		},
	}

	// Login formu için şema
	schemas["login"] = FormSchema{
		Fields: map[string]FieldRules{
			"email": {
				Required: true,
				Pattern:  `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`,
				Message:  "Geçerli bir e-posta adresi giriniz",
			},
			"password": {
				Required: true,
				Message:  "Şifre alanı zorunludur",
			},
		},
	}

	// Şifre sıfırlama formu için şema
	schemas["forgotPassword"] = FormSchema{
		Fields: map[string]FieldRules{
			"email": {
				Required: true,
				Pattern:  `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`,
				Message:  "Geçerli bir e-posta adresi giriniz",
			},
		},
	}

	// Şifre değiştirme formu için şema
	schemas["resetPassword"] = FormSchema{
		Fields: map[string]FieldRules{
			"password": {
				Required:  true,
				MinLength: 8,
				Message:   "Şifre en az 8 karakter olmalıdır",
			},
			"password_confirm": {
				Required: true,
				Message:  "Şifre tekrarı zorunludur",
			},
		},
	}

	return schemas
}
