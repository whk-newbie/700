package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 从环境变量读取数据库配置
	dbHost := getEnv("DATABASE_HOST", "localhost")
	dbPort := getEnv("DATABASE_PORT", "5432")
	dbUser := getEnv("DATABASE_USER", "lineuser")
	dbPassword := getEnv("DATABASE_PASSWORD", "linepass")
	dbName := getEnv("DATABASE_DBNAME", "line_management")

	// 构建DSN
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		dbHost, dbUser, dbPassword, dbName, dbPort,
	)

	// 连接数据库
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("连接数据库失败: %v\n", err)
		os.Exit(1)
	}

	// 获取管理员信息
	username := getEnv("ADMIN_USERNAME", "admin")
	password := getEnv("ADMIN_PASSWORD", "admin123")
	email := getEnv("ADMIN_EMAIL", "")

	// 生成密码hash
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("生成密码hash失败: %v\n", err)
		os.Exit(1)
	}

	// 检查用户是否已存在
	var count int64
	db.Model(&User{}).Where("username = ?", username).Count(&count)
	if count > 0 {
		fmt.Printf("用户 %s 已存在\n", username)
		os.Exit(0)
	}

	// 创建管理员用户
	user := User{
		Username:     username,
		PasswordHash: string(passwordHash),
		Email:        email,
		Role:         "admin",
		IsActive:     true,
	}

	if err := db.Create(&user).Error; err != nil {
		fmt.Printf("创建管理员失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("管理员账号创建成功！\n")
	fmt.Printf("用户名: %s\n", username)
	fmt.Printf("密码: %s\n", password)
}

type User struct {
	ID           uint   `gorm:"primaryKey"`
	Username     string `gorm:"uniqueIndex;not null;size:50"`
	PasswordHash string `gorm:"not null;size:255"`
	Email        string `gorm:"size:100"`
	Role         string `gorm:"not null;size:20;default:'user'"`
	MaxGroups    *int
	IsActive     bool   `gorm:"default:true"`
	CreatedBy    *uint
	CreatedAt    int64
	UpdatedAt    int64
	DeletedAt    *int64
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

