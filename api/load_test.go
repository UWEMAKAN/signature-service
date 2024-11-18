package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/uwemakan/signing-service/domain"
	"github.com/uwemakan/signing-service/utils"
)
var config = &utils.Config{ServerAddress: ":0", AESKey: []byte("1234567890123456")}

func TestLoadCreateSignatureDevice(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TestLoadCreateSignatureDevice in short mode.")
	}
	var wg sync.WaitGroup
	server := NewServer(config)
	// You can vary the number of devices
	numberOfDevices := 100
	for i := range numberOfDevices {
		wg.Add(1)
		go func(t *testing.T, n int) {
			defer wg.Done()
			requires := require.New(t)
			id, _ := uuid.NewRandom()
			payload := &domain.SignatureDeviceRequest{
				ID:        id.String(),
				Algorithm: utils.Algorithms[n%2],
			}
			recorder := httptest.NewRecorder()

			b, err := json.Marshal(payload)
			requires.NoError(err)

			url := "/api/v0/signature-devices"
			body := bytes.NewReader(b)

			request, err := http.NewRequest(http.MethodPost, url, body)
			requires.NoError(err)

			server.Handler(recorder, request)
			requires.Equal(http.StatusCreated, recorder.Code)
		}(t, i)
	}
	wg.Wait()
	requires := require.New(t)

	url := "/api/v0/signature-devices"
	request, err := http.NewRequest(http.MethodGet, url, nil)
	requires.NoError(err)

	recorder := httptest.NewRecorder()
	server.Handler(recorder, request)
	requires.Equal(http.StatusOK, recorder.Code)

	body, err := io.ReadAll(recorder.Body)
	requires.NoError(err)
	requires.NotNil(body)
	var response Response
	err = json.Unmarshal(body, &response)
	requires.NoError(err)
	b, err := json.Marshal(response.Data)
	requires.NoError(err)
	var devices []domain.SignatureDevice
	err = json.Unmarshal(b, &devices)
	requires.NoError(err)
	requires.Len(devices, numberOfDevices)
}

func createSignatureDevice(t *testing.T, server *Server, n int, dch chan time.Duration) *domain.SignatureDevice {
	requires := require.New(t)
	id, _ := uuid.NewRandom()
	payload := &domain.SignatureDeviceRequest{
		ID:        id.String(),
		Algorithm: utils.Algorithms[n%2],
	}
	recorder := httptest.NewRecorder()

	b, err := json.Marshal(payload)
	requires.NoError(err)

	url := "/api/v0/signature-devices"
	body := bytes.NewReader(b)

	request, err := http.NewRequest(http.MethodPost, url, body)
	requires.NoError(err)
	start := time.Now()
	server.Handler(recorder, request)
	end := time.Now()
	dch <- end.Sub(start)
	requires.Equal(http.StatusCreated, recorder.Code)
	br, err := io.ReadAll(recorder.Body)
	requires.NoError(err)
	requires.NotNil(br)
	var response Response
	err = json.Unmarshal(br, &response)
	requires.NoError(err)
	b, err = json.Marshal(response.Data)
	requires.NoError(err)
	var device domain.SignatureDevice
	err = json.Unmarshal(b, &device)
	requires.NoError(err)
	requires.Equal(id.String(), device.ID)
	requires.Equal(utils.Algorithms[n%2], device.Algorithm)
	requires.Equal(base64.StdEncoding.EncodeToString([]byte(id.String())), device.LastSignature)
	requires.Equal(0, device.SignatureCounter)
	return &device
}

