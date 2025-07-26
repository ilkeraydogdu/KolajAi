package validation

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// RequiredRule checks if a value is not empty
func RequiredRule(value string, options ...interface{}) (bool, string) {
	if strings.TrimSpace(value) == "" {
		return false, "Bu alan zorunludur"
	}
	return true, ""
}

// EmailRule validates email format
func EmailRule(value string, options ...interface{}) (bool, string) {
	if value == "" {
		return true, "" // Empty is valid, use required rule to make mandatory
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(value) {
		return false, "Geçerli bir e-posta adresi giriniz"
	}
	return true, ""
}

// MinLengthRule validates minimum string length
func MinLengthRule(value string, options ...interface{}) (bool, string) {
	if value == "" {
		return true, "" // Empty is valid, use required rule to make mandatory
	}

	if len(options) == 0 {
		return false, "Min uzunluk değeri belirtilmemiş"
	}

	minLength, err := strconv.Atoi(options[0].(string))
	if err != nil {
		return false, "Geçersiz min uzunluk değeri"
	}

	if len(value) < minLength {
		return false, "En az " + strconv.Itoa(minLength) + " karakter gereklidir"
	}
	return true, ""
}

// MaxLengthRule validates maximum string length
func MaxLengthRule(value string, options ...interface{}) (bool, string) {
	if value == "" {
		return true, "" // Empty is valid, use required rule to make mandatory
	}

	if len(options) == 0 {
		return false, "Max uzunluk değeri belirtilmemiş"
	}

	maxLength, err := strconv.Atoi(options[0].(string))
	if err != nil {
		return false, "Geçersiz max uzunluk değeri"
	}

	if len(value) > maxLength {
		return false, "En fazla " + strconv.Itoa(maxLength) + " karakter olmalıdır"
	}
	return true, ""
}

// NumericRule validates if a string contains only numbers
func NumericRule(value string, options ...interface{}) (bool, string) {
	if value == "" {
		return true, "" // Empty is valid, use required rule to make mandatory
	}

	numericRegex := regexp.MustCompile(`^[0-9]+$`)
	if !numericRegex.MatchString(value) {
		return false, "Sadece rakam giriniz"
	}
	return true, ""
}

// AlphaRule validates if a string contains only letters
func AlphaRule(value string, options ...interface{}) (bool, string) {
	if value == "" {
		return true, "" // Empty is valid, use required rule to make mandatory
	}

	alphaRegex := regexp.MustCompile(`^[a-zA-ZğüşıöçĞÜŞİÖÇ ]+$`)
	if !alphaRegex.MatchString(value) {
		return false, "Sadece harf giriniz"
	}
	return true, ""
}

// AlphaNumericRule validates if a string contains only letters and numbers
func AlphaNumericRule(value string, options ...interface{}) (bool, string) {
	if value == "" {
		return true, "" // Empty is valid, use required rule to make mandatory
	}

	alphaNumericRegex := regexp.MustCompile(`^[a-zA-Z0-9ğüşıöçĞÜŞİÖÇ ]+$`)
	if !alphaNumericRegex.MatchString(value) {
		return false, "Sadece harf ve rakam giriniz"
	}
	return true, ""
}

// PhoneRule validates Turkish phone number format
func PhoneRule(value string, options ...interface{}) (bool, string) {
	if value == "" {
		return true, "" // Empty is valid, use required rule to make mandatory
	}

	// Türk telefon numarası formatı: 0 ile başlayan herhangi bir format
	phoneRegex := regexp.MustCompile(`^0[0-9 ]{10,14}$`)
	if !phoneRegex.MatchString(value) {
		return false, "Telefon numarası 0 ile başlamalıdır"
	}
	return true, ""
}

// URLRule validates if a string is a valid URL
func URLRule(value string, options ...interface{}) (bool, string) {
	if value == "" {
		return true, "" // Empty is valid, use required rule to make mandatory
	}

	urlRegex := regexp.MustCompile(`^(http|https):\/\/[a-zA-Z0-9]+([\-\.]{1}[a-zA-Z0-9]+)*\.[a-zA-Z]{2,}(:[0-9]{1,5})?(\/.*)?$`)
	if !urlRegex.MatchString(value) {
		return false, "Geçerli bir URL giriniz"
	}
	return true, ""
}

// DateRule validates if a string is a valid date
func DateRule(value string, options ...interface{}) (bool, string) {
	if value == "" {
		return true, "" // Empty is valid, use required rule to make mandatory
	}

	format := "2006-01-02" // Default format
	if len(options) > 0 {
		format = options[0].(string)
	}

	_, err := time.Parse(format, value)
	if err != nil {
		return false, "Geçerli bir tarih giriniz"
	}
	return true, ""
}

// MatchRule validates if a string matches another field
func MatchRule(value string, options ...interface{}) (bool, string) {
	if len(options) == 0 {
		return false, "Eşleşme alanı belirtilmemiş"
	}

	// options[0] should be the field name, options[1] should be the value to match
	if len(options) < 2 {
		return false, "Eşleşme değeri sağlanmamış"
	}

	fieldName := options[0].(string)
	matchValue := options[1].(string)

	if value != matchValue {
		return false, fieldName + " alanı ile eşleşmiyor"
	}
	return true, ""
}
