package middleware

import (
	"net/http"

	"goapp/services"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(authService *services.AuthService) gin.HandlerFunc {

	return func(c *gin.Context) {

		tokenStr, err := c.Cookie("access_token")

		if err != nil || tokenStr == "" {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				gin.H{"error": "missing token"},
			)
			return
		}

		token, err := authService.ValidateAccessToken(tokenStr)

		if err != nil || token == nil || !token.Valid {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				gin.H{"error": "invalid token"},
			)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)

		if !ok {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				gin.H{"error": "invalid claims"},
			)
			return
		}

		userID, ok := claims["sub"].(string)

		if !ok {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				gin.H{"error": "no user id"},
			)
			return
		}

		c.Set("currentUser", userID)

		c.Next()
	}
}
