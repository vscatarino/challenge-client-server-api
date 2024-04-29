package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Exchange struct {
	Bid string `json:"bid"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)

	if err != nil {
		log.Printf("Não foi possível criar a requisição. %v", err)
		panic(err)
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Printf("Não foi possível realizar a requisição. %v", err)
		panic(err)
	}
	defer resp.Body.Close()

	f, err := os.Create("cotacao.txt")

	if err != nil {
		log.Printf("Não foi possível criar o arquivo cotacao. %v", err)
		panic(err)
	}
	defer f.Close()

	body, err := io.ReadAll(resp.Body)

	var exchange Exchange
	err = json.Unmarshal(body, &exchange)
	if err != nil {
		log.Printf("Não foi possível realizar o parse dos dados. %v", err)
		panic(err)
	}

	_, err = f.WriteString("Dólar:{" + exchange.Bid + "}")
	if err != nil {
		log.Printf("Não foi possível alterar o arquivo cotacao.txt. %v", err)
		panic(err)
	}

}
