package models

import "errors"

var (
	ErrNoRecord = errors.New("models: no matching record foud")

	ErrInvalidCredentials = errors.New("models: invalid credentials")

	ErrDuplicateEmail = errors.New("models: duplicate email")
)
