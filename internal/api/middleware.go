package api

import (
	"crypto/rsa"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

		// Parse and validate the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			// Get the public key from environment variable
			publicKey, err := getKeycloakPublicKeyFromEnv()
			if err != nil {
				return nil, err
			}
			return publicKey, nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: " + err.Error()})
			c.Abort()
			return
		}

		// Extract claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Store user info in the context
			c.Set("userID", claims["sub"])
			c.Set("username", claims["preferred_username"])
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}
	}
}

// Get the Keycloak public key from environment variable
func getKeycloakPublicKeyFromEnv() (*rsa.PublicKey, error) {
	// Get the public key from environment variable
	publicKeyBase64 := os.Getenv("KEYCLOAK_PUBLIC_KEY")
	if publicKeyBase64 == "" {
		return nil, fmt.Errorf("KEYCLOAK_PUBLIC_KEY environment variable not set")
	}

	// Format the key with proper PEM structure
	publicKeyPEM := fmt.Sprintf("-----BEGIN PUBLIC KEY-----\n%s\n-----END PUBLIC KEY-----", publicKeyBase64)

	// Parse the PEM encoded public key
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(publicKeyPEM))
	if err != nil {
		return nil, fmt.Errorf("error parsing public key: %v", err)
	}

	return publicKey, nil
}
