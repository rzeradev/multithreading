# Golang Multithreading Challenge

Este projeto implementa um desafio de multithreading em Golang, onde são feitas requisições simultâneas a duas APIs para buscar o resultado mais rápido. O objetivo é:

- Acatar a API que entregar a resposta mais rápida e descartar a resposta mais lenta.
- Exibir o resultado da requisição no terminal com os dados do endereço, bem como qual API enviou a resposta.
- Limitar o tempo de resposta em 1 segundo. Caso contrário, exibir uma mensagem de erro de timeout.

As APIs usadas são:

- https://brasilapi.com.br/api/cep/v1/ + cep
- http://viacep.com.br/ws/" + cep + "/json/

## Estrutura de Pastas

O projeto segue a estrutura de pastas recomendada pelo Golang:

```
multithreading/
├── cmd/
│   └── main.go
├── pkg/
│   ├── cepservice/
│   │   └── cepservice.go
├── go.mod
├── go.sum
└── readme.md

```

- `cmd/main.go`: Ponto de entrada da aplicação.
- `pkg/cepservice/cepservice.go`: Lógica para fazer requisições para as APIs e retornar a resposta mais rápida.
- `go.mod` e `go.sum`: Gerenciamento de dependências do projeto.

## Implementação

### cmd/main.go

Este arquivo contém o ponto de entrada da aplicação. Configura um contexto com timeout de 1 segundo e chama a função `GetFasterAPIResult` para buscar o resultado mais rápido entre as duas APIs.

```go
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/rzeradev/multithreading/pkg/cepservice"
)

func main() {
	cep := "70150900" // CEP Palácio do Planalto
	if len(os.Args) > 1 {
		cep = os.Args[1]
	}
	timeout := 1 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	result, err := cepservice.GetFasterAPIResult(ctx, cep)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("API:", result.API)
	fmt.Printf("Address: %+v\n", result.Address)
}

```

### pkg/cepservice/cepservice.go

Este arquivo contém a lógica para fazer as requisições para as APIs e retornar a resposta mais rápida. A função `fetchFromAPI` faz a requisição para uma API e envia o resultado através de um canal. A função `GetFasterAPIResult` inicia as goroutines para ambas as APIs e espera pelo primeiro resultado ou um timeout.

```go
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
```

## Dependências

O projeto utiliza a biblioteca [resty](https://github.com/go-resty/resty) para fazer as requisições HTTP. Esta dependência está especificada no arquivo `go.mod`.

## Executando o Projeto

Para rodar o projeto, siga os seguintes passos:

1. Clone o repositório:

   ```sh
   git clone https://github.com/rzeradev/multithreading.git
   ```

2. Navegue até o diretório do projeto:

   ```sh
   cd multithreading
   ```

3. Instale as dependências:

   ```sh
   go mod tidy
   ```

4. Execute o projeto passando um cep:

   ```sh
   go run ./cmd/main.go 01153000
   ```

5. Opcionalmente se não passar um cep o cep consultado será o 70150900:
   ```sh
   go run ./cmd/main.go
   ```

Isso fará com que as requisições sejam feitas simultaneamente para as duas APIs, exibindo no terminal o resultado mais rápido e a API que enviou a resposta. Se ambas as requisições demorarem mais de 1 segundo, será exibida uma mensagem de erro de timeout.

## Licença

Este projeto está licenciado sob a Licença MIT. Veja o arquivo [LICENSE](LICENSE) para mais detalhes.
