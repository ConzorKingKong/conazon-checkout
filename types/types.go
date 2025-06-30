package types

import "time"

type Checkout struct {
	Id             int       `json:"id"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	UserId         int       `json:"userId"`
	TotalPrice     string    `json:"totalPrice"`
	BillingStatus  string    `json:"billingStatus"`
	ShippingStatus string    `json:"shippingStatus"`
	TrackingNumber string    `json:"trackingNumber"`
}

type CheckoutResponse struct {
	Status  int      `json:"status"`
	Message string   `json:"message"`
	Data    Checkout `json:"data"`
}

type CheckoutsResponse struct {
	Status  int        `json:"status"`
	Message string     `json:"message"`
	Data    []Checkout `json:"data"`
}