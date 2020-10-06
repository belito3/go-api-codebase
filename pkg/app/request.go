package app

import (
	"github.com/belito3/go-api-codebase/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"strings"
)

const (
	prefix = "gin-go"
	// UserIDKey
	UserIDKey        = prefix + "/user-id"
	ReqBodyKey       = prefix + "/req-body"
	ResBodyKey       = prefix + "/res-body"
	LoggerReqBodyKey = prefix + "/logger-req-body"
)

func GetToken(c *gin.Context) string {
	var token string
	auth := c.GetHeader("Authorization")
	prefix := "Bearer "
	if auth != "" && strings.HasPrefix(auth, prefix) {
		token = auth[len(prefix):]
	}
	return token
}
func GetUserID(c *gin.Context) string {
	return c.GetString(UserIDKey)
}

// SetUserID
func SetUserID(c *gin.Context, userID string) {
	c.Set(UserIDKey, userID)
}
func ParseJSON(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		return errors.Wrap400Response(err)
	}
	return nil
}

// Parse Query
func ParseQuery(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindQuery(obj); err != nil {
		return errors.Wrap400Response(err)
	}
	return nil
}

// Parse Form
func ParseForm(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindWith(obj, binding.Form); err != nil {
		return errors.Wrap400Response(err)
	}
	return nil
}
