package main

import (
	"github.com/gorilla/pat"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetNewsItemH(t *testing.T) {
	router := pat.New()
	router.Get("/news/{id:[0-9A-z]+}", getNewsItemH)
	http.Handle("/", router)

	server := httptest.NewServer(router)
	defer server.Close()

	url := server.URL + "/news/koe"

	resp, err := http.Get(url)
	if err != nil {
		t.Error(err.Error())
		return
	}

	if resp.StatusCode == http.StatusInternalServerError {
		t.Error(err)
		return
	}
}
