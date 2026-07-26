package middlewares

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tenSunFree/luma-lang-go/internal/config"
	"github.com/tenSunFree/luma-lang-go/internal/constants"
)

// SecurityHeadersMiddleware sets a small but high-leverage set of
// browser-side security headers on every response. The API itself
// doesn't render HTML, but credentialed XHR / fetch calls from a
// browser still benefit, and these headers are cheap insurance against
// future endpoints that *do* serve HTML (admin panels, email previews,
// etc.) where a missing header would be a real exposure.
//
//	X-Content-Type-Options: nosniff             — disables MIME sniffing
//	X-Frame-Options:        DENY                — blocks clickjacking via <iframe>
//	Referrer-Policy:        strict-origin-...   — caps the data leaked in Referer
//	Content-Security-Policy: default-src 'none' — APIs return JSON, never need to load anything
//	Strict-Transport-Security                    — production only, requires real HTTPS
func SecurityHeadersMiddleware() gin.HandlerFunc {
	isProduction := config.AppConfig.Environment == constants.EnvironmentProduction
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		if strings.HasPrefix(c.Request.URL.Path, "/swagger/") {
			h.Set("Content-Security-Policy",
				"default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self' data:; frame-ancestors 'none'")
		} else {
			h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		}
		h.Set("Permissions-Policy", "accelerometer=(), camera=(), geolocation=(), gyroscope=(), magnetometer=(), microphone=(), payment=(), usb=()")
		if isProduction {
			h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		c.Next()
	}
}
