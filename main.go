package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

var PORT, PORTExists = "", false
var JwtSecret, jwtSecretExists = "", false
var DatabaseURLEnv, DatabaseURLExists = "", false

var SECRETKEY []byte

func main() {

	godotenv.Load()

	PORT, PORTExists = os.LookupEnv("PORT")
	JwtSecret, jwtSecretExists = os.LookupEnv("JWTSECRET")
	DatabaseURLEnv, DatabaseURLExists = os.LookupEnv("DATABASEURL")

	SECRETKEY = []byte(JwtSecret)

	if !jwtSecretExists || !DatabaseURLExists {
		log.Fatal("Required environment variable not set")
	}

	if !PORTExists {
		PORT = "8083"
	}

	http.HandleFunc("/", Root)
	http.HandleFunc("/checkout/", CheckoutHandler)
	http.HandleFunc("/checkout/{id}", CheckoutId)
	http.HandleFunc("/checkout/user/{id}", UserId)

	fmt.Println("server starting on port", PORT)
	http.ListenAndServe(":"+PORT, nil)
}
