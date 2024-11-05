package api

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/chukwuka-emi/healthdecodepro/paystack_webhook/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// ProcessPaystackWebhook processes incoming paystack webhook requests
func ProcessPaystackWebhook(payload types.PaystackEvent, dbClient *mongo.Client) error {
	context := context.Background()
	return processPaymentEvent(context, dbClient, payload)
}

func processPaymentEvent(context context.Context, dbClient *mongo.Client, event types.PaystackEvent) error {

	db := dbClient.Database("healthdecodepro_db")
	userCollection := db.Collection("users")

	user, err := getUserByEmail(context, userCollection, event.Data.Customer.Email)
	if err != nil {
		return err
	}

	if user.UserSubscription == nil {
		return errors.New("user subscription is null")
	}

	if err := validatePaymentPurpose(event); err != nil {
		return err
	}

	if err := processSubscriptionPayment(context, dbClient, userCollection, event, *user); err != nil {
		return err
	}

	return nil
}

func processSubscriptionPayment(ctx context.Context, dbClient *mongo.Client, userCollection *mongo.Collection, event types.PaystackEvent, user types.User) error {
	userSubscription := *user.UserSubscription

	if event.Data.Metadata.PaymentPurpose == types.CreditTopUp && !userSubscription.IsActive {
		return errors.New("user subscription is not active")
	}

	if event.Data.Metadata.PaymentPurpose != types.CreditTopUp && userSubscription.IsActive {
		return errors.New("user subscription is already active")
	}

	amountInNaira := event.Data.Amount / 100

	log.Println("CURRENT USER CREDITS == ", *user.Credits, "USER ID: ", user.ID)
	log.Println("AMOUNT TO ADD IN NAIRA == ", amountInNaira)
	log.Println("UPDATE ATTEMPTED AT:", time.Now())

	var totalCredits float32
	if user.Credits == nil {
		totalCredits = amountInNaira
	} else {
		totalCredits = *user.Credits + amountInNaira
	}

	log.Println("Updating user & subscription")

	filter := bson.M{"email": event.Data.Customer.Email}
	updateData := bson.M{
		"subscription.is_active":  true,
		"subscription.updated_at": time.Now(),
		"credits":                 totalCredits,
		"updated_at":              time.Now(),
	}

	if event.Data.Metadata.PaymentPurpose == types.SubscriptionAdded {
		updateData["onboarding_step"] = "onboarded"
	}

	// start a session
	session, err := dbClient.StartSession()
	if err != nil {
		return err
	}
	// defer session end
	defer session.EndSession(ctx)

	// start transaction
	err = session.StartTransaction()
	if err != nil {
		return err
	}

	result, err := userCollection.UpdateOne(ctx, filter, bson.M{"$set": updateData})

	if err != nil {
		session.AbortTransaction(ctx)
		return err
	}

	if result.MatchedCount == 0 {
		session.AbortTransaction(ctx)
		return errors.New("user not found")
	}

	if err := session.CommitTransaction(ctx); err != nil {
		return err
	}

	return nil
}

func getUserByEmail(ctx context.Context, userCollection *mongo.Collection, email string) (*types.User, error) {
	var user types.User
	err := userCollection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

func validatePaymentPurpose(event types.PaystackEvent) error {
	switch event.Data.Metadata.PaymentPurpose {
	case types.SubscriptionAdded, types.SubscriptionUpdated, types.CreditTopUp:
		return nil
	default:
		return errors.New("invalid payment purpose")
	}
}
