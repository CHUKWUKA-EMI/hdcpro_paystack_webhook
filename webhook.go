package webhook

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

// import "$PROJECT_ROOT"

func init() {
	functions.HTTP("PaystackWebhook", PaystackWebhook)
}

func isAllowedIncomingRequestIP(remoteIP string) bool {
	// check if the request is from paystack
	allowedIPs := os.Getenv("ALLOWED_IP_ADDRESSES")
	allowedIPAddressesList := strings.Split(allowedIPs, ",")
	if !slices.Contains(allowedIPAddressesList, remoteIP) {
		log.Println("Request is not from Paystack")
		return false
	}
	return true
}

// PaystackWebhook is an HTTP Cloud Function with a request parameter.
func PaystackWebhook(w http.ResponseWriter, r *http.Request) {
	// validate incoming request IP
	if isAllowedIncomingRequestIP(r.Header.Get("X-Forwarded-For")) {

		var ctx = context.Background()
		//paystack event
		var event paystackEvent
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			log.Println("Error decoding request body:", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		log.Printf("Event: %+v\n", event)
		if event.Event == "charge.success" {

			dbConn := createDatabaseConnection(ctx)
			defer dbConn.Close(ctx)

			// process paystack event
			err := processPaymentEvent(ctx, dbConn, event)
			if err != nil {
				log.Println("Error processing payment event:", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			log.Println("Payment event processed successfully")
		}

		// return 200 OK response to paystack
		w.WriteHeader(http.StatusOK)

	} else {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
}
