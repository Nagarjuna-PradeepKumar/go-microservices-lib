package casbin

import (
	internalcasbin "github.com/beezlabs-org/go-microservices-lib/internal/authorization/casbin"
	casbin "github.com/casbin/casbin/v2"
	"github.com/go-kit/log"
	"os"
)

var httpRoutePolicies = [][]string{
	{"role_user", "/user/:id", "GET"},
	{"*", "/user/signup", "POST"},
	{"role_user", "/user/app", "GET"},
	{"role_user", "/metrics", "GET"},
}

var domainPolicies = [][]string{
	{"sub", "obj", "act"},
}

var endpointPolicies = [][]string{
	{"role_user", "create-user", "*"},
	{"role_user", "check-if-user-exists", "*"},
	{"role_user", "get-user-by-id", "*"},
	{"role_user", "get-app-data", "*"},
}

func InitDefaultPolicyAndGetEnforcer(logger log.Logger) *casbin.Enforcer {
	enforcer := internalcasbin.InitCasbinAndGetEnforcer(logger)
	policies := append(httpRoutePolicies, domainPolicies...)
	policies = append(policies, endpointPolicies...)
	for _, policy := range policies {
		_, err := enforcer.AddPolicy(policy)
		if err != nil {
			os.Exit(1)
		}
	}
	enforcer.AddFunction("customMatch", KeyMatchFunc)
	return enforcer
}

func KeyMatchFunc(_ ...interface{}) (interface{}, error) {
	//Your custom key match logic can go here
	return false, nil
}
