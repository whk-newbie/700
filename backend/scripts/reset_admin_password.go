package main

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"github.com/spf13/viper"
)

type User struct {
	ID           uint   `gorm:"primaryKey"`
	Username     string `gorm:"uniqueIndex;not null;size:50"`
	PasswordHash string `gorm:"not null;size:255"`
	Email        string `gorm:"size:100"`
	Role         string `gorm:"not null;size:20;default:'user'"`
	IsActive     bool   `gorm:"default:true"`
}

func main() {
	// 尝试从.env文件读取配置
	workDir, _ := os.Getwd()
	envPath := filepath.Join(workDir, ".env")
	if _, err := os.Stat(envPath); err == nil {
		viper.SetConfigFile(envPath)
		viper.SetConfigType("env")
		viper.ReadInConfig()
	}

	// 从环境变量或配置文件读取数据库配置
	dbHost := getConfig("DATABASE_HOST", "localhost")
	dbPort := getConfig("DATABASE_PORT", "5432")
	dbUser := getConfig("DATABASE_USER", "lineuser")
	dbPassword := getConfig("DATABASE_PASSWORD", "linepass")
	dbName := getConfig("DATABASE_DBNAME", "line_management")

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

	username := "admin"
	password := "admin123"

	// 查找用户
	var user User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			fmt.Printf("用户 %s 不存在，正在创建...\n", username)
			
			// 生成密码hash
			passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				fmt.Printf("生成密码hash失败: %v\n", err)
				os.Exit(1)
			}

			// 创建管理员用户
			user = User{
				Username:     username,
				PasswordHash: string(passwordHash),
				Role:         "admin",
				IsActive:     true,
			}

			if err := db.Create(&user).Error; err != nil {
				fmt.Printf("创建管理员失败: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("✅ 管理员账号创建成功！\n")
			fmt.Printf("   用户名: %s\n", username)
			fmt.Printf("   密码: %s\n", password)
		} else {
			fmt.Printf("查询用户失败: %v\n", err)
			os.Exit(1)
		}
	} else {
		// 验证现有密码hash
		err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
		if err != nil {
			fmt.Printf("⚠️  现有密码hash不正确，正在重置...\n")
			
			// 重新生成密码hash
			passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				fmt.Printf("生成密码hash失败: %v\n", err)
				os.Exit(1)
			}

			// 更新密码
			user.PasswordHash = string(passwordHash)
			if err := db.Save(&user).Error; err != nil {
				fmt.Printf("更新密码失败: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("✅ 管理员密码已重置！\n")
			fmt.Printf("   用户名: %s\n", username)
			fmt.Printf("   密码: %s\n", password)
		} else {
			fmt.Printf("✅ 管理员账号已存在且密码正确！\n")
			fmt.Printf("   用户名: %s\n", username)
			fmt.Printf("   密码: %s\n", password)
		}
	}
}

func getConfig(key, defaultValue string) string {
	// 先尝试从viper读取
	if viper.IsSet(key) {
		return viper.GetString(key)
	}
	// 再尝试从环境变量读取
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

