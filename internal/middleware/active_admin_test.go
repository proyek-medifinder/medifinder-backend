package middleware

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"testing"
)

func TestDeactivatedAdminLosesExistingSessionAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	active := true
	failed := false
	role := "admin_apotek"
	checks := 0
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set("role", role); c.Set("user_id", "same-session-user"); c.Next() })
	router.Use(ActiveAdminAccount(func(id string) (bool, error) {
		checks++
		if id != "same-session-user" {
			t.Fatal("wrong session identity")
		}
		if failed {
			return false, errors.New("database unavailable")
		}
		return active, nil
	}))
	router.GET("/admin", func(c *gin.Context) { c.Status(200) })
	request := func() int {
		out := httptest.NewRecorder()
		router.ServeHTTP(out, httptest.NewRequest("GET", "/admin", nil))
		return out.Code
	}
	if request() != 200 {
		t.Fatal("active admin blocked")
	}
	active = false
	if request() != 403 {
		t.Fatal("deactivated account retained access with its existing session")
	}
	active = true
	if request() != 200 {
		t.Fatal("reactivated account remains blocked")
	}
	failed = true
	if request() != 503 {
		t.Fatal("database error must fail closed")
	}
	role = "super_admin"
	before := checks
	if request() != 200 || checks != before {
		t.Fatal("admin-only check affected Super Admin")
	}
}
