package service

import (
	"fmt"
	"github.com/belito3/go-web-api/app/repository/impl"
	"github.com/belito3/go-web-api/pkg/logger"
	"github.com/gin-gonic/gin"
	"net/http"
)

type AccountService struct {
	store 	impl.IStore
}

func NewAccountService(store impl.IStore) *AccountService {
	return &AccountService{store: store}
}

func (s *AccountService) Add(c *gin.Context) {
	// Add account
	ctx := c.Request.Context()
	var arg impl.CreateAccountParams
	if err := c.ShouldBindJSON(&arg); err != nil {
		logger.Errorf(ctx, "Invalid input parameter")
		ResponseError(c, http.StatusBadRequest, fmt.Sprintf("Invalid input parameter: %v", err))
		return
	}
	account, err := s.store.CreateAccount(ctx, arg)
	if err != nil {
		logger.Errorf(ctx, "Add account error %v\n", err)
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}
	r := map[string]interface{}{
		"account": account,
	}
	ResponseSuccess(c, http.StatusOK, r)
}