package common

import (
	"context"
	user "github.com/beezlabs-org/go-microservices-lib/service/domain/user"
	"github.com/golang-jwt/jwt"
)

func GetLoggedUserIdFromJwtClaims(ctx context.Context) string {
	jwtMap := ctx.Value("props").(jwt.MapClaims)
	return jwtMap["oid"].(string)
}

func CheckIfUserExists(ctx context.Context, id string) (bool, error) {
	svc := user.GetUserService()
	return svc.CheckIfLoggedInUserExists(ctx, id)
}
