package common

import (
	"context"
	"github.com/golang-jwt/jwt"
)

func GetLoggedUserIdFromJwtClaims(ctx context.Context) string {
	jwtMap := ctx.Value("props").(jwt.MapClaims)
	return jwtMap["oid"].(string)
}

func CheckIfUserExists(ctx context.Context, id string) (bool, error) {
	return true, nil
}
