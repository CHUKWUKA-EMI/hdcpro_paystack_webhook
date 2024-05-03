package webhook

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

type customFields struct {
	DisplayName  string `json:"display_name"`
	VariableName string `json:"variable_name"`
	Value        string `json:"value"`
}

type metadata struct {
	CancelAction   string         `json:"cancel_action"`
	PaymentPurpose Purpose        `json:"payment_purpose"`
	CustomFields   []customFields `json:"custom_fields"`
}
type customer struct {
	Email string `json:"email"`
}

type data struct {
	ID          uint     `json:"id"`
	Reference   string   `json:"reference"`
	Amount      float32  `json:"amount"`
	Currency    string   `json:"currency"`
	Transaction string   `json:"transaction"`
	Status      string   `json:"status"`
	IP          string   `json:"ip"`
	Channel     string   `json:"channel"`
	Metadata    metadata `json:"metadata"`
	Customer    customer `json:"customer"`
}

type paystackEvent struct {
	Event string `json:"event"`
	Data  data   `json:"data"`
}

type user struct {
	ID      int32    `json:"id"`
	Email   string   `json:"email"`
	Credits *float32 `json:"credits"`
}

type userSubscription struct {
	ID             int32  `json:"id"`
	UserID         int32  `json:"user_id" binding:"required"`
	SubscriptionID string `json:"subscription_id" binding:"required"`
	IsActive       bool   `json:"is_active"`
}
