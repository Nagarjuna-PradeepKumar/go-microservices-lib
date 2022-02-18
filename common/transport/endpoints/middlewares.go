package endpoints

import (
	"context"
	user "github.com/beezlabs-org/go-microservices-lib/common"
	"github.com/beezlabs-org/go-microservices-lib/internal/authorization/casbin"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/log"
)

func AuthorizationMiddleware(endpointName string, logger log.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			_ = logger.Log("msg", "AuthorizationMiddleware", "endpoint", endpointName)
			ok, err := casbin.GetService().GetDefaultEnforcer().
				Enforce(user.GetLoggedUserIdFromJwtClaims(ctx), endpointName, "*")
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, user.ErrUserNotAuthorised
			}
			return next(ctx, request)
		}
	}
}
