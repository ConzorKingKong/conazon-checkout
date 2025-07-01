package types

import (
	"encoding/json"
	"time"

	carttypes "github.com/conzorkingkong/conazon-cart/types"
	usertypes "github.com/conzorkingkong/conazon-users-and-auth/types"
)

type Checkout struct {
	ID           int             `json:"id" db:"id"`
	CreatedAt    time.Time       `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time       `json:"updatedAt" db:"updated_at"`
	UserID       int             `json:"userId" db:"user_id"`
	TotalPrice   string          `json:"totalPrice" db:"total_price"`
	CartItemIDs  []int           `json:"cartItemIds" db:"cart_item_ids"`
	CartSnapshot json.RawMessage `json:"cartSnapshot" db:"cart_snapshot"`
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

type CheckoutCall struct {
	User  usertypes.User   `json:"user"`
	Carts []carttypes.Cart `json:"carts"`
	Total int              `json:"total"`
}
