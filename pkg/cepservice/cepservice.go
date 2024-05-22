package cepservice

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-resty/resty/v2"
)

type Address struct {
	CEP        string `json:"cep"`
	Logradouro string `json:"logradouro"`
	Bairro     string `json:"bairro"`
	Localidade string `json:"localidade"`
	UF         string `json:"uf"`
}

type BrasilAPIResponse struct {
	CEP          string `json:"cep"`
	State        string `json:"state"`
	City         string `json:"city"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
}

type Result struct {
	API     string
	Address Address
}

func ConvertBrasilAPIResponseToAddress(response BrasilAPIResponse) Address {
	return Address{
		CEP:        response.CEP,
		Logradouro: response.Street,
		Bairro:     response.Neighborhood,
		Localidade: response.City,
		UF:         response.State,
	}
}

func fetchFromAPI(ctx context.Context, url string, ch chan<- Result, apiName string) {
	client := resty.New()

	if apiName == "BrasilAPI" {
		var brasilResponse BrasilAPIResponse
		resp, err := client.R().
			SetContext(ctx).
			SetResult(&brasilResponse).
			Get(url)

		if err != nil || resp.StatusCode() != 200 {
			return
		}

		address := ConvertBrasilAPIResponseToAddress(brasilResponse)
		ch <- Result{API: apiName, Address: address}
	} else if apiName == "ViaCEP" {
		var viaCEPResponse Address
		resp, err := client.R().
			SetContext(ctx).
			SetResult(&viaCEPResponse).
			Get(url)

		if err != nil || resp.StatusCode() != 200 {
			return
		}

		ch <- Result{API: apiName, Address: viaCEPResponse}
	}
}

func GetFasterAPIResult(ctx context.Context, cep string) (Result, error) {
	ch := make(chan Result, 2)
	urls := map[string]string{
		"BrasilAPI": fmt.Sprintf("https://brasilapi.com.br/api/cep/v1/%s", cep),
		"ViaCEP":    fmt.Sprintf("http://viacep.com.br/ws/%s/json/", cep),
	}

	for apiName, url := range urls {
		go fetchFromAPI(ctx, url, ch, apiName)
	}

	select {
	case result := <-ch:
		return result, nil
	case <-ctx.Done():
		return Result{}, errors.New("timeout exceeded")
	}
}
