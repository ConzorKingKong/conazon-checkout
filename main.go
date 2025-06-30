package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/conzorkingkong/conazon-checkout/config"
	"github.com/conzorkingkong/conazon-checkout/controllers"
	authcontrollers "github.com/conzorkingkong/conazon-users-and-auth/controllers"
	"github.com/joho/godotenv"
)

var PORT, PORTExists = "", false
var JwtSecret, jwtSecretExists = "", false
var DatabaseURLEnv, DatabaseURLExists = "", false
var EmailPassword, EmailPasswordExists = "", false
var RabbitMQURL, RabbitMQURLExists = "", false

func main() {

	godotenv.Load()

	PORT, PORTExists = os.LookupEnv("PORT")
	JwtSecret, jwtSecretExists = os.LookupEnv("JWTSECRET")
	EmailPassword, EmailPasswordExists = os.LookupEnv("EMAILPASSWORD")
	DatabaseURLEnv, DatabaseURLExists = os.LookupEnv("DATABASEURL")
	RabbitMQURL, RabbitMQURLExists = os.LookupEnv("RABBITMQURL")

	if !jwtSecretExists || !DatabaseURLExists {
		log.Fatal("Required environment variable not set")
	}

	if !RabbitMQURLExists {
		RabbitMQURL = "amqp://guest:guest@rabbitmq"
	}

	if !PORTExists {
		PORT = "8083"
	}

	config.SECRETKEY = []byte(JwtSecret)
	config.DatabaseURLEnv = DatabaseURLEnv
	config.EmailPassword = EmailPassword
	config.RabbitMQURL = RabbitMQURL

	http.HandleFunc("/", authcontrollers.Root)

	http.HandleFunc("/checkout/", controllers.CheckoutHandler)
	http.HandleFunc("/checkout/{id}", controllers.CheckoutId)
	http.HandleFunc("/checkout/user/{id}", controllers.UserId)

	http.HandleFunc("/healthz", authcontrollers.Healthz)

	fmt.Println("server starting on port", PORT)
	http.ListenAndServe(":"+PORT, nil)
}
