package common

import (
	"encoding/json"
	"errors"
	"net/http"
)

var (
	ErrUserNotFound      = errors.New("ERR-101:user not found, create user first")
	ErrUserNotAuthorised = errors.New("ERR-102:user not authorised")
	ErrJwtTokenInvalid   = errors.New("ERR-103:invalid JWT token")
)

func ProvideErrorResponse(w http.ResponseWriter, pError error, httpErrorCode int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(httpErrorCode)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": pError.Error(),
	})
}
