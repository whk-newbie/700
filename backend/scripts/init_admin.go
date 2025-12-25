package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type User struct {
	ID           uint   `gorm:"primaryKey"`
	Username     string `gorm:"uniqueIndex;not null"`
	PasswordHash string `gorm:"not null"`
	Role         string `gorm:"not null;default:'user'"`
	IsActive     bool   `gorm:"default:true"`
}

func main() {
	// 从环境变量获取数据库配置
	dbHost := os.Getenv("DATABASE_HOST")
	if dbHost == "" {
		dbHost = "postgres"
	}
	dbPort := os.Getenv("DATABASE_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbUser := os.Getenv("DATABASE_USER")
	if dbUser == "" {
		dbUser = "lineuser"
	}
	dbPassword := os.Getenv("DATABASE_PASSWORD")
	if dbPassword == "" {
		dbPassword = "linepass"
	}
	dbName := os.Getenv("DATABASE_DBNAME")
	if dbName == "" {
		dbName = "line_management"
	}

	// 构建数据库连接字符串
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	// 连接数据库
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("❌ 连接数据库失败: %v\n", err)
		os.Exit(1)
	}

	// 检查admin用户是否存在
	var user User
	result := db.Where("username = ?", "admin").First(&user)

	if result.Error == nil {
		// admin用户已存在，更新密码
		password := "admin123"
		hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
		if err != nil {
			fmt.Printf("❌ 生成密码哈希失败: %v\n", err)
			os.Exit(1)
		}

		db.Model(&user).Update("password_hash", string(hash))
		fmt.Printf("✅ Admin用户密码已更新\n")
		fmt.Printf("   用户名: admin\n")
		fmt.Printf("   密码: admin123\n")
	} else if result.Error == gorm.ErrRecordNotFound {
		// admin用户不存在，创建新用户
		password := "admin123"
		hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
		if err != nil {
			fmt.Printf("❌ 生成密码哈希失败: %v\n", err)
			os.Exit(1)
		}

		adminUser := User{
			Username:     "admin",
			PasswordHash: string(hash),
			Role:         "admin",
			IsActive:     true,
		}

		if err := db.Create(&adminUser).Error; err != nil {
			fmt.Printf("❌ 创建admin用户失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Admin用户创建成功\n")
		fmt.Printf("   用户名: admin\n")
		fmt.Printf("   密码: admin123\n")
	} else {
		fmt.Printf("❌ 查询admin用户失败: %v\n", result.Error)
		os.Exit(1)
	}
}

