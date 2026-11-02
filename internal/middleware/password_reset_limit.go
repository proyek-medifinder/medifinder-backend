package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// Request limits are per application instance; account cooldown lives in PostgreSQL.
func PasswordResetLimit(limit int, window time.Duration) gin.HandlerFunc {
	type entry struct {
		count int
		until time.Time
	}
	var mu sync.Mutex
	entries := map[string]entry{}
	return func(c *gin.Context) {
		now := time.Now()
		key := c.ClientIP()
		mu.Lock()
		for ip, item := range entries {
			if !now.Before(item.until) {
				delete(entries, ip)
			}
		}
		item, exists := entries[key]
		if !exists {
			item.until = now.Add(window)
		}
		allowed := item.count < limit && (exists || len(entries) < 4096)
		if allowed {
			item.count++
			entries[key] = item
		}
		mu.Unlock()
		if !allowed {
			retry := int(time.Until(item.until).Seconds()) + 1
			if retry < 1 {
				retry = int(window.Seconds()) + 1
			}
			c.Header("Retry-After", strconv.Itoa(retry))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"message": "Terlalu banyak permintaan. Tunggu beberapa menit sebelum mencoba lagi."})
			return
		}
		c.Header("Cache-Control", "no-store")
		c.Next()
	}
}
