package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 数据库连接
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://lineuser:123456@localhost:5432/line_management?sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("连接数据库失败: %v\n", err)
		os.Exit(1)
	}

	// 生成新密码哈希
	password := "admin123"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		fmt.Printf("生成密码哈希失败: %v\n", err)
		os.Exit(1)
	}

	// 更新admin用户密码
	result := db.Exec("UPDATE users SET password_hash = ? WHERE username = ?", string(hash), "admin")
	if result.Error != nil {
		fmt.Printf("更新密码失败: %v\n", result.Error)
		os.Exit(1)
	}

	if result.RowsAffected == 0 {
		fmt.Println("未找到admin用户")
		os.Exit(1)
	}

	fmt.Printf("✅ 成功重置admin用户密码！\n")
	fmt.Printf("用户名: admin\n")
	fmt.Printf("密码: admin123\n")
	fmt.Printf("新密码哈希: %s\n", string(hash))
}

