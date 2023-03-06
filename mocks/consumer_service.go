package mocks

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
)

type MockConsumerService struct {
}

var defaultPath = "./uploads/"

func (u MockConsumerService) RunServer() {
	r := mux.NewRouter()
	r.HandleFunc("/{dataHash}", func(w http.ResponseWriter, r *http.Request) {

		dataHash := mux.Vars(r)["dataHash"]

		filename := filepath.Join(defaultPath, dataHash)
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
		file.Write(body)
	}).Methods("POST")

	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	if err := server.ListenAndServe(); err != nil {
		fmt.Printf("listen: %s\n", err)
	}
}

func (u MockConsumerService) Get(dataHash string) ([]byte, error) {
	return os.ReadFile(filepath.Join(defaultPath, dataHash))
}

func RemoveMockData() {
	if info, _ := os.Stat(defaultPath); info != nil {
		os.RemoveAll(defaultPath)
	}
}
