package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func DoctorAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			secret = "pharmeasy_super_secret_key_2024"
		}

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		claims := token.Claims.(jwt.MapClaims)

		// Must be doctor role
		if claims["role"] != "doctor" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Doctor access required"})
			c.Abort()
			return
		}

		doctorID := uint(claims["doctor_id"].(float64))
		fmt.Println("🩺 Authenticated doctor_id:", doctorID)
		c.Set("doctor_id", doctorID)
		c.Next()
	}
}
