package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	authuc "github.com/tenSunFree/luma-lang-go/internal/business/usecases/auth"
	"github.com/tenSunFree/luma-lang-go/internal/http/datatransfers/requests"
	v1 "github.com/tenSunFree/luma-lang-go/internal/http/handlers/v1"
	"github.com/tenSunFree/luma-lang-go/pkg/audit"
	"github.com/tenSunFree/luma-lang-go/pkg/logger"
	"github.com/tenSunFree/luma-lang-go/pkg/validators"
)

// ForgotPassword godoc
// @Summary      Issue a password-reset code
// @Description  Emails a six-digit reset code to the address. Always returns 200 even if the email isn't registered, to defeat user enumeration.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      requests.ForgotPasswordRequest  true  "Email address"
// @Success      200  {object}  v1.BaseResponse  "Reset code queued (or email not registered — same response either way)"
// @Failure      400  {object}  v1.BaseResponse  "Malformed JSON body"
// @Failure      422  {object}  v1.BaseResponse  "Validation error"
// @Router       /auth/password/forgot [post]
func (h Handler) ForgotPassword(ctx *gin.Context) {
	const (
		controllerName = "auth"
		funcName       = "ForgotPassword"
		fileName       = "auth.forgot_password.go"
	)
	var req requests.ForgotPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.WarnWithContext(ctx.Request.Context(), "ForgotPassword: invalid request body", logger.Fields{
			"controller": controllerName,
			"method":     funcName,
			"file":       fileName,
			"error":      err.Error(),
		})
		v1.NewErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validators.ValidatePayloads(req); err != nil {
		logger.WarnWithContext(ctx.Request.Context(), "ForgotPassword: validation error", logger.Fields{
			"controller": controllerName,
			"method":     funcName,
			"file":       fileName,
			"error":      err.Error(),
			"request": logger.Fields{
				"email": req.Email,
			},
		})
		v1.RespondWithError(ctx, err)
		return
	}

	if err := h.usecase.ForgotPassword(ctx.Request.Context(), authuc.ForgotPasswordRequest{Email: req.Email}); err != nil {
		ev := auditFromGin(ctx)
		ev.Type = audit.EventPasswordForgotFail
		ev.Email = req.Email
		ev.Reason = err.Error()
		audit.Record(ev)
		logger.ErrorWithContext(ctx.Request.Context(), "ForgotPassword failed in controller", logger.Fields{
			"controller": controllerName,
			"method":     funcName,
			"file":       fileName,
			"error":      err.Error(),
			"email":      req.Email,
		})
		v1.RespondWithError(ctx, err)
		return
	}

	ev := auditFromGin(ctx)
	ev.Type = audit.EventPasswordForgotOK
	ev.Success = true
	ev.Email = req.Email
	audit.Record(ev)

	v1.NewSuccessResponse(ctx, http.StatusOK, "if the email is registered, a reset code has been sent", nil)
}
