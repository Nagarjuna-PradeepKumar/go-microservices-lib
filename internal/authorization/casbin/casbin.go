package casbin

import (
	pgadapter "github.com/casbin/casbin-pg-adapter"
	"github.com/casbin/casbin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"os"
	"sync"
)

var once sync.Once
var svc *casbinService

type casbinService struct {
	defaultEnforcer *casbin.Enforcer
	adapter         *pgadapter.Adapter
	logger          log.Logger
}

type Service interface {
	GetDefaultEnforcer() *casbin.Enforcer
	GetNewEnforcer() *casbin.Enforcer
	GetNewFilteredEnforcer(filter interface{}) *casbin.Enforcer
}

func newCasbinService(adapter *pgadapter.Adapter, logger log.Logger) Service {
	svc = &casbinService{
		adapter: adapter,
		logger:  logger,
	}
	return svc
}

func InitCasbinAndGetEnforcer(dbSource interface{}, logger log.Logger) *casbin.Enforcer {
	once.Do(func() {
		_ = level.Info(logger).Log("msg", "Initializing the casbin postgres adapter")
		adaptor, err := pgadapter.NewAdapter(dbSource)
		if err != nil {
			os.Exit(1)
		}
		newCasbinService(adaptor, logger)
	})
	svc := GetService()
	setDefaultEnforcer(svc.GetNewEnforcer())
	return svc.GetDefaultEnforcer()
}

func GetService() Service {
	return svc
}

func setDefaultEnforcer(enforcer *casbin.Enforcer) {
	svc.defaultEnforcer = enforcer
}

func (c casbinService) GetDefaultEnforcer() *casbin.Enforcer {
	return c.defaultEnforcer
}

func (c casbinService) GetNewEnforcer() *casbin.Enforcer {
	enforcer := getEnforcer(c)
	_ = level.Info(c.logger).Log("msg", "Loading policy from database")
	err := enforcer.LoadPolicy()
	if err != nil {
		os.Exit(1)
	}
	return enforcer
}

func (c casbinService) GetNewFilteredEnforcer(filter interface{}) *casbin.Enforcer {
	enforcer := getEnforcer(c)
	_ = level.Info(c.logger).Log("msg", "Loading policy from database")
	err := enforcer.LoadFilteredPolicy(filter)
	if err != nil {
		os.Exit(1)
	}
	return enforcer
}

func getEnforcer(c casbinService) *casbin.Enforcer {
	_ = level.Info(c.logger).Log("msg", "Initializing the casbin enforcer")
	enforcer, err := casbin.NewEnforcer("common/authorization/casbin/casbin_model.conf", c.adapter)
	if err != nil {
		os.Exit(1)
	}
	return enforcer
}
