package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/cyancyan2020/iam-platform/internal/handler"
	"github.com/cyancyan2020/iam-platform/internal/middleware"
	"github.com/cyancyan2020/iam-platform/internal/model"
	"github.com/cyancyan2020/iam-platform/internal/repository"
	"github.com/cyancyan2020/iam-platform/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}

	viper.SetEnvPrefix("IAM")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	db, err := gorm.Open(mysql.Open(viper.GetString("database.dsn")), &gorm.Config{})
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("获取数据库实例失败: %v", err)
	}
	sqlDB.SetMaxOpenConns(viper.GetInt("database.max_open_conns"))
	sqlDB.SetMaxIdleConns(viper.GetInt("database.max_idle_conns"))

	rdb := redis.NewClient(&redis.Options{
		Addr:     viper.GetString("redis.addr"),
		Password: viper.GetString("redis.password"),
		DB:       viper.GetInt("redis.db"),
	})

	jwtSecret := viper.GetString("jwt.secret")
	jwtExpireHours := viper.GetInt("jwt.expire_hours")

	userRepo := repository.NewUserRepository(db)
	tokenVersionRepo := repository.NewTokenVersionRepository(rdb)
	permRepo := repository.NewPermissionRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	rolePermRepo := repository.NewRolePermissionRepository(db)
	logRepo := repository.NewOperationLogRepository(db)

	userSvc := service.NewUserService(userRepo, tokenVersionRepo, roleRepo, jwtSecret, jwtExpireHours)
	roleSvc := service.NewRoleService(roleRepo, permRepo, rolePermRepo, userRepo)
	logSvc := service.NewLogService(logRepo)

	userHandler := handler.NewUserHandler(userSvc)
	roleHandler := handler.NewRoleHandler(roleSvc)
	permHandler := handler.NewPermissionHandler(roleSvc)
	logHandler := handler.NewLogHandler(logSvc)

	// 操作日志 channel 和 consumer
	logChan := make(chan model.OperationLog, 1000)
	go middleware.LogConsumer(logRepo, logChan)

	gin.SetMode(viper.GetString("server.mode"))

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.OperationLogMiddleware(logChan))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "iam-platform",
		})
	})

	api := r.Group("/api/v1")
	{
		api.POST("/users/register", userHandler.Register)
		api.POST("/users/login", userHandler.Login)

		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(jwtSecret, tokenVersionRepo))
		protected.Use(middleware.PermissionCheck(permRepo))
		{
			protected.GET("/profile", userHandler.Profile)

			// 用户管理（GET 列表需要数据权限过滤）
			scope := protected.Group("")
			scope.Use(middleware.DataScopeMiddleware(userRepo, roleRepo))
			{
				scope.GET("/users", userHandler.ListUsers)
			}
			protected.POST("/users", userHandler.CreateUser)
			protected.PUT("/users/:id", userHandler.UpdateUser)
			protected.DELETE("/users/:id", userHandler.DeleteUser)

			// 用户角色分配
			protected.POST("/users/:id/role", roleHandler.AssignRole)

			// 角色管理
			protected.GET("/roles", roleHandler.ListRoles)
			protected.POST("/roles", roleHandler.CreateRole)
			protected.PUT("/roles/:id", roleHandler.UpdateRole)
			protected.DELETE("/roles/:id", roleHandler.DeleteRole)

			// 角色权限分配
			protected.GET("/roles/:id/permissions", roleHandler.GetRolePermissions)
			protected.POST("/roles/:id/permissions", roleHandler.SetRolePermissions)

			// 操作日志
			protected.GET("/logs", logHandler.Query)

			// 权限管理
			protected.GET("/permissions", permHandler.ListPermissions)
			protected.POST("/permissions", permHandler.CreatePermission)
			protected.PUT("/permissions/:id", permHandler.UpdatePermission)
			protected.DELETE("/permissions/:id", permHandler.DeletePermission)
		}
	}

	port := viper.GetString("server.port")

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		fmt.Printf("IAM Platform 启动中, 监听端口: %s\n", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务启动失败: %v", err)
		}
	}()

	<-quit
	fmt.Println("server shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("服务关闭失败: %v", err)
	}

	close(logChan)
	sqlDB.Close()
	rdb.Close()
	fmt.Println("server exited")
}
