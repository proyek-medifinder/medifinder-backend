package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPasswordResetLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.SetTrustedProxies(nil)
	router.POST("/forgot", PasswordResetLimit(2, time.Minute), func(c *gin.Context) { c.Status(200) })
	request := func(ip, forwarded string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/forgot", nil)
		req.RemoteAddr = ip + ":1234"
		req.Header.Set("X-Forwarded-For", forwarded)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		return response
	}
	for i := 0; i < 2; i++ {
		if request("192.0.2.1", "198.51.100.1").Code != 200 {
			t.Fatal("early rejection")
		}
	}
	rejected := request("192.0.2.1", "198.51.100.2")
	if rejected.Code != 429 || rejected.Header().Get("Retry-After") == "" {
		t.Fatal("spoofed header bypassed limit")
	}
	if request("192.0.2.2", "").Code != 200 {
		t.Fatal("different IP should have a separate allowance")
	}
}
