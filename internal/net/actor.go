package net

import (
	"github.com/gin-gonic/gin"
	"github.com/patrickishaf/property-mgt/common"
)

// GetActor returns the actor that VerifyJWT attached to the request context.
func GetActor(c *gin.Context) (*common.Actor, bool) {
	value, exists := c.Get(common.CTX_KEY_ACTOR)
	if !exists {
		return nil, false
	}
	actor, isValid := value.(*common.Actor)
	if !isValid || actor == nil {
		return nil, false
	}
	return actor, true
}
