package main

import (
	"fmt"
	"net/http"

	sdk "github.com/hnimtadd/run/sdk/go"

	"github.com/go-chi/chi/v5"
)

func handleLogin(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("hello from the login handler YADA"))
	fmt.Println("enter login")
}

func handleDashboard(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("hello from the dashboard handler"))
	fmt.Println("enter dashboard")
}

func handleIndex(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "text/html")
	_, _ = w.Write([]byte("<html>login page: <a href=\"/login\" /><br />Dashboard page: <a href=\"/dashboard\" /></html>"))
	fmt.Println("enter index")
}

func main() {
	router := chi.NewMux()
	router.Get("/dashboard", handleDashboard)
	router.Get("/login", handleLogin)
	router.Get("/", handleIndex)

	sdk.Handle(router)
}
