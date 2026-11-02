package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// Recheck the database so an existing JWT cannot bypass account deactivation.
func ActiveAdminAccount(check func(string) (bool, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetString("role") != "admin_apotek" {
			c.Next()
			return
		}
		active, err := check(c.GetString("user_id"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "Status akun belum dapat diperiksa."})
			return
		}
		if !active {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Akun admin dinonaktifkan. Hubungi Super Admin."})
			return
		}
		c.Next()
	}
}
