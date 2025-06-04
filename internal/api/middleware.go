package api

import (
	"github.com/gin-gonic/gin"

	keycloakauth "github.com/JorgeSaicoski/keycloak-auth"
)

func AuthMiddleware() gin.HandlerFunc {
	config := keycloakauth.DefaultConfig()
	config.LoadFromEnv() // This loads KEYCLOAK_URL and KEYCLOAK_REALM from environment

	// Add any additional configuration
	config.SkipPaths = []string{"/health"}
	config.RequiredClaims = []string{"sub", "preferred_username"}

	return keycloakauth.SimpleAuthMiddleware(config)
}
