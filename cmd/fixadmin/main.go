package main

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	dsn := "root:1234@tcp(127.0.0.1:3306)/iam_platform?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	result := db.Exec("UPDATE user SET role_id = 1 WHERE username = 'admin'")
	if result.Error != nil {
		log.Fatalf("更新失败: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		fmt.Println("未找到 admin 用户，请先注册 username=admin")
	} else {
		fmt.Println("已将 admin 设置为管理员角色，刷新页面重新登录即可")
	}
}
