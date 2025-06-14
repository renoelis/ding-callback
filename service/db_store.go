package service

import (
	"database/sql"
	"errors"
	"log"

	"ding_call_back/config"
	"ding_call_back/model"

	"github.com/google/uuid"
)

// DBStore 数据库存储服务
type DBStore struct {
	db *sql.DB
}

// NewDBStore 创建数据库存储服务
func NewDBStore() *DBStore {
	db := config.GetDB()
	if db == nil {
		log.Println("警告: 数据库连接为nil，请确保已调用config.InitDB()")
		return nil
	}

	return &DBStore{
		db: db,
	}
}

// StoreConfig 存储配置到数据库
func (s *DBStore) StoreConfig(config model.CallbackParams) (string, error) {
	if s.db == nil {
		return "", errors.New("数据库连接未初始化")
	}

	// 生成UUID作为唯一标识
	uuid := uuid.New().String()

	// 插入数据库
	_, err := s.db.Exec(
		`INSERT INTO ding_callback_configs (uuid, url, aes_key, token, corp_id) 
		 VALUES ($1, $2, $3, $4, $5)`,
		uuid, config.URL, config.AESKey, config.Token, config.CorpID,
	)
	if err != nil {
		log.Printf("存储配置失败: %v", err)
		return "", err
	}

	return uuid, nil
}

// GetConfig 从数据库获取配置
func (s *DBStore) GetConfig(uuid string) (model.CallbackParams, bool) {
	if s.db == nil {
		log.Println("数据库连接未初始化")
		return model.CallbackParams{}, false
	}

	var config model.CallbackParams

	err := s.db.QueryRow(
		`SELECT url, aes_key, token, corp_id FROM ding_callback_configs 
		 WHERE uuid = $1`,
		uuid,
	).Scan(&config.URL, &config.AESKey, &config.Token, &config.CorpID)

	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("查询配置失败: %v", err)
		}
		return model.CallbackParams{}, false
	}

	return config, true
}
