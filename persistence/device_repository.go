package persistence

import (
	"github.com/uwemakan/signing-service/domain"
)

type SignatureDeviceRepository interface {
	CreateDevice(id, algorithm, publicKey, privateKey, label string) (*domain.SignatureDevice, error)
	GetDevice(id string) (*domain.SignatureDevice, error)
	ListDevices() ([]*domain.SignatureDevice, error)
	UpdateDevice(deviceID, newSignature string) error
}
