package response

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Ciobi0212/httpfromtcp/request"
)

type Middleware func(next Handler) Handler

// Pre-defined middleware for common use

type CorsOptions struct {
	AllowedOrigins   []string // e.g., []string{"https://mydomain.com", "http://localhost:3000"}
	AllowAllOrigins  bool     // If true, AllowedOrigins is ignored and '*' is used.
	AllowedMethods   []string // e.g., []string{"GET", "POST", "PUT", "DELETE"}
	AllowedHeaders   []string // e.g., []string{"Content-Type", "Authorization"}
	AllowCredentials bool
	MaxAge           int // In seconds, for Access-Control-Max-Age
}

func NewCORSMiddleware(options CorsOptions) Middleware {
	return func(next Handler) Handler {
		return func(res ResponseWriter, req *request.Request) *HandlerError {
			// Pre-flight request
			if req.RequestLine.Method == string(OPTIONS) {
				requestOrigin := req.Headers.Get("Origin")

				allowOriginValue := ""
				if options.AllowAllOrigins {
					allowOriginValue = "*"
				} else {
					for _, allowed := range options.AllowedOrigins {
						if allowed == requestOrigin {
							allowOriginValue = allowed
							break
						}
					}
				}

				if allowOriginValue != "" {
					fmt.Println("hey ")
					res.Headers.Add("Access-Control-Allow-Origin", allowOriginValue)

					if options.AllowCredentials && allowOriginValue != "*" {
						res.Headers.Add("Access-Control-Allow-Credentials", "true")
					}

					// Join slices into comma-separated strings for headers
					if len(options.AllowedMethods) > 0 {
						res.Headers.Add("Access-Control-Allow-Methods", strings.Join(options.AllowedMethods, ", "))
					}
					if len(options.AllowedHeaders) > 0 {
						res.Headers.Add("Access-Control-Allow-Headers", strings.Join(options.AllowedHeaders, ", "))
					}
					if options.MaxAge > 0 {
						res.Headers.Add("Access-Control-Max-Age", strconv.Itoa(options.MaxAge))
					}
				}
				res.WriteHeaders(NoContent)
				return nil
			}

			// Actual request logic (e.g. POST, PUT)
			if options.AllowAllOrigins {
				res.Headers.Add("Access-Control-Allow-Origin", "*")
			} else if requestOrigin := req.Headers.Get("Origin"); requestOrigin != "" {
				for _, allowed := range options.AllowedOrigins {
					if allowed == requestOrigin {
						res.Headers.Add("Access-Control-Allow-Origin", requestOrigin)
						if options.AllowCredentials {
							res.Headers.Add("Access-Control-Allow-Credentials", "true")
						}
						break
					}
				}
			}

			return next(res, req)
		}
	}
}
