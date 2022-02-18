package internal

import (
	"encoding/json"
	"errors"
	"net/http"
)

var (
	ErrEmptyAuthHeader        = errors.New("ERR-001:authorization header is empty")
	ErrMalformedToken         = errors.New("ERR-002:malformed Token")
	ErrKIDNotFound            = errors.New("ERR-003:kid header not found")
	ErrUnableToParsePublicKey = errors.New("ERR-004:could not parse public key")
	ErrUnexpectedTokenVersion = errors.New("ERR-005:unexpected token version")
	ErrJwtTokenInvalid        = errors.New("ERR-006:invalid JWT token")
)

func ProvideErrorResponse(w http.ResponseWriter, pError error, httpErrorCode int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(httpErrorCode)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": pError.Error(),
	})
}
