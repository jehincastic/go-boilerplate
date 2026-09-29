package validator

import "regexp"

// EmailRX matches a practical subset of email addresses.
var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// Validator collects field errors for a single request.
type Validator struct {
	Errors map[string]string
}

// Error is returned when one or more fields are invalid.
type Error struct {
	Errors map[string]string
}

func (e *Error) Error() string {
	return "validation failed"
}

// New returns an empty validator.
func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

// Valid reports whether the validator has collected no errors.
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// Err returns nil when the validator is valid.
func (v *Validator) Err() error {
	if v.Valid() {
		return nil
	}
	return &Error{Errors: v.Errors}
}

// AddError records the first message for key.
func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

// Check records message when ok is false.
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

// PermittedValue reports whether value is in permittedValues.
func PermittedValue[T comparable](value T, permittedValues ...T) bool {
	for i := range permittedValues {
		if value == permittedValues[i] {
			return true
		}
	}
	return false
}

// Matches reports whether value matches rx.
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

// Unique reports whether values contains no duplicates.
func Unique[T comparable](values []T) bool {
	seen := make(map[T]struct{}, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			return false
		}
		seen[value] = struct{}{}
	}
	return true
}

// ValidateName checks that a display name is present and within a bounded length.
func ValidateName(v *Validator, name string) {
	v.Check(name != "", "name", "must be provided")
	v.Check(len(name) <= 100, "name", "must not be more than 100 bytes long")
}

// ValidateEmail checks that email is present and well formed.
func ValidateEmail(v *Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(Matches(email, EmailRX), "email", "must be a valid email address")
}

// ValidatePasswordPlaintext checks that a password is present and within a bounded length.
func ValidatePasswordPlaintext(v *Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 128, "password", "must not be more than 128 bytes long")
}
