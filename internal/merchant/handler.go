package merchant

import (
	"binanga/internal/account"
	"binanga/internal/config"
	"binanga/internal/database"
	merchantDB "binanga/internal/merchant/database"
	"binanga/internal/merchant/model"
	"binanga/internal/middleware"
	"binanga/internal/middleware/handler"
	"binanga/pkg/logging"
	"binanga/pkg/validate"
	"net/http"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func NewHandler(merchantDB merchantDB.MerchantDB) *Handler {
	return &Handler{
		merchantDB: merchantDB,
	}
}

type Handler struct {
	merchantDB merchantDB.MerchantDB
}

// saveMerchant handles POST /v1/api/merchants
func (h *Handler) saveArticle(c *gin.Context) {
	handler.HandleRequest(c, func(c *gin.Context) *handler.Response {
		logger := logging.FromContext(c)
		// bind
		type RequestBody struct {
			Merchant struct {
				Name string `json:"name" binding:"required,min=5"`
			} `json:"merchant"`
		}
		var body RequestBody
		if err := c.ShouldBindJSON(&body); err != nil {
			logger.Errorw("merchant.handler.register failed to bind", "err", err)
			var details []*validate.ValidationErrDetail
			if vErrs, ok := err.(validator.ValidationErrors); ok {
				details = validate.ValidationErrorDetails(&body.Merchant, "json", vErrs)
			}
			return handler.NewErrorResponse(http.StatusBadRequest, handler.InvalidBodyValue, "invalid merchant request in body", details)
		}

		// save merchant
		currentUser := account.MustCurrentUser(c)

		merchant := model.Merchant{
			Name:   body.Merchant.Name,
			User:   *currentUser,
			UserID: currentUser.ID,
		}
		err := h.merchantDB.SaveMerchant(c.Request.Context(), &merchant)
		if err != nil {
			if database.IsKeyConflictErr(err) {
				return handler.NewErrorResponse(http.StatusConflict, handler.DuplicateEntry, "duplicate merchant title", nil)
			}
			return handler.NewInternalErrorResponse(err)
		}
		return handler.NewSuccessResponse(http.StatusCreated, NewMerchantResponse(&merchant))
	})
}

// getMerchant handles POST /v1/api/merchants/:id
func (h *Handler) getMerchant(c *gin.Context) {
	handler.HandleRequest(c, func(c *gin.Context) *handler.Response {
		logger := logging.FromContext(c)
		// bind
		type RequestUri struct {
			Slug string `uri:"id"`
		}
		var uri RequestUri
		if err := c.ShouldBindUri(&uri); err != nil {
			logger.Errorw("merchant.handler.getMerchant failed to bind", "err", err)
			var details []*validate.ValidationErrDetail
			if vErrs, ok := err.(validator.ValidationErrors); ok {
				details = validate.ValidationErrorDetails(&uri, "uri", vErrs)
			}
			return handler.NewErrorResponse(http.StatusBadRequest, handler.InvalidUriValue, "invalid merchant request in uri", details)
		}

		// find
		merchant, err := h.merchantDB.FindMerchantById(c.Request.Context(), uri.Slug)
		if err != nil {
			if database.IsRecordNotFoundErr(err) {
				return handler.NewErrorResponse(http.StatusNotFound, handler.NotFoundEntity, "not found merchant", nil)
			}
			return handler.NewInternalErrorResponse(err)
		}
		return handler.NewSuccessResponse(http.StatusOK, NewMerchantResponse(merchant))
	})
}

func RouteV1(cfg *config.Config, h *Handler, r *gin.Engine, auth *jwt.GinJWTMiddleware) {
	v1 := r.Group("v1/api")
	v1.Use(middleware.RequestIDMiddleware(), middleware.TimeoutMiddleware(cfg.ServerConfig.WriteTimeout))

	merchantV1 := v1.Group("merchant")

	// auth required
	merchantV1.Use(auth.MiddlewareFunc())
	{
		merchantV1.POST("", h.saveArticle)
		merchantV1.GET(":id", h.saveArticle)

	}
}
