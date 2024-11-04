package awslambda

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/chukwuka-emi/healthdecodepro/paystack_webhook/api"
	"github.com/chukwuka-emi/healthdecodepro/paystack_webhook/types"
	"github.com/chukwuka-emi/healthdecodepro/paystack_webhook/utils"
	"go.mongodb.org/mongo-driver/mongo"
)

func lambdaHandler(db *mongo.Client) func(events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return func(request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		if !utils.IsAllowedIncomingRequestIP(request.RequestContext.HTTP.SourceIP) {
			errorMsg := "Request is not from Paystack"
			return handleResponse(http.StatusForbidden, true, errorMsg)
		}

		if !isValidHTTPRequest(request) {
			errorMsg := "Not Found"
			return handleResponse(http.StatusBadRequest, true, errorMsg)
		}

		var event types.PaystackEvent
		json.Unmarshal([]byte(request.Body), &event)

		log.Printf("Event: %+v\n", event)
		if event.Event == "charge.success" {
			err := api.ProcessPaystackWebhook(event, db)
			if err != nil {
				log.Println("Error processing payment event:", err)
				return handleResponse(http.StatusInternalServerError, true, err.Error())
			}
			log.Println("Payment event processed successfully")
		}

		// return 200 OK response to paystack
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 200,
		}, nil
	}
}

func isValidHTTPRequest(request events.APIGatewayV2HTTPRequest) bool {
	if request.RawPath != fmt.Sprintf("/%s/paystack-webhook", os.Getenv("ENV")) || request.RequestContext.HTTP.Method != "POST" {
		return false
	}
	return true
}

func handleResponse(statusCode int, isError bool, responseBody interface{}) (events.APIGatewayV2HTTPResponse, error) {
	var b map[string]string
	if isError {
		b = map[string]string{"error": (responseBody).(string)}
	} else {
		b = map[string]string{"message": (responseBody).(string)}
	}

	body, _ := json.Marshal(b)
	return events.APIGatewayV2HTTPResponse{
		Body:       string(body),
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
}

// StartLambdaHandler is a function that registers the aws lambda handler
func StartLambdaHandler(db *mongo.Client) {
	handler := lambdaHandler(db)
	lambda.Start(handler)
}
