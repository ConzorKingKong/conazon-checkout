package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/conzorkingkong/conazon-checkout/config"
	"github.com/conzorkingkong/conazon-checkout/token"
	"github.com/conzorkingkong/conazon-checkout/types"
	emailtypes "github.com/conzorkingkong/conazon-email-service/types"
	authhelpers "github.com/conzorkingkong/conazon-users-and-auth/helpers"
	authtypes "github.com/conzorkingkong/conazon-users-and-auth/types"
	"github.com/jackc/pgx/v5"
	amqp "github.com/rabbitmq/amqp091-go"
)

func CheckoutHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// GET ALL ORDERS
	if r.Method == "GET" {

		TokenData, err := token.ValidateAndReturnSession(w, r)
		if err != nil {
			return
		}

		conn, err := pgx.Connect(context.Background(), config.DatabaseURLEnv)

		if err != nil {
			log.Printf("Error connecting to database: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(authtypes.Response{Status: http.StatusInternalServerError, Message: "internal service error", Data: ""})
			return
		}

		defer conn.Close(context.Background())

		rows, err := conn.Query(context.Background(), "select id, created_at, updated_at, user_id, total_price, cart_item_ids, cart_snapshot from checkout.checkout where user_id=$1", TokenData.Id)

		if err != nil {
			log.Printf("Error getting checkouts with id %d - %s", TokenData.Id, err)
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(authtypes.Response{Status: http.StatusNotFound, Message: "Checkouts not found", Data: ""})
			return
		}

		var rowSlice []types.Checkout

		for rows.Next() {
			var checkout types.Checkout
			err = rows.Scan(&checkout.ID, &checkout.CreatedAt, &checkout.UpdatedAt, &checkout.UserID, &checkout.TotalPrice, &checkout.CartItemIDs, &checkout.CartSnapshot)
			if err != nil {
				log.Printf("Error getting checkout with id %d - %s", TokenData.Id, err)
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(authtypes.Response{Status: http.StatusNotFound, Message: "Error loading checkout", Data: ""})
				return
			}
			rowSlice = append(rowSlice, checkout)
		}

		defer rows.Close()

		// We error on no checkouts found. Maybe just return an empty one. Double check this
		if rowSlice == nil {
			log.Printf("Error: No checkouts found for user %d", TokenData.Id)
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(authtypes.Response{Status: http.StatusNotFound, Message: "No checkouts found for user", Data: ""})
			return
		}

		json.NewEncoder(w).Encode(types.CheckoutsResponse{Status: http.StatusOK, Message: "Success", Data: rowSlice})

		// CREATE ORDER
	} else if r.Method == "POST" {

		TokenData, err := token.ValidateAndReturnSession(w, r)
		if err != nil {
			return
		}

		// receive user info, cart array, and total price
		// Later, user won't send anything, we will
		// read their id and check all active cart items
		// and price ourselves
		call := types.CheckoutCall{}

		err = json.NewDecoder(r.Body).Decode(&call)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// reach out for item total prices
		// handle payment
		// update cart database as purchased

		conn, err := pgx.Connect(context.Background(), config.DatabaseURLEnv)

		if err != nil {
			log.Printf("Error connecting to database: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(authtypes.Response{Status: http.StatusInternalServerError, Message: "internal service error", Data: ""})
			return
		}

		defer conn.Close(context.Background())

		var cartItemIDs []int
		for _, cart := range call.Carts {
			cartItemIDs = append(cartItemIDs, cart.ID)
		}

		cartSnapshot, err := json.Marshal(call.Carts)
		if err != nil {
			log.Printf("Error marshaling cart snapshot: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(authtypes.Response{Status: http.StatusInternalServerError, Message: "internal service error", Data: ""})
			return
		}

		checkout := types.Checkout{
			UserID:       TokenData.Id,
			TotalPrice:   strconv.Itoa(call.Total),
			CartItemIDs:  cartItemIDs,
			CartSnapshot: cartSnapshot,
		}

		queryString := "insert into checkout.checkout (user_id, total_price, cart_item_ids, cart_snapshot, created_at, updated_at) values ($1, $2, $3, $4, NOW(), NOW()) returning id"

		err = conn.QueryRow(context.Background(), queryString, checkout.UserID, checkout.TotalPrice, checkout.CartItemIDs, checkout.CartSnapshot).Scan(&checkout.ID)
		if err != nil {
			log.Printf("Error saving user: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(authtypes.Response{Status: http.StatusInternalServerError, Message: "Internal Service Error", Data: ""})
			return
		}

		// send confirmation email
		email := emailtypes.Email{
			Checkout: checkout,
			User:     call.User,
		}

		mqConn, err := amqp.Dial(config.RabbitMQURL)
		failOnError(err, "Failed to connect to RabbitMQ")
		defer mqConn.Close()

		mqCh, mqErr := mqConn.Channel()
		failOnError(mqErr, "Failed to open a channel")
		defer mqCh.Close()

		q, err := mqCh.QueueDeclare(
			"email", // name
			false,   // durable
			false,   // delete when unused
			false,   // exclusive
			false,   // no-wait
			nil,     // arguments
		)
		failOnError(err, "Failed to declare a queue")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		body, err := json.Marshal(email)
		failOnError(err, "Failed to marshal email struct")

		err = mqCh.PublishWithContext(ctx,
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(body),
			})
		failOnError(err, "Failed to publish a message")
		log.Printf(" [x] Sent %s\n", body)

		json.NewEncoder(w).Encode(types.CheckoutResponse{Status: http.StatusOK, Message: "Success", Data: checkout})

	} else {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(authtypes.Response{Status: http.StatusBadRequest, Message: "Bad Request", Data: ""})
		return
	}

}

func CheckoutId(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "GET" {
		routeId, parsedRouteId, err := authhelpers.RouteIdHelper(w, r)
		if err != nil {
			return
		}

		TokenData, err := token.ValidateAndReturnSession(w, r)
		if err != nil {
			return
		}

		conn, err := pgx.Connect(context.Background(), config.DatabaseURLEnv)
		if err != nil {
			log.Printf("Error connecting to database: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(authtypes.Response{Status: http.StatusInternalServerError, Message: "internal service error", Data: ""})
			return
		}

		defer conn.Close(context.Background())

		checkout := types.Checkout{}

		// verify owner of checkout with db call
		err = conn.QueryRow(context.Background(), "select id, created_at, updated_at, user_id, total_price, cart_item_ids, cart_snapshot from checkout.checkout where id=$1", routeId).Scan(&checkout.ID, &checkout.CreatedAt, &checkout.UpdatedAt, &checkout.UserID, &checkout.TotalPrice, &checkout.CartItemIDs, &checkout.CartSnapshot)
		if err != nil {
			log.Printf("Error getting checkout with id %s - %s", routeId, err)
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(authtypes.Response{Status: http.StatusNotFound, Message: "checkout not found", Data: ""})
			return
		}

		if TokenData.Id != checkout.UserID {
			log.Printf("Error: user tried reading checkout they don't own")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(authtypes.Response{Status: http.StatusUnauthorized, Message: "Unauthorized", Data: ""})
			return
		}

		fmt.Printf("%d %s %d", TokenData.Id, routeId, parsedRouteId)

		json.NewEncoder(w).Encode(types.CheckoutResponse{Status: http.StatusOK, Message: "Success", Data: checkout})

	} else {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(authtypes.Response{Status: http.StatusBadRequest, Message: "Bad Request", Data: ""})
		return
	}

}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
