package main

import (
	"context"
	"encoding/json"
	"fmt"
	model "github.com/murilocarbol/multithreading/cmd/model"
	"net/http"
	"time"
)

const (
	CEP       = "13201005"
	BrasilAPI = "BrasilAPI"
	ViaCEP    = "ViaCEP"
	timeout   = 1 * time.Second
)

func fetchingFromAPI(ctx context.Context, cep string, source string, address interface{}, resultChan chan<- model.CepResponse) {
	req, err := http.NewRequestWithContext(ctx, "GET", formatUrl(cep, source), nil)
	if err != nil {
		resultChan <- model.CepResponse{Error: err}
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		resultChan <- model.CepResponse{Error: err}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		resultChan <- model.CepResponse{Error: fmt.Errorf("unexpected status code: %d", resp.StatusCode)}
		return
	}

	if err := json.NewDecoder(resp.Body).Decode(&address); err != nil {
		resultChan <- model.CepResponse{Error: err}
		return
	}

	resultChan <- model.CepResponse{Address: address, Source: source}
}

func main() {

	fmt.Printf("Iniciando processamento de busca para o cep: %s\n", CEP)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resultChan := make(chan model.CepResponse, 2)

	go fetchingFromAPI(ctx, CEP, BrasilAPI, &model.BrasilApiResponse{}, resultChan)
	go fetchingFromAPI(ctx, CEP, ViaCEP, &model.ViaCepResponse{}, resultChan)

	// Valida a primeira resposta ou timeout
	select {
	case result := <-resultChan:
		if result.Error != nil {
			fmt.Println(result.Error)
		} else {
			fmt.Printf("Source: %s\nAddress: %+v\n", result.Source, result.Address)
		}
	case <-ctx.Done():
		fmt.Println("Error: timeout")
	}
}

func formatUrl(cep, apiName string) string {
	if apiName == BrasilAPI {
		return fmt.Sprintf("https://brasilapi.com.br/api/cep/v1/%s", cep)
	} else {
		return fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep)
	}
}
