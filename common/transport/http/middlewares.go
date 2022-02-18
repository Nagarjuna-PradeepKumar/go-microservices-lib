package http

import (
	common "github.com/beezlabs-org/go-microservices-lib/common"
	casbin "github.com/beezlabs-org/go-microservices-lib/internal/authorization/casbin"
	"github.com/golang-jwt/jwt"
	"net/http"
	"os"
)

func MiddlewareForMicrosoftIdentityClientIdAndScopeVerification(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jwtMap := r.Context().Value("props").(jwt.MapClaims)
		if jwtMap["appid"] != os.Getenv("MICROSOFT_APP_ID") || jwtMap["scp"] != os.Getenv("MICROSOFT_APP_SCOPE") {
			common.ProvideErrorResponse(w, common.ErrJwtTokenInvalid, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func MiddlewareToAllowUserSignUp(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ifUserExists, err := common.CheckIfUserExists(r.Context(), common.GetLoggedUserIdFromJwtClaims(r.Context()))
		if err != nil {
			common.ProvideErrorResponse(w, err, http.StatusInternalServerError)
			return
		}
		if !ifUserExists && r.Method != http.MethodPost && r.URL.Path != "/users/signup" {
			common.ProvideErrorResponse(w, common.ErrUserNotFound, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func MiddlewareToCheckUserAuthorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enforcer := casbin.GetService().GetDefaultEnforcer()
		ok, err := enforcer.Enforce(common.GetLoggedUserIdFromJwtClaims(r.Context()), r.URL.Path, r.Method)
		if err != nil {
			common.ProvideErrorResponse(w, err, http.StatusUnauthorized)
			return
		}
		if !ok {
			common.ProvideErrorResponse(w, common.ErrUserNotAuthorised, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
