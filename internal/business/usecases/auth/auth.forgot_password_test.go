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

func TestForgotPassword(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		setup       func(f *fixture)
		wantErr     bool
		wantErrType apperror.ErrorType
	}{
		{
			name:  "happy path increments rate counter, persists code, queues email",
			email: "patrick@example.com",
			setup: func(f *fixture) {
				f.redis.On("Incr", mock.Anything, "forgot_attempts:patrick@example.com").Return(int64(1), nil).Once()
				f.redis.On("Expire", mock.Anything, "forgot_attempts:patrick@example.com", mock.AnythingOfType("time.Duration")).Return(nil).Once()
				f.users.On("GetByEmail", mock.Anything, users.GetByEmailRequest{Email: "patrick@example.com"}).Return(users.GetByEmailResponse{User: activeUser(t)}, nil).Once()
				// ForgotPassword will clear old code/attempts upon entry, immediately invalidating the old code.
				f.redis.On("Del", mock.Anything, "pwd_reset_code:patrick@example.com").Return(nil).Once()
				f.redis.On("Del", mock.Anything, "pwd_reset_attempts:patrick@example.com").Return(nil).Once()
				f.redis.On("Set", mock.Anything, "pwd_reset_code:patrick@example.com", mock.MatchedBy(func(code string) bool {
					if len(code) != 6 {
						return false
					}
					for _, c := range code {
						if c < '0' || c > '9' {
							return false
						}
					}
					return true
				})).Return(nil).Once()
				f.redis.On("Expire", mock.Anything, "pwd_reset_code:patrick@example.com", mock.AnythingOfType("time.Duration")).Return(nil).Once()
				f.mailer.On("SendPasswordReset", mock.AnythingOfType("string"), "patrick@example.com").Return(nil).Once()
			},
		},
		{
			// Defeat user enumeration: unknown email still consumes a counter slot (so attacker can't probe for free) but returns 200 OK.
			name:  "unknown email increments counter but is swallowed silently",
			email: "ghost@example.com",
			setup: func(f *fixture) {
				f.redis.On("Incr", mock.Anything, "forgot_attempts:ghost@example.com").Return(int64(1), nil).Once()
				f.redis.On("Expire", mock.Anything, "forgot_attempts:ghost@example.com", mock.AnythingOfType("time.Duration")).Return(nil).Once()
				f.users.On("GetByEmail", mock.Anything, users.GetByEmailRequest{Email: "ghost@example.com"}).
					Return(users.GetByEmailResponse{}, apperror.NotFound("email not found")).Once()
			},
		},
		{
			name:  "rate limit exceeded surfaces as Forbidden, no GetByEmail call",
			email: "victim@example.com",
			setup: func(f *fixture) {
				// Fixture caps at ForgotMaxAttempts=3; the 4th request trips it.
				f.redis.On("Incr", mock.Anything, "forgot_attempts:victim@example.com").Return(int64(4), nil).Once()
			},
			wantErr:     true,
			wantErrType: apperror.ErrTypeForbidden,
		},
		{
			name:  "infra error from users.GetByEmail bubbles up",
			email: "patrick@example.com",
			setup: func(f *fixture) {
				f.redis.On("Incr", mock.Anything, "forgot_attempts:patrick@example.com").Return(int64(1), nil).Once()
				f.redis.On("Expire", mock.Anything, "forgot_attempts:patrick@example.com", mock.AnythingOfType("time.Duration")).Return(nil).Once()
				f.users.On("GetByEmail", mock.Anything, users.GetByEmailRequest{Email: "patrick@example.com"}).
					Return(users.GetByEmailResponse{}, apperror.InternalCause(errors.New("redis down"))).Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t)
			tt.setup(f)
			err := f.usecase.ForgotPassword(context.Background(), auth.ForgotPasswordRequest{Email: tt.email})
			if !tt.wantErr {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			if tt.wantErrType != 0 {
				var domErr *apperror.DomainError
				require.True(t, errors.As(err, &domErr))
				assert.Equal(t, tt.wantErrType, domErr.Type)
			}
		})
	}
}
