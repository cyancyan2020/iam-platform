package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/cyancyan2020/iam-platform/internal/middleware"
	pkgl "github.com/cyancyan2020/iam-platform/pkg/log"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// Wire 依赖注入：自动装配所有组件
	comp, err := InitComponents()
	if err != nil {
		pkgl.Fatal("组件初始化失败", zap.Error(err))
	}

	// 初始化 Zap 日志
	pkgl.Init(string(comp.GinMode))
	defer pkgl.Sync()

	// 操作日志 consumer
	var logWg sync.WaitGroup
	logWg.Add(1)
	go func() {
		defer logWg.Done()
		middleware.LogConsumer(comp.LogRepo, comp.LogChan)
	}()

	gin.SetMode(string(comp.GinMode))

	r := gin.New()
	r.Use(middleware.TraceIDMiddleware())
	r.Use(middleware.ZapLoggerMiddleware())
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.OperationLogMiddleware(comp.LogChan))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "iam-platform",
		})
	})

	api := r.Group("/api/v1")
	{
		api.POST("/users/register", comp.UserHandler.Register)
		api.POST("/users/login", middleware.LoginRateLimitMiddleware(comp.RateLimiter, 5, time.Minute), comp.UserHandler.Login)

		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(string(comp.JWTSecret), comp.TokenVersionRepo))
		protected.Use(middleware.PermissionCheck(comp.PermRepo))
		{
			protected.GET("/profile", comp.UserHandler.Profile)

			scope := protected.Group("")
			scope.Use(middleware.DataScopeMiddleware(comp.UserRepo, comp.RoleRepo))
			{
				scope.GET("/users", comp.UserHandler.ListUsers)
			}
			protected.POST("/users", comp.UserHandler.CreateUser)
			protected.PUT("/users/:id", comp.UserHandler.UpdateUser)
			protected.DELETE("/users/:id", comp.UserHandler.DeleteUser)

			protected.POST("/users/:id/role", comp.RoleHandler.AssignRole)

			protected.GET("/roles", comp.RoleHandler.ListRoles)
			protected.POST("/roles", comp.RoleHandler.CreateRole)
			protected.PUT("/roles/:id", comp.RoleHandler.UpdateRole)
			protected.DELETE("/roles/:id", comp.RoleHandler.DeleteRole)

			protected.GET("/roles/:id/permissions", comp.RoleHandler.GetRolePermissions)
			protected.POST("/roles/:id/permissions", comp.RoleHandler.SetRolePermissions)

			protected.GET("/logs", comp.LogHandler.Query)

			protected.GET("/permissions", comp.PermHandler.ListPermissions)
			protected.POST("/permissions", comp.PermHandler.CreatePermission)
			protected.PUT("/permissions/:id", comp.PermHandler.UpdatePermission)
			protected.DELETE("/permissions/:id", comp.PermHandler.DeletePermission)
		}
	}

	port := string(comp.ServerPort)
	srv := &http.Server{Addr: ":" + port, Handler: r}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		pkgl.Info("服务启动", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			pkgl.Fatal("服务启动失败", zap.Error(err))
		}
	}()

	<-quit
	pkgl.Info("server shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		pkgl.Fatal("服务关闭失败", zap.Error(err))
	}

	close(comp.LogChan)
	logWg.Wait()
	comp.SQLDB.Close()
	comp.RedisClient.Close()
	pkgl.Info("server exited")
}
