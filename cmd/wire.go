//go:build wireinject
// +build wireinject

package main

import (
	"database/sql"
	"strings"

	"github.com/cyancyan2020/iam-platform/internal/handler"
	"github.com/cyancyan2020/iam-platform/internal/middleware"
	"github.com/cyancyan2020/iam-platform/internal/model"
	"github.com/cyancyan2020/iam-platform/internal/repository"
	"github.com/cyancyan2020/iam-platform/internal/service"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ---------- 具名类型 ----------

type ginMode string
type serverPort string
type jwtSecret string
type jwtExpireHours int
type logChanSize int

// ---------- Provider 函数 ----------

func provideViper() (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	v.SetEnvPrefix("IAM")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	return v, nil
}

func provideGormDB(v *viper.Viper) (*gorm.DB, error) {
	return gorm.Open(mysql.Open(v.GetString("database.dsn")), &gorm.Config{})
}

func provideSQLDB(db *gorm.DB, v *viper.Viper) (*sql.DB, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(v.GetInt("database.max_open_conns"))
	sqlDB.SetMaxIdleConns(v.GetInt("database.max_idle_conns"))
	return sqlDB, nil
}

func provideRedisClient(v *viper.Viper) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     v.GetString("redis.addr"),
		Password: v.GetString("redis.password"),
		DB:       v.GetInt("redis.db"),
	})
}

func provideGinMode(v *viper.Viper) ginMode           { return ginMode(v.GetString("server.mode")) }
func provideServerPort(v *viper.Viper) serverPort       { return serverPort(v.GetString("server.port")) }
func provideJWTSecret(v *viper.Viper) jwtSecret         { return jwtSecret(v.GetString("jwt.secret")) }
func provideJWTExpire(v *viper.Viper) jwtExpireHours    { return jwtExpireHours(v.GetInt("jwt.expire_hours")) }
func provideLogChanSizeVal(v *viper.Viper) logChanSize {
	size := v.GetInt("log.channel_size")
	if size <= 0 { size = 1000 }
	if size > 100000 { size = 100000 }
	return logChanSize(size)
}
func provideLogChan(size logChanSize) chan model.OperationLog {
	return make(chan model.OperationLog, int(size))
}

// 包装构造函数，将具名类型转为原始类型
func provideUserService(
	userRepo repository.UserRepository,
	tokenVersionRepo repository.TokenVersionRepository,
	roleRepo repository.RoleRepository,
	secret jwtSecret,
	expire jwtExpireHours,
) service.UserService {
	return service.NewUserService(userRepo, tokenVersionRepo, roleRepo, string(secret), int(expire))
}

// ---------- Wire Sets ----------

var infraSet = wire.NewSet(
	provideViper,
	provideGormDB,
	provideSQLDB,
	provideRedisClient,
	provideGinMode,
	provideServerPort,
	provideJWTSecret,
	provideJWTExpire,
	provideLogChanSizeVal,
	provideLogChan,
)

var repoSet = wire.NewSet(
	repository.NewUserRepository,
	repository.NewTokenVersionRepository,
	repository.NewPermissionRepository,
	repository.NewRoleRepository,
	repository.NewRolePermissionRepository,
	repository.NewOperationLogRepository,
)

var svcSet = wire.NewSet(
	provideUserService,
	service.NewRoleService,
	service.NewLogService,
)

var handlerSet = wire.NewSet(
	handler.NewUserHandler,
	handler.NewRoleHandler,
	handler.NewPermissionHandler,
	handler.NewLogHandler,
)

// ---------- Components ----------

type Components struct {
	DB          *gorm.DB
	SQLDB       *sql.DB
	RedisClient *redis.Client
	GinMode     ginMode
	ServerPort  serverPort
	JWTSecret   jwtSecret
	JWTExpire   jwtExpireHours
	LogChan     chan model.OperationLog
	UserRepo         repository.UserRepository
	RoleRepo         repository.RoleRepository
	PermRepo         repository.PermissionRepository
	TokenVersionRepo repository.TokenVersionRepository
	LogRepo          repository.OperationLogRepository
	RateLimiter      middleware.RateLimiter
	UserHandler *handler.UserHandler
	RoleHandler *handler.RoleHandler
	PermHandler *handler.PermissionHandler
	LogHandler  *handler.LogHandler
}

func InitComponents() (*Components, error) {
	panic(wire.Build(
		infraSet,
		repoSet,
		svcSet,
		handlerSet,
		middleware.NewRedisRateLimiter,
		wire.Struct(new(Components), "*"),
	))
}
