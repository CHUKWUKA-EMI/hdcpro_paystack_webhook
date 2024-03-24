package webhook

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
)

const userTable string = "users"
const userSubscriptionTable string = "user_subscriptions"

func createDatabaseConnection(context context.Context) *pgx.Conn {
	connection, err := pgx.Connect(context, os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	if err := connection.Ping(context); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to ping database: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Connected to database")
	return connection
}

func getUserByEmail(tx pgx.Tx, email string) (*user, error) {
	var user user
	row := tx.QueryRow(context.Background(), "SELECT id, email, credits FROM users WHERE email = $1 FOR UPDATE", email)
	err := row.Scan(&user.ID, &user.Email, &user.Credits)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

func getUserSubscriptionByUserID(tx pgx.Tx, userSubscriptionID, userID uint64) (*userSubscription, error) {
	var userSubscription userSubscription
	row := tx.QueryRow(context.Background(), "SELECT id, user_id, subscription_id, is_active FROM user_subscriptions WHERE id=$1 AND user_id = $2", userSubscriptionID, userID)
	err := row.Scan(&userSubscription.ID, &userSubscription.UserID, &userSubscription.SubscriptionID, &userSubscription.IsActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user subscription not found")
		}
		return nil, err
	}
	return &userSubscription, nil
}

func processSubscriptionPayment(tx pgx.Tx, event paystackEvent, userSubscription *userSubscription) error {
	if event.Data.Metadata.PaymentPurpose == CreditTopUp {
		if !userSubscription.IsActive {
			return errors.New("user subscription is not active")
		}
		return nil
	}

	if event.Data.Metadata.PaymentPurpose == SubscriptionUpdated && userSubscription.IsActive {
		return errors.New("user subscription is already active")
	}

	log.Println("Updating user subscription")
	_, err := tx.Exec(context.Background(), "UPDATE user_subscriptions SET is_active = $1 WHERE id = $2", true, userSubscription.ID)
	if err != nil {
		return err
	}

	return nil
}

func updateOnboardingStep(tx pgx.Tx, userID uint64) error {
	_, err := tx.Exec(context.Background(), "UPDATE users SET onboarding_step = $1 WHERE id = $2", "onboarded", userID)
	return err
}

func updateUserCredits(tx pgx.Tx, user *user, amountInNaira float64) error {

	fmt.Println("CURRENT USER CREDITS == ", *user.Credits, "USER ID: ", user.ID)
	fmt.Println("AMOUNT TO ADD IN NAIRA == ", amountInNaira)
	fmt.Println("UPDATE ATTEMPTED AT:", time.Now())
	var totalCredits float64
	if user.Credits == nil {
		totalCredits = amountInNaira
	} else {
		totalCredits = *user.Credits + amountInNaira
	}
	_, err := tx.Exec(context.Background(), "UPDATE users SET credits = $1 WHERE id = $2", totalCredits, user.ID)
	return err
}

func extractUserSubscriptionID(event paystackEvent) (uint64, error) {
	for _, v := range event.Data.Metadata.CustomFields {
		if v.VariableName == "user_subscription_id" {
			userSubscriptionID, err := strconv.ParseUint(v.Value, 10, 64)
			if err != nil {
				return 0, err
			}
			return userSubscriptionID, nil
		}
	}
	return 0, errors.New("user subscription ID not found")
}

func parseAmount(amountStr string) (float64, error) {
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return 0, err
	}
	return amount / 100, nil
}

func validatePaymentPurpose(event paystackEvent) error {
	switch event.Data.Metadata.PaymentPurpose {
	case SubscriptionAdded, SubscriptionUpdated, CreditTopUp:
		return nil
	default:
		return errors.New("invalid payment purpose")
	}
}

func processPaymentEvent(context context.Context, conn *pgx.Conn, event paystackEvent) error {
	tx, err := conn.Begin(context)
	if err != nil {
		return err
	}

	defer tx.Rollback(context)

	user, err := getUserByEmail(tx, event.Data.Customer.Email)
	if err != nil {

		return err
	}

	userSubscriptionID, err := extractUserSubscriptionID(event)
	if err != nil {
		return err
	}

	userSubscription, err := getUserSubscriptionByUserID(tx, userSubscriptionID, user.ID)
	if err != nil {

		return err
	}

	amountInNaira := event.Data.Amount / 100

	if err := validatePaymentPurpose(event); err != nil {
		return err
	}

	if err := processSubscriptionPayment(tx, event, userSubscription); err != nil {
		return err
	}

	if err := updateUserCredits(tx, user, amountInNaira); err != nil {
		return err
	}

	if event.Data.Metadata.PaymentPurpose == SubscriptionAdded {
		if err := updateOnboardingStep(tx, user.ID); err != nil {
			return err
		}
	}

	return tx.Commit(context)
}
