package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"

	_ "github.com/lib/pq"
)

// DBConfig 数据库配置
type DBConfig struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
}

// 从环境变量获取数据库配置
func getDBConfigFromEnv() *DBConfig {
	host := getEnvOrDefault("DB_HOST", "120.46.147.53")
	portStr := getEnvOrDefault("DB_PORT", "5432")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Printf("警告: 无法解析DB_PORT环境变量，使用默认值5432: %v", err)
		port = 5432
	}

	return &DBConfig{
		Host:     host,
		Port:     port,
		Database: getEnvOrDefault("DB_NAME", "pro_db"),
		User:     getEnvOrDefault("DB_USER", "renoelis"),
		Password: getEnvOrDefault("DB_PASSWORD", "renoelis02@gmail.com"),
	}
}

// 获取环境变量，如果不存在则返回默认值
func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

var (
	db     *sql.DB
	dbOnce sync.Once
	dbErr  error
)

// InitDB 初始化数据库连接
func InitDB() error {
	dbOnce.Do(func() {
		// 获取数据库配置
		dbConfig := getDBConfigFromEnv()

		connStr := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable",
			dbConfig.Host, dbConfig.Port, dbConfig.Database, dbConfig.User, dbConfig.Password)

		log.Printf("尝试连接数据库: %s:%d/%s", dbConfig.Host, dbConfig.Port, dbConfig.Database)

		db, dbErr = sql.Open("postgres", connStr)
		if dbErr != nil {
			log.Printf("数据库连接失败: %v", dbErr)
			return
		}

		// 测试连接
		if dbErr = db.Ping(); dbErr != nil {
			log.Printf("数据库Ping失败: %v", dbErr)
			return
		}

		log.Println("数据库连接成功")

		// 创建表（如果不存在）
		_, dbErr = db.Exec(`
			CREATE TABLE IF NOT EXISTS ding_callback_configs (
				id SERIAL PRIMARY KEY,
				uuid VARCHAR(36) UNIQUE NOT NULL,
				url TEXT,
				aes_key VARCHAR(100) NOT NULL,
				token VARCHAR(100) NOT NULL,
				corp_id VARCHAR(100) NOT NULL,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`)
		if dbErr != nil {
			log.Printf("创建表失败: %v", dbErr)
			return
		}

		log.Println("数据库表初始化成功")
	})

	return dbErr
}

// GetDB 获取数据库连接
func GetDB() *sql.DB {
	if db == nil {
		log.Println("警告: 尝试获取未初始化的数据库连接，自动初始化")
		if err := InitDB(); err != nil {
			log.Printf("自动初始化数据库失败: %v", err)
			return nil
		}
	}
	return db
}
