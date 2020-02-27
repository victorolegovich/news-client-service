package main

import (
	"github.com/gorilla/pat"
	"go.uber.org/zap"
	"net/http"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		println("failed to create a logger zap")
		return
	}

	router := pat.New()
	router.Get("/news/{id:[0-9A-z]+}", getNewsItemH)
	http.Handle("/", router)

	if err = http.ListenAndServe("localhost:8080", nil); err != nil {
		logger.Info("failed to start the server", zap.Error(err))
		return
	}
}
