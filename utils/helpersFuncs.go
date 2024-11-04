package utils

import (
	"log"
	"os"
	"slices"
	"strings"
)

// IsAllowedIncomingRequestIP checks if the incoming request is from Paystack
func IsAllowedIncomingRequestIP(remoteIP string) bool {
	// check if the request is from paystack
	allowedIPs := os.Getenv("ALLOWED_IP_ADDRESSES")
	allowedIPAddressesList := strings.Split(allowedIPs, ",")
	if !slices.Contains(allowedIPAddressesList, remoteIP) {
		log.Println("Request is not from Paystack. IP:", remoteIP)
		return false
	}
	return true
}
