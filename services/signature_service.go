package services

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/uwemakan/signing-service/crypto"
	"github.com/uwemakan/signing-service/domain"
	"github.com/uwemakan/signing-service/persistence"
	"github.com/uwemakan/signing-service/utils"
)

var aesKey = []byte("1234567890123456")

type SignatureService interface {
	ListSignatureDevices() ([]*domain.SignatureDevice, error)
	GetSignatureDevice(deviceId string) (*domain.SignatureDevice, error)
	CreateSignatureDevice(request *domain.SignatureDeviceRequest) (*domain.SignatureDevice, error)
	SignTransaction(deviceId, data string) (*domain.SignTransactionResponse, error)
}

type signatureService struct {
	repo           persistence.SignatureDeviceRepository
	keyPairFactory *crypto.KeyPairFactory
	signerFactory  *crypto.SignerFactory
}

type SignatureServiceParams struct {
	Repo           persistence.SignatureDeviceRepository
	KeyPairFactory *crypto.KeyPairFactory
	SignerFactory  *crypto.SignerFactory
}

func NewSignatureService(params SignatureServiceParams) SignatureService {
	return &signatureService{repo: params.Repo, keyPairFactory: params.KeyPairFactory, signerFactory: params.SignerFactory}
}

func (s *signatureService) GetSignatureDevice(deviceId string) (*domain.SignatureDevice, error) {
	return s.repo.GetDevice(deviceId)
}

func (s *signatureService) CreateSignatureDevice(request *domain.SignatureDeviceRequest) (*domain.SignatureDevice, error) {
	publicKey, privateKey, err := s.keyPairFactory.GenerateKeyPair(request.Algorithm)
	if err != nil {
		return nil, err
	}
	encryptedPrivateKey, err := crypto.EncryptAES(privateKey, aesKey)
	if err != nil {
		return nil, err
	}
	label := ""
	if request.Label != nil {
		label = *request.Label
	}
	return s.repo.CreateDevice(request.ID, request.Algorithm, string(publicKey), encryptedPrivateKey, label)
}

func (s *signatureService) SignTransaction(deviceId, data string) (*domain.SignTransactionResponse, error) {
	dataSlice := strings.Split(data, "_")
	device, err := s.repo.GetDevice(deviceId)
	if err != nil {
		return nil, err
	}
	if fmt.Sprint(device.SignatureCounter) != dataSlice[0] {
		return nil, utils.ErrInvalidSignatureCounter
	}
	if device.LastSignature != dataSlice[2] {
		return nil, utils.ErrInvalidLastSignature
	}
	decryptedPrivateKey, err := crypto.DecryptAES(device.PrivateKey, aesKey)
	if err != nil {
		return nil, err
	}
	signer, err := s.signerFactory.GetSigner(device.Algorithm, []byte(decryptedPrivateKey))
	if err != nil {
		return nil, err
	}
	signature, err := signer.Sign([]byte(dataSlice[1]))
	if err != nil {
		return nil, err
	}
	encodedSignature := base64.StdEncoding.EncodeToString(signature)
	s.repo.UpdateDevice(deviceId, encodedSignature)
	return &domain.SignTransactionResponse{Signature: encodedSignature, SignedData: data}, nil
}

func (s *signatureService) ListSignatureDevices() ([]*domain.SignatureDevice, error) {
	return s.repo.ListDevices()
}
