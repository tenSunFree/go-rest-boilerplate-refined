package auth

import (
	"context"
	"crypto/subtle"
	"fmt"
	"time"

	"github.com/tenSunFree/luma-lang-go/internal/apperror"
	"github.com/tenSunFree/luma-lang-go/internal/business/domain"
	"github.com/tenSunFree/luma-lang-go/internal/business/usecases/users"
	"github.com/tenSunFree/luma-lang-go/pkg/helpers"
	"github.com/tenSunFree/luma-lang-go/pkg/logger"
)

func (uc *usecase) ResetPassword(ctx context.Context, req ResetPasswordRequest) (err error) {
	const (
		usecaseName = "auth"
		funcName    = "ResetPassword"
		fileName    = "auth.reset_password.go"
	)
	startTime := time.Now()
	email := domain.NormalizeEmail(req.Email)
	code := req.Code
	newPassword := req.NewPassword

	logger.InfoWithContext(ctx, fmt.Sprintf("Upper %s", funcName), logger.Fields{
		"usecase": usecaseName, "method": funcName, "file": fileName,
		"request": logger.Fields{
			"email": helpers.MaskEmail(email), "has_code": code != "", "has_new_password": newPassword != "",
		},
	})
	defer func() {
		logger.InfoWithContext(ctx, fmt.Sprintf("Lower %s", funcName), logger.Fields{
			"usecase": usecaseName, "method": funcName, "file": fileName,
			"duration": time.Since(startTime).Milliseconds(),
		})
	}()

	if newPassword == "" {
		return apperror.BadRequest("new password is required")
	}
	if code == "" {
		return apperror.BadRequest("reset code is required")
	}

	maxAttempts := uc.cfg.PwdResetMaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 5
	}

	attemptsKey := pwdResetAttemptsKey(email)
	attempts, incrErr := uc.redisCache.Incr(ctx, attemptsKey)
	if incrErr != nil {
		logger.ErrorWithContext(ctx, "Reset password: failed to track attempts (non-fatal)", logger.Fields{
			"usecase": usecaseName, "method": funcName, "file": fileName,
			"step": "redis_incr_attempts", "error": incrErr.Error(), "email": helpers.MaskEmail(email),
		})
	} else if attempts == 1 {
		_ = uc.redisCache.Expire(ctx, attemptsKey, uc.cfg.PwdResetCodeTTL)
	}
	if attempts > int64(maxAttempts) {
		err = apperror.Forbidden("too many invalid attempts, please request a new code")
		logger.ErrorWithContext(ctx, "Reset password failed: lockout (max attempts exceeded)", logger.Fields{
			"usecase": usecaseName, "method": funcName, "file": fileName,
			"step": "check_lockout", "error": err.Error(), "email": helpers.MaskEmail(email), "attempts": attempts,
		})
		return err
	}

	lookupResp, lookupErr := uc.users.GetByEmail(ctx, users.GetByEmailRequest{Email: email})
	if lookupErr != nil {
		err = apperror.Unauthorized("reset code is invalid or expired")
		logger.ErrorWithContext(ctx, "Reset password failed: user lookup error", logger.Fields{
			"usecase": usecaseName, "method": funcName, "file": fileName,
			"step": "get_user_by_email", "error": lookupErr.Error(), "email": helpers.MaskEmail(email),
		})
		return err
	}
	user := lookupResp.User

	codeKey := pwdResetCodeKey(email)
	storedCode, getErr := uc.redisCache.Get(ctx, codeKey)
	if getErr != nil || storedCode == "" {
		err = apperror.Unauthorized("reset code is invalid or expired")
		logger.ErrorWithContext(ctx, "Reset password failed: code expired or not found", logger.Fields{
			"usecase": usecaseName, "method": funcName, "file": fileName,
			"step": "redis_get_reset_code", "error": err.Error(), "email": helpers.MaskEmail(email),
		})
		return err
	}

	if subtle.ConstantTimeCompare([]byte(storedCode), []byte(code)) != 1 {
		err = apperror.Unauthorized("reset code is invalid or expired")
		logger.ErrorWithContext(ctx, "Reset password failed: invalid code", logger.Fields{
			"usecase": usecaseName, "method": funcName, "file": fileName,
			"step": "compare_reset_code", "error": err.Error(), "email": helpers.MaskEmail(email),
		})
		return err
	}

	if delErr := uc.redisCache.Del(ctx, codeKey); delErr != nil {
		logger.ErrorWithContext(ctx, "Reset password: failed to delete used code (non-fatal, may allow replay until TTL expiry)", logger.Fields{
			"usecase": usecaseName, "method": funcName, "file": fileName,
			"step": "redis_del_reset_code", "error": delErr.Error(), "email": helpers.MaskEmail(email),
		})
	}
	if delErr := uc.redisCache.Del(ctx, attemptsKey); delErr != nil {
		logger.ErrorWithContext(ctx, "Reset password: failed to delete attempts counter (non-fatal)", logger.Fields{
			"usecase": usecaseName, "method": funcName, "file": fileName,
			"step": "redis_del_attempts", "error": delErr.Error(), "email": helpers.MaskEmail(email),
		})
	}

	if hashErr := user.ChangePassword(newPassword, uc.cfg.BcryptCost); hashErr != nil {
		err = apperror.InternalCause(fmt.Errorf("hash reset password: %w", hashErr))
		logger.ErrorWithContext(ctx, "Reset password failed: hash error", logger.Fields{
			"usecase": usecaseName, "method": funcName, "file": fileName,
			"step": "hash_password", "error": hashErr.Error(), "user_id": user.ID,
		})
		return err
	}
	if updateErr := uc.users.UpdatePassword(ctx, users.UpdatePasswordRequest{User: &user}); updateErr != nil {
		err = updateErr
		logger.ErrorWithContext(ctx, "Reset password failed: update error", logger.Fields{
			"usecase": usecaseName, "method": funcName, "file": fileName,
			"step": "users_update_password", "error": updateErr.Error(), "user_id": user.ID,
		})
		return err
	}
	return nil
}
