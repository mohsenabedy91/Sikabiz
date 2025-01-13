package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mohsenabedy91/Sikabiz/internal/adaper/http/presenter"
	"github.com/mohsenabedy91/Sikabiz/internal/adaper/http/request"
	"github.com/mohsenabedy91/Sikabiz/internal/core/port"
	"github.com/mohsenabedy91/Sikabiz/pkg/translation"
	"net/http"
)

type UserHandler struct {
	trans       translation.Translator
	userService port.UserService
	uowFactory  func() port.UserUnitOfWork
}

func NewUserHandler(
	trans translation.Translator,
	userService port.UserService,
	uowFactory func() port.UserUnitOfWork,
) *UserHandler {
	return &UserHandler{
		trans:       trans,
		userService: userService,
		uowFactory:  uowFactory,
	}
}

// Get godoc
// @x-kong {"service": "user-management-http-service"}
// @Security AuthBearer[READ_USER]
// @Summary Get User
// @Description Get User By ID
// @Tags User
// @Accept json
// @Produce json
// @Param language path string true "language 2 abbreviations" default(en)
// @Param userID path integer true "user id should be id"
// @Success 200 {object} presenter.Response{data=presenter.User} "Successful response"
// @Failure 400 {object} presenter.Error "Failed response"
// @Failure 401 {object} presenter.Error "Unauthorized"
// @Failure 422 {object} presenter.Response{validationErrors=[]presenter.ValidationError} "Validation error"
// @Failure 500 {object} presenter.Error "Internal server error"
// @ID get_language_v1_users_userID
// @Router /{language}/v1/users/{userID} [get]
func (r UserHandler) Get(ctx *gin.Context) {
	var userReq request.UserUUIDUri
	if err := ctx.ShouldBindUri(&userReq); err != nil {
		presenter.NewResponse(ctx, r.trans).Validation(err).Echo(http.StatusUnprocessableEntity)
		return
	}

	uowFactory := r.uowFactory()
	if err := uowFactory.BeginTx(ctx); err != nil {
		presenter.NewResponse(ctx, r.trans, StatusCodeMapping).Error(err).Echo()
		return
	}

	user, err := r.userService.GetByID(uowFactory, userReq.UUIDStr)
	if err != nil {
		if rErr := uowFactory.Rollback(); rErr != nil {
			presenter.NewResponse(ctx, r.trans, StatusCodeMapping).Error(rErr).Echo()
			return
		}
		presenter.NewResponse(ctx, r.trans, StatusCodeMapping).Error(err).Echo()
		return
	}

	if commitErr := uowFactory.Commit(); commitErr != nil {
		presenter.NewResponse(ctx, r.trans, StatusCodeMapping).Error(commitErr).Echo()
		return
	}

	presenter.NewResponse(ctx, r.trans).Payload(
		presenter.ToUserResource(user),
	).Echo()
}
