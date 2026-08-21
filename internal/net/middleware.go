package net

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/patrickishaf/property-mgt/common"
	"golang.org/x/time/rate"
)

type Middleware func() gin.HandlerFunc

var (
	limiters = make(map[string]*rate.Limiter)
	mu       sync.Mutex
)

func AttachRequestID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestID := ctx.GetHeader(common.REQUEST_ID_KEY)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		ctx.Set(common.REQUEST_ID_KEY, requestID)
		ctx.Next()
	}
}

func getLimiter(key string, requestsPerMinute, burstRequestLimit int) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()
	l, exists := limiters[key]
	if !exists {
		l := rate.NewLimiter(rate.Limit(requestsPerMinute), burstRequestLimit)
		limiters[key] = l
	}
	return l
}

func RateLimit(requestsPerMinute, burstRequestLimit int) gin.HandlerFunc {
	limiter := rate.NewLimiter(rate.Limit(requestsPerMinute), burstRequestLimit)
	return func(c *gin.Context) {
		if limiter.Allow() {
			c.Next()
		} else {
			c.IndentedJSON(429, ErrorResponse(common.ErrRequestRateLimited, nil))
			c.Abort()
			return
		}
	}
}

func RateLimitByIPAddress(requestsPerMinute, burstRequestLimit int) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		limiter := getLimiter(key, requestsPerMinute, burstRequestLimit)

		if limiter.Allow() {
			c.Next()
		} else {
			c.IndentedJSON(429, ErrorResponse(common.ErrRequestRateLimited, nil))
			c.Abort()
			return
		}
	}
}

func RateLimitByUsername(requestsPerMinute, burstRequestLimit int) gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.GetString(common.REQUEST_USER_UNAME_KEY)
		limiter := getLimiter(username, requestsPerMinute, burstRequestLimit)

		if limiter.Allow() {
			c.Next()
		} else {
			c.IndentedJSON(429, ErrorResponse(common.ErrRequestRateLimited, nil))
			c.Abort()
			return
		}
	}
}