func TestLoadSignTransaction(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TestLoadSignTransaction in short mode.")
	}

	var wg sync.WaitGroup
	server := NewServer(config)
	// You can vary the number of devices and the number of signings
	numberOfDevices := 100
	numberOfSignings := 1000
	dch := make(chan time.Duration, numberOfDevices)
	sch := make(chan time.Duration, numberOfDevices*numberOfSignings)

	for i := 0; i < numberOfDevices; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			requires := require.New(t)
			device := createSignatureDevice(t, server, n, dch)
			requires.NotNil(device)

			for j := 0; j < numberOfSignings; j++ {
				data := fmt.Sprintf("%d_TESTDATA_%s", device.SignatureCounter, device.LastSignature)
				pd := &domain.SignTransactionRequest{
					ID:   device.ID,
					Data: data,
				}
				b, err := json.Marshal(pd)
				requires.NoError(err)

				url := "/api/v0/signature-devices/sign"
				body := bytes.NewReader(b)

				req, err := http.NewRequest(http.MethodPost, url, body)
				requires.NoError(err)
				recorder := httptest.NewRecorder()
				start := time.Now()
				server.SignTransaction(recorder, req)
				end := time.Now()
				sch <- end.Sub(start)
				requires.Equal(http.StatusOK, recorder.Code)
				bb, err := io.ReadAll(recorder.Body)
				requires.NoError(err)
				requires.NotNil(bb)

				var response Response
				err = json.Unmarshal(bb, &response)
				requires.NoError(err)

				b, err = json.Marshal(response.Data)
				requires.NoError(err)

				var signatureData domain.SignTransactionResponse
				err = json.Unmarshal(b, &signatureData)
				requires.NoError(err)
				requires.NotZero(signatureData.Signature)
				requires.Equal(data, signatureData.SignedData)

				device.SignatureCounter++
				device.LastSignature = signatureData.Signature
			}
		}(i)
	}

	wg.Wait()
	close(dch)
	close(sch)
	prefix := utils.RandomString(6)
	wg.Add(1)
	go generateReport(dch, numberOfDevices*numberOfSignings, fmt.Sprintf("../reports/%s-devices-%d.txt", prefix, numberOfDevices), &wg)
	wg.Add(1)
	go generateReport(sch, numberOfDevices*numberOfSignings, fmt.Sprintf("../reports/%s-signatures-%d.txt", prefix, numberOfDevices*numberOfSignings), &wg)
	wg.Wait()
	requires := require.New(t)
	recorder := httptest.NewRecorder()

	url := "/api/v0/signature-devices"

	request, err := http.NewRequest(http.MethodGet, url, nil)
	requires.NoError(err)

	server.Handler(recorder, request)
	requires.Equal(http.StatusOK, recorder.Code)

	body, err := io.ReadAll(recorder.Body)
	requires.NoError(err)
	requires.NotNil(body)
	var response Response
	err = json.Unmarshal(body, &response)
	requires.NoError(err)
	b, err := json.Marshal(response.Data)
	requires.NoError(err)
	var devices []domain.SignatureDevice
	err = json.Unmarshal(b, &devices)
	requires.NoError(err)
	requires.Len(devices, numberOfDevices)
	for _, device := range devices {
		requires.Equal(numberOfSignings, device.SignatureCounter)
	}
}

func generateReport(ch chan time.Duration, n int, filename string, wg *sync.WaitGroup) {
	defer wg.Done()
	ds := make([]time.Duration, 0, n)
	min := time.Hour
	max := time.Duration(0)
	sum := time.Duration(0)
	for d := range ch {
		ds = append(ds, d)
		if d < min {
			min = d
		}
		if d > max {
			max = d
		}
		sum += d
	}
	sort.Slice(ds, func(i, j int) bool {
		return ds[i] < ds[j]
	})
	p100 := ds[int(len(ds)-1)]
	p99 := ds[int(float64(len(ds))*0.99)]
	p90 := ds[int(float64(len(ds))*0.90)]
	p75 := ds[int(float64(len(ds))*0.75)]
	p50 := ds[int(float64(len(ds))*0.50)]
	p25 := ds[int(float64(len(ds))*0.25)]
	p10 := ds[int(float64(len(ds))*0.10)]
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Min:\t%dms\n", min.Milliseconds()))
	sb.WriteString(fmt.Sprintf("Max:\t%dms\n", max.Milliseconds()))
	sb.WriteString(fmt.Sprintf("Avg:\t%fms\n", float64(sum.Milliseconds())/float64(n)))
	sb.WriteString(fmt.Sprintf("10%%:\t%dms\n", p10.Milliseconds()))
	sb.WriteString(fmt.Sprintf("25%%:\t%dms\n", p25.Milliseconds()))
	sb.WriteString(fmt.Sprintf("50%%:\t%dms\n", p50.Milliseconds()))
	sb.WriteString(fmt.Sprintf("75%%:\t%dms\n", p75.Milliseconds()))
	sb.WriteString(fmt.Sprintf("90%%:\t%dms\n", p90.Milliseconds()))
	sb.WriteString(fmt.Sprintf("99%%:\t%dms\n", p99.Milliseconds()))
	sb.WriteString(fmt.Sprintf("100%%:\t%dms\n", p100.Milliseconds()))

	err := os.WriteFile(filename, []byte(sb.String()), 0644)
	if err != nil {
		panic(err)
	}
}
