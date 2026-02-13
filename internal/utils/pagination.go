package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetPagination(c *gin.Context) (limit, offset int) {

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, _ := strconv.Atoi(pageStr)
	l, _ := strconv.Atoi(limitStr)

	if page < 1 {
		page = 1
	}

	if l < 1 {
		l = 10
	}

	if l > 50 {
		l = 50
	}

	limit = l
	offset = (page - 1) * limit

	return
}

func GetPaginationAdvanced(c *gin.Context) (page, limit, offset int) {

	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ = strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	offset = (page - 1) * limit
	return
}
