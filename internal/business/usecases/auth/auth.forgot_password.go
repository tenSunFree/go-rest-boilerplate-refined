package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/tenSunFree/luma-lang-go/internal/apperror"
	"github.com/tenSunFree/luma-lang-go/internal/business/domain"
	"github.com/tenSunFree/luma-lang-go/internal/business/usecases/users"
	"github.com/tenSunFree/luma-lang-go/pkg/helpers"
	"github.com/tenSunFree/luma-lang-go/pkg/logger"
)

// pwdResetCodeKey / pwdResetAttemptsKey are namespaced separately
// from user_otp:<email> (registration) so the two flows can never
// collide or be replayed against each other.
func pwdResetCodeKey(email string) string     { return fmt.Sprintf("pwd_reset_code:%s", email) }
func pwdResetAttemptsKey(email string) string { return fmt.Sprintf("pwd_reset_attempts:%s", email) }

// ForgotPassword issues a 6-digit code, persists it in Redis with
// TTL, and emails it to the user. To defeat email enumeration the
// response is identical whether the email exists or not.
func (uc *usecase) ForgotPassword(ctx context.Context, req ForgotPasswordRequest) (err error) {
	const (
		usecaseName = "auth"
		funcName    = "ForgotPassword"
		fileName    = "auth.forgot_password.go"
	)
	startTime := time.Now()
	email := domain.NormalizeEmail(req.Email)

	logger.InfoWithContext(ctx, fmt.Sprintf("Upper %s", funcName), logger.Fields{
		"usecase": usecaseName, "method": funcName, "file": fileName,
		"request": logger.Fields{"email": helpers.MaskEmail(email)},
	})
	defer func() {
		logger.InfoWithContext(ctx, fmt.Sprintf("Lower %s", funcName), logger.Fields{
			"usecase": usecaseName, "method": funcName, "file": fileName,
			"duration": time.Since(startTime).Milliseconds(),
		})
	}()

	if uc.cfg.ForgotMaxAttempts > 0 {
		key := forgotAttemptsKey(email)
		attempts, incrErr := uc.redisCache.Incr(ctx, key)
		if incrErr != nil {
			logger.ErrorWithContext(ctx, "ForgotPassword: failed to track attempts (non-fatal)", logger.Fields{
				"usecase": usecaseName, "method": funcName, "file": fileName,
				"step": "redis_incr_attempts", "error": incrErr.Error(), "email": helpers.MaskEmail(email),
			})
		} else if attempts == 1 {
			_ = uc.redisCache.Expire(ctx, key, uc.cfg.ForgotLockoutTTL)
		}
		if attempts > int64(uc.cfg.ForgotMaxAttempts) {
			err = apperror.Forbidden("too many password reset requests, please try again later")
			logger.ErrorWithContext(ctx, "ForgotPassword failed: rate limit exceeded", logger.Fields{
				"usecase": usecaseName, "method": funcName, "file": fileName,
				"step": "check_rate_limit", "error": err.Error(), "email": helpers.MaskEmail(email), "attempts": attempts,
			})
			return err
		}
	}

	lookupResp, lookupErr := uc.users.GetByEmail(ctx, users.GetByEmailRequest{Email: email})
	if lookupErr != nil {
		var domErr *apperror.DomainError
		if errors.As(lookupErr, &domErr) && domErr.Type == apperror.ErrTypeNotFound {
			return nil
		}
		err = lookupErr
		logger.ErrorWithContext(ctx, "Forgot password failed: user lookup error", logger.Fields{
			"usecase": usecaseName, "method": funcName, "file": fileName,
			"step": "get_user_by_email", "error": lookupErr.Error(), "email": helpers.MaskEmail(email),
		})
		return err
	}

	if !lookupResp.User.Active {
		return nil
	}

	code, otpErr := helpers.GenerateOTPCode(6)
	if otpErr != nil {
		err = apperror.InternalCause(fmt.Errorf("generate reset code: %w", otpErr))
		logger.ErrorWithContext(ctx, "Forgot password failed: code generation error", logger.Fields{
			"usecase": usecaseName, "method": funcName, "file": fileName,
			"step": "generate_reset_code", "error": otpErr.Error(), "email": helpers.MaskEmail(email),
		})
		return err
	}

	codeKey := pwdResetCodeKey(email)
	attemptsKey := pwdResetAttemptsKey(email)

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

	if setErr := uc.redisCache.Set(ctx, codeKey, code); setErr != nil {
		err = apperror.InternalCause(fmt.Errorf("persist reset code: %w", setErr))
		logger.ErrorWithContext(ctx, "Forgot password failed: persist code error", logger.Fields{
			"usecase": usecaseName, "method": funcName, "file": fileName,
			"step": "redis_set_reset_code", "error": setErr.Error(), "email": helpers.MaskEmail(email),
		})
		return err
	}
	if expireErr := uc.redisCache.Expire(ctx, codeKey, uc.cfg.PwdResetCodeTTL); expireErr != nil {
		logger.ErrorWithContext(ctx, "Forgot password: failed to set TTL on reset code (non-fatal)", logger.Fields{
			"usecase": usecaseName, "method": funcName, "file": fileName,
			"step": "redis_expire_reset_code", "error": expireErr.Error(),
		})
	}

	if mailErr := uc.mailer.SendPasswordReset(code, email); mailErr != nil {
		_ = uc.redisCache.Del(ctx, codeKey)
		err = apperror.InternalCause(fmt.Errorf("send reset email: %w", mailErr))
		logger.ErrorWithContext(ctx, "Forgot password failed: mailer error", logger.Fields{
			"usecase": usecaseName, "method": funcName, "file": fileName,
			"step": "mailer_send_password_reset", "error": mailErr.Error(), "email": helpers.MaskEmail(email),
		})
		return err
	}
	return nil
}
