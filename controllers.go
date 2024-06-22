package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
)

func routeIdHelper(w http.ResponseWriter, r *http.Request) (string, int, error) {
	routeId := r.PathValue("id")

	parsedRouteId, err := strconv.Atoi(routeId)
	if err != nil {
		log.Printf("Error parsing route id: %s", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(GenericResponse{Status: http.StatusInternalServerError, Message: "Internal Service Error"})
		return "", 0, err
	}

	return routeId, parsedRouteId, nil
}

func Root(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GenericResponse{Status: 200, Message: "hello world"})
}

func CheckoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(GenericResponse{Status: http.StatusBadRequest, Message: "Bad Request"})
		return
	}

	TokenData, err := validateAndReturnSession(w, r)
	if err != nil {
		return
	}

	conn, err := pgx.Connect(context.Background(), DatabaseURLEnv)

	if err != nil {
		log.Printf("Error connecting to database: %s", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(GenericResponse{Status: http.StatusInternalServerError, Message: "internal service error"})
		return
	}

	defer conn.Close(context.Background())

	checkout := Checkout{
		UserId:         TokenData.Id,
		TotalPrice:     "0",
		BillingStatus:  "unpaid",
		ShippingStatus: "unshipped",
		TrackingNumber: "",
	}

	queryString := "insert into checkout.checkout (user_id, total_price, billing_status, shipping_status, tracking_number) values ($1, $2, $3, $4, $5) returning id"

	err = conn.QueryRow(context.Background(), queryString, checkout.UserId, checkout.TotalPrice, checkout.BillingStatus, checkout.ShippingStatus, checkout.TrackingNumber).Scan(&checkout.Id)
	if err != nil {
		log.Printf("Error saving user: %s", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(GenericResponse{Status: http.StatusInternalServerError, Message: "Internal Service Error"})
		return
	}

	// calculate total price (verify with my own database)

	// rabbitmq kicks off to double check inv, price,
	// update total price
	// handle payment
	// update status to paid
	// clear cart
	// kick off shipment/create tracking
	// email customer w tracking and order info

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(checkout)
}

func CheckoutId(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(GenericResponse{Status: http.StatusBadRequest, Message: "Bad Request"})
		return
	}

	routeId, parsedRouteId, err := routeIdHelper(w, r)
	if err != nil {
		return
	}

	TokenData, err := validateAndReturnSession(w, r)
	if err != nil {
		return
	}

	conn, err := pgx.Connect(context.Background(), DatabaseURLEnv)
	if err != nil {
		log.Printf("Error connecting to database: %s", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(GenericResponse{Status: http.StatusInternalServerError, Message: "internal service error"})
		return
	}

	defer conn.Close(context.Background())

	checkout := Checkout{}

	// verify owner of checkout with db call
	err = conn.QueryRow(context.Background(), "select id, user_id, total_price, billing_status, shipping_status, tracking_number from checkout.checkout where id=$1", routeId).Scan(&checkout.Id, &checkout.UserId, &checkout.TotalPrice, &checkout.BillingStatus, &checkout.ShippingStatus, &checkout.TrackingNumber)
	if err != nil {
		log.Printf("Error getting checkout with id %s - %s", routeId, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(GenericResponse{Status: http.StatusNotFound, Message: "checkout not found"})
		return
	}

	if TokenData.Id != checkout.UserId {
		log.Printf("Error: user tried reading checkout they don't own")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(GenericResponse{Status: http.StatusUnauthorized, Message: "Unauthorized"})
		return
	}

	fmt.Printf("%d %s %d", TokenData.Id, routeId, parsedRouteId)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(checkout)
}

func UserId(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(GenericResponse{Status: http.StatusBadRequest, Message: "Bad Request"})
		return
	}

	routeId, parsedRouteId, err := routeIdHelper(w, r)
	if err != nil {
		return
	}

	TokenData, err := validateAndReturnSession(w, r)
	if err != nil {
		return
	}

	conn, err := pgx.Connect(context.Background(), DatabaseURLEnv)

	if err != nil {
		log.Printf("Error connecting to database: %s", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(GenericResponse{Status: http.StatusInternalServerError, Message: "internal service error"})
		return
	}

	defer conn.Close(context.Background())

	rows, err := conn.Query(context.Background(), "select id, user_id, total_price, billing_status, shipping_status, tracking_number from checkout.checkout where user_id=$1", TokenData.Id)

	if err != nil {
		log.Printf("Error getting checkouts with id %s - %s", routeId, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(GenericResponse{Status: http.StatusNotFound, Message: "Checkouts not found"})
		return
	}

	var rowSlice []Checkout

	for rows.Next() {
		var checkout Checkout
		err = rows.Scan(&checkout.Id, &checkout.UserId, &checkout.TotalPrice, &checkout.BillingStatus, &checkout.ShippingStatus, &checkout.TrackingNumber)
		if err != nil {
			log.Printf("Error getting checkout with id %d - %s", TokenData.Id, err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(GenericResponse{Status: http.StatusNotFound, Message: "Error loading checkout"})
			return
		}
		rowSlice = append(rowSlice, checkout)
	}

	defer rows.Close()

	if rowSlice == nil {
		log.Printf("Error: No checkouts found for user %d", TokenData.Id)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(GenericResponse{Status: http.StatusNotFound, Message: "No checkouts found for user"})
		return
	}

	fmt.Printf("%d", parsedRouteId)

	// return data
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rowSlice)
}

// func rabbitMqConnect() {
// conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
// failOnError(err, "Failed to connect to RabbitMQ")
// defer conn.Close()

// ch, err := conn.Channel()
// failOnError(err, "Failed to open a channel")
// defer ch.Close()

// q, err := ch.QueueDeclare(
// 	"hello", // name
// 	false,   // durable
// 	false,   // delete when unused
// 	false,   // exclusive
// 	false,   // no-wait
// 	nil,     // arguments
// )
// failOnError(err, "Failed to declare a queue")

// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// defer cancel()

// body := "Hello World!"
// err = ch.PublishWithContext(ctx,
// 	"",     // exchange
// 	q.Name, // routing key
// 	false,  // mandatory
// 	false,  // immediate
// 	amqp.Publishing{
// 		ContentType: "text/plain",
// 		Body:        []byte(body),
// 	})
// failOnError(err, "Failed to publish a message")
// log.Printf(" [x] Sent %s\n", body)
// }

// func failOnError(err error, msg string) {
// 	if err != nil {
// 		log.Panicf("%s: %s", msg, err)
// 	}
// }
