package main

import (
	"log"
	"os"

	awslambda "github.com/chukwuka-emi/healthdecodepro/paystack_webhook/platform/aws"
	"github.com/chukwuka-emi/healthdecodepro/paystack_webhook/utils"
	"go.mongodb.org/mongo-driver/mongo"
)

// DBClient is the MongoDB client
var DBClient *mongo.Client

func init() {
	db, err := utils.ConnectDB(os.Getenv("DB_URL"))
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	DBClient = db
}

func main() {
	awslambda.StartLambdaHandler(DBClient)
}
