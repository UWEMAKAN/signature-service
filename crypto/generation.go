package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"

	"github.com/uwemakan/signing-service/utils"
)

type KeyPairFactory struct {
	rsaGenerator  *RSAGenerator
	eccGenerator  *ECCGenerator
}

type KeyPair struct {
	Public  []byte
	Private []byte
}

func NewKeyPairFactory() *KeyPairFactory {
	return &KeyPairFactory{
		rsaGenerator:  &RSAGenerator{
			rsaMarshaler:  &RSAMarshaler{},
		},
		eccGenerator:  &ECCGenerator{
			eccMarshaler:  &ECCMarshaler{},
		},
	}
}

func (f *KeyPairFactory) GenerateKeyPair(algorithm string) ([]byte, []byte, error) {
	switch algorithm {
	case "RSA":
		return f.rsaGenerator.GenerateMarshaled()
	case "ECC":
		return f.eccGenerator.GenerateMarshaled()
	default:
		return nil, nil, utils.ErrUnsupportedAlgorithm
	}
}

// RSAGenerator generates a RSA key pair.
type RSAGenerator struct{
	rsaMarshaler  *RSAMarshaler
}

// Generate generates a new RSAKeyPair.
func (g *RSAGenerator) Generate() (*RSAKeyPair, error) {
	// Security has been ignored for the sake of simplicity.
	// I increased the key size from 512 to 2048 because the key size of 512 was too small to sign the messages.
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	return &RSAKeyPair{
		Public:  &key.PublicKey,
		Private: key,
	}, nil
}

// Generates and Returns a new RSAKeyPair Marshaled.
func (g *RSAGenerator) GenerateMarshaled() ([]byte, []byte, error) {
	keyPair, err := g.Generate()
	if err != nil {
		return nil, nil, err
	}
	return g.rsaMarshaler.Marshal(*keyPair)
}

// ECCGenerator generates an ECC key pair.
type ECCGenerator struct{
	eccMarshaler  *ECCMarshaler
}

// Generate generates a new ECCKeyPair.
func (g *ECCGenerator) Generate() (*ECCKeyPair, error) {
	// Security has been ignored for the sake of simplicity.
	key, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return nil, err
	}

	return &ECCKeyPair{
		Public:  &key.PublicKey,
		Private: key,
	}, nil
}

// Generates and Returns a new ECCKeyPair Marshaled.
func (g *ECCGenerator) GenerateMarshaled() ([]byte, []byte, error) {
	keyPair, err := g.Generate()
	if err != nil {
		return nil, nil, err
	}
	return g.eccMarshaler.Encode(*keyPair)
}
