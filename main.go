package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	ctrl "github.com/adopabianko/p2p-auth/controllers"
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", ctrl.IndexPage).Methods("GET")
	r.HandleFunc("/auth/register", ctrl.RegisterPage).Methods("POST")
	r.HandleFunc("/auth/check-user-account", ctrl.CheckUserAccountPage).Methods("GET")
	r.HandleFunc("/auth/verification-account", ctrl.VerificationAccountPage).Methods("POST")
	r.HandleFunc("/auth/login", ctrl.LoginPage).Methods("POST")

	fmt.Print("Starting web server at http://localhost:3333/")
	log.Fatal(http.ListenAndServe(":3333", r))
}