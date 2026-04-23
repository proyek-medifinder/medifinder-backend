package middleware

import (
	"log"
	"time"

	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
)

var (
	green  = color.New(color.FgGreen).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
	cyan   = color.New(color.FgCyan).SprintFunc()
)

func statusColor(status int) func(a ...interface{}) string {
	switch {
	case status >= 200 && status < 300:
		return green
	case status >= 400 && status < 500:
		return yellow
	default:
		return red
	}
}

func statusEmoji(status int) string {
	switch {
	case status >= 200 && status < 300:
		return "🚀"
	case status >= 400 && status < 500:
		return "⚠️"
	default:
		return "💥"
	}
}

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		status := c.Writer.Status()
		latency := time.Since(start)

		colorFunc := statusColor(status)
		emoji := statusEmoji(status)

		log.Printf("%s %-6s %-30s %s %10v",
			emoji,
			cyan(method),
			path,
			colorFunc(status),
			latency,
		)
	}
}
