package crypto

import (
	"github.com/uwemakan/signing-service/utils"
)

// Signer defines a contract for different types of signing implementations.
type Signer interface {
	Sign(dataToBeSigned []byte) ([]byte, error)
}

type SignerFactory struct{
	eccMarshaler *ECCMarshaler
	rsaMarshaler *RSAMarshaler
}

// NewSignerFactory returns a new SignerFactory.
func NewSignerFactory() *SignerFactory {
	return &SignerFactory{
		eccMarshaler: &ECCMarshaler{},
		rsaMarshaler: &RSAMarshaler{},
	}
}

// GetSigner returns a Signer implementation for a given algorithm and private key.
func (f *SignerFactory) GetSigner(algorithm string, privateKey []byte) (Signer, error) {
    switch algorithm {
    case "RSA":
		return f.rsaMarshaler.Unmarshal(privateKey)
    case "ECC":
        return f.eccMarshaler.Decode(privateKey)
    default:
        return nil, utils.ErrUnsupportedAlgorithm
    }
}
