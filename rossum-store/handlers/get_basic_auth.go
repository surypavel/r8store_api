package handlers

import (
	"encoding/base64"
	"fmt"
	"strings"
)

func GetBasicAuth(authHeader string) (string, string, error) {
	if authHeader == "" {
		return "", "", nil
	}

	if !strings.HasPrefix(authHeader, "Basic ") {
		return "", "", fmt.Errorf("invalid Authorization header")
	}

	// Remove "Basic " prefix
	encodedCredentials := strings.TrimPrefix(authHeader, "Basic ")

	// Decode base64 credentials
	decodedBytes, err := base64.StdEncoding.DecodeString(encodedCredentials)
	if err != nil {
		return "", "", fmt.Errorf("failed to decode credentials: %v", err)
	}

	// Split into username and password
	credentials := strings.SplitN(string(decodedBytes), ":", 2)
	if len(credentials) != 2 {
		return "", "", fmt.Errorf("invalid credential format")
	}

	return credentials[0], credentials[1], nil
}
