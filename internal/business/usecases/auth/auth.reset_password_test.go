package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tenSunFree/luma-lang-go/internal/apperror"
	"github.com/tenSunFree/luma-lang-go/internal/business/usecases/auth"
	"github.com/tenSunFree/luma-lang-go/internal/business/usecases/users"
)

func TestResetPassword(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		code        string
		newPassword string
		setup       func(f *fixture)
		wantErr     bool
		wantErrType apperror.ErrorType
	}{
		{
			name:        "happy path verifies code, updates password, deletes code",
			email:       "patrick@example.com",
			code:        "483921",
			newPassword: "Newpwd_999!",
			setup: func(f *fixture) {
				f.users.On("GetByEmail", mock.Anything, users.GetByEmailRequest{Email: "patrick@example.com"}).
					Return(users.GetByEmailResponse{User: activeUser(t)}, nil).Once()
				f.redis.On("Incr", mock.Anything, "pwd_reset_attempts:patrick@example.com").Return(int64(1), nil).Once()
				f.redis.On("Expire", mock.Anything, "pwd_reset_attempts:patrick@example.com", mock.AnythingOfType("time.Duration")).Return(nil).Once()
				f.redis.On("Get", mock.Anything, "pwd_reset_code:patrick@example.com").Return("483921", nil).Once()
				f.users.On("UpdatePassword", mock.Anything, mock.MatchedBy(func(req users.UpdatePasswordRequest) bool {
					u := req.User
					return u.ID == "user-1" && u.PasswordChangedAt != nil
				})).Return(nil).Once()
				f.redis.On("Del", mock.Anything, "pwd_reset_code:patrick@example.com").Return(nil).Once()
				f.redis.On("Del", mock.Anything, "pwd_reset_attempts:patrick@example.com").Return(nil).Once()
			},
		},
		{
			name:        "missing code returns BadRequest",
			email:       "patrick@example.com",
			code:        "",
			newPassword: "Newpwd_999!",
			setup:       func(f *fixture) {},
			wantErr:     true,
			wantErrType: apperror.ErrTypeBadRequest,
		},
		{
			name:        "empty new password returns BadRequest",
			email:       "patrick@example.com",
			code:        "483921",
			newPassword: "",
			setup:       func(f *fixture) {},
			wantErr:     true,
			wantErrType: apperror.ErrTypeBadRequest,
		},
		{
			name:        "unknown email returns Unauthorized",
			email:       "ghost@example.com",
			code:        "483921",
			newPassword: "Newpwd_999!",
			setup: func(f *fixture) {
				f.redis.On("Incr", mock.Anything, "pwd_reset_attempts:ghost@example.com").Return(int64(1), nil).Once()
				f.redis.On("Expire", mock.Anything, "pwd_reset_attempts:ghost@example.com", mock.AnythingOfType("time.Duration")).Return(nil).Once()
				f.users.On("GetByEmail", mock.Anything, users.GetByEmailRequest{Email: "ghost@example.com"}).
					Return(users.GetByEmailResponse{}, apperror.NotFound("email not found")).Once()
			},
			wantErr:     true,
			wantErrType: apperror.ErrTypeUnauthorized,
		},
		{
			name:        "wrong code returns Unauthorized and increments attempts",
			email:       "patrick@example.com",
			code:        "000000",
			newPassword: "Newpwd_999!",
			setup: func(f *fixture) {
				f.users.On("GetByEmail", mock.Anything, users.GetByEmailRequest{Email: "patrick@example.com"}).
					Return(users.GetByEmailResponse{User: activeUser(t)}, nil).Once()
				f.redis.On("Incr", mock.Anything, "pwd_reset_attempts:patrick@example.com").Return(int64(1), nil).Once()
				f.redis.On("Expire", mock.Anything, "pwd_reset_attempts:patrick@example.com", mock.AnythingOfType("time.Duration")).Return(nil).Once()
				f.redis.On("Get", mock.Anything, "pwd_reset_code:patrick@example.com").Return("483921", nil).Once()
			},
			wantErr:     true,
			wantErrType: apperror.ErrTypeUnauthorized,
		},
		{
			name:        "lockout after too many attempts returns Forbidden",
			email:       "patrick@example.com",
			code:        "483921",
			newPassword: "Newpwd_999!",
			setup: func(f *fixture) {
				f.redis.On("Incr", mock.Anything, "pwd_reset_attempts:patrick@example.com").Return(int64(6), nil).Once()
			},
			wantErr:     true,
			wantErrType: apperror.ErrTypeForbidden,
		},
		{
			name:        "expired or missing code returns Unauthorized",
			email:       "patrick@example.com",
			code:        "483921",
			newPassword: "Newpwd_999!",
			setup: func(f *fixture) {
				f.users.On("GetByEmail", mock.Anything, users.GetByEmailRequest{Email: "patrick@example.com"}).
					Return(users.GetByEmailResponse{User: activeUser(t)}, nil).Once()
				f.redis.On("Incr", mock.Anything, "pwd_reset_attempts:patrick@example.com").Return(int64(1), nil).Once()
				f.redis.On("Expire", mock.Anything, "pwd_reset_attempts:patrick@example.com", mock.AnythingOfType("time.Duration")).Return(nil).Once()
				f.redis.On("Get", mock.Anything, "pwd_reset_code:patrick@example.com").Return("", errors.New("redis: nil")).Once()
			},
			wantErr:     true,
			wantErrType: apperror.ErrTypeUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t)
			tt.setup(f)
			err := f.usecase.ResetPassword(context.Background(), auth.ResetPasswordRequest{
				Email:       tt.email,
				Code:        tt.code,
				NewPassword: tt.newPassword,
			})
			if !tt.wantErr {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			var domErr *apperror.DomainError
			require.True(t, errors.As(err, &domErr))
			assert.Equal(t, tt.wantErrType, domErr.Type)
		})
	}
}
