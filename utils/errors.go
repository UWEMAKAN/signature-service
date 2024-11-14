package utils

import "errors"

var (
    ErrUnsupportedAlgorithm = errors.New("unsupported algorithm")
	ErrInvalidSignatureCounter = errors.New("invalid signature counter")
	ErrInvalidLastSignature    = errors.New("invalid last signature")
	ErrInvalidData             = errors.New("invalid data")
	ErrDeviceNotFound      = errors.New("device not found")
	ErrDeviceAlreadyExists = errors.New("device already exists")
	ErrInvalidDeviceId = errors.New("device ID must be a valid UUID")
)
