package mocks

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type MockConsumerService struct {
}

var defaultPath = "./uploads/"

func (u MockConsumerService) Run() {
	r := mux.NewRouter()
	http.HandleFunc("/{key}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		dataHash := mux.Vars(r)["key"]

		file, err := os.CreateTemp(defaultPath, dataHash)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer file.Close()

		_, err = io.Copy(file, r.Body)
		if err != nil {
			os.Remove(file.Name())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	addr := ":8080"
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}
	log.Infof("Starting server at %s", addr)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("ListenAndServe(): %v", err)
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
