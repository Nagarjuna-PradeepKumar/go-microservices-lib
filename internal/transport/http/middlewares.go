package http

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"github.com/beezlabs-org/go-microservices-lib/internal"
	"github.com/beezlabs-org/go-microservices-lib/internal/cache"
	"github.com/golang-jwt/jwt"
	"github.com/lestrrat-go/jwx/jwk"
	"net/http"
	"strings"
)

func GenericMiddlewareToSetHTTPHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		next.ServeHTTP(w, r)
	})
}

func GenericMiddlewareToUpdateEndpointContextForCacheProcessing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var endpointContext cache.RedisEndpointContext
		endpointContext.NotFromCache = r.Header.Get("X-No-Cache") == "true"
		endpointContext.Key = r.URL.Path
		endpointContext.Cacheable = r.Method == http.MethodGet
		ctx := context.WithValue(r.Context(), "cacheable-endpoint-context", endpointContext)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func JwtMiddlewareForMicrosoftIdentity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jwtToken, err := getJWTTokenFromHTTPHeader(r)
		if err != nil {
			internal.ProvideErrorResponse(w, err, http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return getJWKPublicKeyForMicrosoftIdentity(r.Context(), token)
		})
		if err != nil {
			err2 := errors.New(internal.ErrJwtTokenInvalid.Error() + ": " + err.Error())
			internal.ProvideErrorResponse(w, err2, http.StatusUnauthorized)
			return
		}
		//Check if token is valid
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			ctx := context.WithValue(r.Context(), "props", claims)
			//The further verification like client id or scope verification will be handled by the next middleware -
			// MiddlewareForMicrosoftIdentityClientIdAndScopeVerification
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			internal.ProvideErrorResponse(w, internal.ErrEmptyAuthHeader, http.StatusUnauthorized)
			return
		}
	})
}

func getJWTTokenFromHTTPHeader(r *http.Request) (string, error) {
	auth := r.Header.Get("Authorization")
	if len(auth) == 0 {
		return "", internal.ErrEmptyAuthHeader
	}
	authHeader := strings.Split(auth, "Bearer ")
	if len(authHeader) != 2 {
		return "", internal.ErrMalformedToken
	}
	jwtToken := authHeader[1]
	return jwtToken, nil
}

func getJWKPublicKeyForMicrosoftIdentity(ctx context.Context, token *jwt.Token) (interface{}, error) {
	//Get the latest Public JWK keys from Microsoft Identity Platform
	jwtKeyURl, err := getJwkKeyURL(token)
	if err != nil {
		return nil, err
	}
	keySet, err := jwk.Fetch(ctx, jwtKeyURl)
	if err != nil {
		return nil, err
	}
	//Check if kid exists in the token
	kid, ok := token.Header["kid"].(string)
	if !ok {
		return nil, internal.ErrKIDNotFound
	}
	//Get the keys based on kid
	keys, ok := keySet.LookupKeyID(kid)
	if !ok {
		return nil, fmt.Errorf("jwk key %v not found", kid)
	}
	//Get the public key from the key
	publicKey := &rsa.PublicKey{}
	err = keys.Raw(publicKey)
	if err != nil {
		return nil, internal.ErrUnableToParsePublicKey
	}
	return publicKey, nil
}

func getJwkKeyURL(token *jwt.Token) (string, error) {
	tokenVersion := token.Claims.(jwt.MapClaims)["ver"].(string)
	if tokenVersion == "1.0" {
		return "https://login.microsoftonline.com/common/discovery/keys", nil
	} else if tokenVersion == "2.0" {
		return "https://login.microsoftonline.com/common/discovery/v2.0/keys", nil
	} else {
		return "", internal.ErrUnexpectedTokenVersion
	}
}
