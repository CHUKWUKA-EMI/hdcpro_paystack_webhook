package types

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Purpose is the purpose of a payment
type Purpose string

const (
	// SubscriptionAdded is a payment for a subscription
	SubscriptionAdded Purpose = "subscriptionAdded"
	// SubscriptionUpdated is a payment for a subscription
	SubscriptionUpdated Purpose = "subscriptionUpdated"
	// CreditTopUp is a payment for topping up credit
	CreditTopUp Purpose = "creditTopUp"
)

// CustomFields is a paystack event custom fields
type CustomFields struct {
	DisplayName  string `json:"display_name"`
	VariableName string `json:"variable_name"`
	Value        string `json:"value"`
}

// Metadata is a paystack event metadata
type Metadata struct {
	CancelAction   string         `json:"cancel_action"`
	PaymentPurpose Purpose        `json:"payment_purpose"`
	CustomFields   []CustomFields `json:"custom_fields"`
}

// Customer is a paystack customer
type Customer struct {
	Email string `json:"email"`
}

// Data is a paystack event data
type Data struct {
	ID          uint     `json:"id"`
	Reference   string   `json:"reference"`
	Amount      float32  `json:"amount"`
	Currency    string   `json:"currency"`
	Transaction string   `json:"transaction"`
	Status      string   `json:"status"`
	IP          string   `json:"ip"`
	Channel     string   `json:"channel"`
	Metadata    Metadata `json:"metadata"`
	Customer    Customer `json:"customer"`
}

// PaystackEvent is a paystack event
type PaystackEvent struct {
	Event string `json:"event"`
	Data  Data   `json:"data"`
}

// User is a user
type User struct {
	ID               primitive.ObjectID `json:"id" bson:"_id"`
	Email            string             `json:"email" bson:"email"`
	Credits          *float32           `json:"credits" bson:"credits"`
	UserSubscription *UserSubscription  `json:"subscription" bson:"subscription"`
}

// UserSubscription is a user subscription
type UserSubscription struct {
	SubscriptionID string    `json:"subscription_id" binding:"required" bson:"subscription_id"`
	IsActive       bool      `json:"is_active" bson:"is_active"`
	CreatedAt      time.Time `json:"createdAt" bson:"created_at"`
	UpdatedAt      time.Time `json:"updatedAt" bson:"updated_at"`
}
