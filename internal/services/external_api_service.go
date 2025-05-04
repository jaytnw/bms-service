package services

import (
	"errors"

	"github.com/go-resty/resty/v2"
)

type WashingMachine struct {
	IDWashingMachine string `json:"idWashing_Machine"`
	IDDorm           string `json:"idDorm"`
	DormName         string `json:"dormName"`
}

type ExternalApiResponse struct {
	Status string           `json:"status"`
	Data   []WashingMachine `json:"data"`
}

type ExternalAPIService interface {
	FetchWashingMachines() ([]WashingMachine, error)
}

type externalAPIService struct {
	client *resty.Client
	apiURL string
}

func NewExternalAPIService(url string) ExternalAPIService {
	client := resty.New()
	return &externalAPIService{
		client: client,
		apiURL: url,
	}
}

func (s *externalAPIService) FetchWashingMachines() ([]WashingMachine, error) {
	var result ExternalApiResponse

	resp, err := s.client.R().
		SetResult(&result).
		Get(s.apiURL + "/getAlldorm")

	if err != nil {
		return nil, err
	}

	// Log raw response body
	// log.Printf("📦 Raw Response Body: %s", resp.Body())

	// log.Printf("status Code %v", resp.StatusCode())
	// log.Printf("result.Status %v", result.Status)

	if resp.StatusCode() != 200 || result.Status != "1" {
		return nil, errors.New("API response not successful")
	}

	// Log parsed data
	// log.Printf("✅ Parsed Data: %+v", result.Data)

	return result.Data, nil
}
