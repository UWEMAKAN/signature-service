package api

import (
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/uwemakan/signing-service/domain"
	"github.com/uwemakan/signing-service/utils"
)

func validateUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}

func validateData(data string) bool {
	if data == "" {
		return false
	}
	ds := strings.Split(data, "_")
	if len(ds) != 3 {
		return false
	}
	if strings.Trim(ds[0], " ") == "" || strings.Trim(ds[1], " ") == "" || strings.Trim(ds[2], " ") == "" {
		return false
	}
	return true
}

func validateAlgorithm(algorithm string) bool {
	if algorithm == "" {
		return false
	}
	if !slices.Contains(utils.Algorithms, algorithm) {
		return false
	}
	return true
}

func validateLabel(label string) bool {
	return label != ""
}

func validateSignatureDeviceRequest(request *domain.SignatureDeviceRequest) (errs []string) {
	if !validateUUID(request.ID) {
		errs = append(errs, fmt.Sprintf("invalid device id: %s is not a valid UUID", request.ID))
	}
	if !validateAlgorithm(request.Algorithm) {
		errs = append(errs, fmt.Sprintf("algorithm must be one of %s", utils.Algorithms))
	}
	if request.Label != nil {
		if !validateLabel(*request.Label) {
			errs = append(errs, "invalid label")
		}
	}
	return
}

func validateTransactionSignatureRequest(request *domain.SignTransactionRequest) (errs []string) {
	if !validateUUID(request.ID) {
		errs = append(errs, fmt.Sprintf("invalid device id: %s is not a valid UUID", request.ID))
	}
	if !validateData(request.Data) {
		errs = append(errs, "invalid data: data must be in the format signatureCounter_data_lastSignature")
	}
	return
}
