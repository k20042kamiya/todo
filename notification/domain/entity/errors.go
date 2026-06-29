package entity

import "errors"

var (
	ErrNotFound         = errors.New("not found")
	ErrInvalidRecipient = errors.New("invalid recipient")
)
