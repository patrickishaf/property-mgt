package net

import (
	"strings"
	"sync"
	"time"

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

// perMinute is a limiter rate of n events a minute (rate.Limit counts per second).
func perMinute(n int) rate.Limit {
	return rate.Every(time.Minute / time.Duration(n))
}

func getLimiter(key string, requestsPerMinute, burstRequestLimit int) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()
	l, exists := limiters[key]
	if !exists {
		l = rate.NewLimiter(perMinute(requestsPerMinute), burstRequestLimit)
		limiters[key] = l
	}
	return l
}

func RateLimit(requestsPerMinute, burstRequestLimit int) gin.HandlerFunc {
	limiter := rate.NewLimiter(perMinute(requestsPerMinute), burstRequestLimit)
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

// RateLimitByIPAddress limits each client IP to requestsPerMinute (with bursts
// of up to burstRequestLimit) on the routes it guards. name keeps separate
// budgets per route group, so login attempts don't use up password resets.
func RateLimitByIPAddress(name string, requestsPerMinute, burstRequestLimit int) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := name + ":" + c.ClientIP()
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
		limiter := getLimiter("user:"+username, requestsPerMinute, burstRequestLimit)

		if limiter.Allow() {
			c.Next()
		} else {
			c.IndentedJSON(429, ErrorResponse(common.ErrRequestRateLimited, nil))
			c.Abort()
			return
		}
	}
}

func VerifyJWT(jwtKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.IndentedJSON(401, ErrorResponse(common.ErrInvalidRequest, nil))
			c.Abort()
			return
		}
		tokens := strings.Split(authHeader, " ")
		if len(tokens) != 2 || tokens[0] != "Bearer" {
			c.IndentedJSON(401, ErrorResponse(common.ErrInvalidRequest, nil))
			c.Abort()
			return
		}
		actor, err := common.ExtractActorFromJWT(tokens[1], []byte(jwtKey))
		if err != nil {
			c.IndentedJSON(401, ErrorResponse(common.ErrInvalidToken, nil))
			c.Abort()
			return
		}
		c.Set(common.CTX_KEY_ACTOR, actor)
		c.Next()
	}
}
