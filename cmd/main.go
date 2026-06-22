package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/cyancyan2020/iam-platform/internal/handler"
	"github.com/cyancyan2020/iam-platform/internal/middleware"
	"github.com/cyancyan2020/iam-platform/internal/repository"
	"github.com/cyancyan2020/iam-platform/internal/service"
	"github.com/gin-gonic/gin"
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

	jwtSecret := viper.GetString("jwt.secret")
	jwtExpireHours := viper.GetInt("jwt.expire_hours")

	userRepo := repository.NewUserRepository(db)
	userSvc := service.NewUserService(userRepo, jwtSecret, jwtExpireHours)
	userHandler := handler.NewUserHandler(userSvc)

	gin.SetMode(viper.GetString("server.mode"))

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

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
		protected.Use(middleware.AuthMiddleware(jwtSecret))
		{
			protected.GET("/profile", func(c *gin.Context) {
				claims, ok := c.Get("user")
				if !ok {
					c.JSON(http.StatusInternalServerError, gin.H{
						"code":    500,
						"message": "服务器内部错误",
					})
					return
				}
				c.JSON(http.StatusOK, gin.H{
					"code": 200,
					"data": claims,
				})
			})
		}
	}

	port := viper.GetString("server.port")
	fmt.Printf("IAM Platform 启动中, 监听端口: %s\n", port)

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
