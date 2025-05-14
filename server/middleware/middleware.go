package middleware

import (
	"net/http"
	"todorist/config"
	"todorist/pkg/jwttoken"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(db *config.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.Request.Header.Get("Authorization")
		if header != "" {
			claims, err := jwttoken.ValidateToken(header)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid Token"})
				return
			} else {
				db.SetUserId(claims.UserId)
				c.Set("userId", claims.UserId)
			}
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
			return
		}
		c.Next()
	}
}
