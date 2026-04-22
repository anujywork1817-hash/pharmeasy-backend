package middleware

import (
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
		fmt.Println("🔑 Auth Header:", authHeader)

		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized - no token"})
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if len(tokenStr) > 20 {
			fmt.Println("🪙 Token String:", tokenStr[:20], "...")
		} else {
			fmt.Println("🪙 Token String:", tokenStr, "...")
		}
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			secret = "pharmeasy_super_secret_key_2024" // fallback
		}

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secret), nil
		})

		if err != nil {
			fmt.Println("❌ Token parse error:", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: " + err.Error()})
			c.Abort()
			return
		}

		if !token.Valid {
			fmt.Println("❌ Token not valid")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token not valid"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims"})
			c.Abort()
			return
		}

		userID := uint(claims["user_id"].(float64))
		fmt.Println("✅ Authenticated user_id:", userID)
		c.Set("user_id", userID)
		c.Next()
	}
}
