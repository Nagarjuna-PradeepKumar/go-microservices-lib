package http

import (
	internal "github.com/beezlabs-org/go-microservices-lib/internal/transport/http"
	kittransport "github.com/go-kit/kit/transport"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/go-kit/log"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func InitMuxRouter(logger log.Logger, options []kithttp.ServerOption) (*mux.Router, []kithttp.ServerOption) {
	r := mux.NewRouter()
	errorEncoder := kithttp.ServerErrorEncoder(EncodeErrorResponse)
	errorHandler := kithttp.ServerErrorHandler(kittransport.NewLogErrorHandler(logger))
	options = append(options, errorEncoder, errorHandler)

	r.Use(internal.GenericMiddlewareToSetHTTPHeader)
	r.Use(internal.GenericMiddlewareToUpdateEndpointContextForCacheProcessing)
	r.Use(internal.JwtMiddlewareForMicrosoftIdentity)
	r.Use(MiddlewareForMicrosoftIdentityClientIdAndScopeVerification)
	r.Use(MiddlewareToAllowUserSignUp)
	r.Use(MiddlewareToCheckUserAuthorization)

	r.Methods("GET").Path("/metrics").Handler(promhttp.Handler())
	return r, options
}
