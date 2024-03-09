package main

import (
	"github.com/hnimtadd/run/sdk"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func handleLogin(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("hello from the login handler YADA"))
}

func handleDashboard(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("hello from the dashboard handler"))
}

func handleIndex(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("login page: <a href=\"/login\" /><br />Dashboard page: <a href=\"/dashboard\" />"))
}

func main() {
	router := chi.NewMux()
	router.Get("/dashboard", handleDashboard)
	router.Get("/login", handleLogin)
	router.Get("/", handleIndex)
	sdk.Handle(router)
}
