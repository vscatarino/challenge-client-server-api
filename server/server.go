package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type QuotationDTO struct {
	Usdbrl struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

type Quotation struct {
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
	gorm.Model
}

type Exchange struct {
	Bid string `json:"bid"`
}

func main() {
	db, err := gorm.Open(sqlite.Open("exchange.db"), &gorm.Config{})

	if err != nil {
		log.Printf("Não foi possível criar uma conexão com a base de dados. %v", err)
		panic(err)
	}
	db.AutoMigrate(&Quotation{})

	http.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		getQuotationAndSave(w, r, db)
	})
	http.ListenAndServe(":8080", nil)
}

func getQuotationAndSave(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	cotx := r.WithContext(ctx)
	req, err := http.NewRequestWithContext(cotx.Context(), "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		log.Printf("getQuotation: Ocorreu um erro ao montar a requisição: %v \n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("getQuotation: Ocorreu um erro ao executar a requisição: %v \n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("getQuotation: Ocorreu um erro ao tentar ler o corpo da requisição: %v \n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var quotationDTO QuotationDTO
	err = json.Unmarshal(body, &quotationDTO)
	if err != nil {
		log.Printf("getQuotation: Não foi possível criar a estrutura de dados QuotationDTO. %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	quotation := Quotation{
		Code:       quotationDTO.Usdbrl.Code,
		Codein:     quotationDTO.Usdbrl.Codein,
		Name:       quotationDTO.Usdbrl.Name,
		High:       quotationDTO.Usdbrl.High,
		Low:        quotationDTO.Usdbrl.Low,
		VarBid:     quotationDTO.Usdbrl.VarBid,
		PctChange:  quotationDTO.Usdbrl.PctChange,
		Bid:        quotationDTO.Usdbrl.Bid,
		Ask:        quotationDTO.Usdbrl.Ask,
		Timestamp:  quotationDTO.Usdbrl.Timestamp,
		CreateDate: quotationDTO.Usdbrl.CreateDate,
	}

	err = saveQuotation(&quotation, db)
	if err != nil {
		log.Printf("getQuotation: Não foi possível salvar os dados em nossa base de dados: %v \n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	exchange := Exchange{quotation.Bid}
	jsonExchange, err := json.Marshal(exchange)
	if err != nil {
		log.Printf("getQuotation: Não foi possível criar o json de resposta. %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// var quotations []Quotation
	// db.Find(&quotations)
	// for _, quot := range quotations {
	// 	fmt.Printf("quotation %v \n", quot)
	// }

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonExchange)
}

func saveQuotation(quotation *Quotation, db *gorm.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	return db.WithContext(ctx).Create(quotation).Error
}
