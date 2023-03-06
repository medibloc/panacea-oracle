package mocks

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/btcsuite/btcd/btcec"
	"github.com/gorilla/mux"
)

type MockConsumerService struct {
	oraclePubKey *btcec.PublicKey
}

var defaultPath = "./uploads/"

func (u MockConsumerService) RunServer() {
	r := mux.NewRouter()
	r.HandleFunc("/v1/data/{dealID}/{dataHash}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		dealID := vars["dealID"]
		dataHash := vars["dataHash"]

		filename := filepath.Join(defaultPath, dealID+"_"+dataHash)
		err := os.MkdirAll(defaultPath, os.ModePerm)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		file, err := os.Create(filename)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		_, err = file.Write(body)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}).Methods("POST")

	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	if err := server.ListenAndServe(); err != nil {
		fmt.Printf("listen: %s\n", err)
	}
}

func (u MockConsumerService) Get(dealID, dataHash string) ([]byte, error) {
	return os.ReadFile(filepath.Join(defaultPath, dealID+"_"+dataHash))
}

func RemoveMockData() {
	if info, _ := os.Stat(defaultPath); info != nil {
		os.RemoveAll(defaultPath)
	}
}
