package errs

import "errors"

var (
	// ErrNotFound is returned when a query matches no rows.
	ErrNotFound = errors.New("record not found")
	// ErrEditConflict is returned when an update loses an optimistic-lock check.
	ErrEditConflict = errors.New("edit conflict")
	// ErrDuplicateEmail is returned when a user email is already stored.
	ErrDuplicateEmail = errors.New("duplicate email")
	// ErrInvalidCredentials is returned when an email and password do not match a user.
	ErrInvalidCredentials = errors.New("invalid credentials")
)
