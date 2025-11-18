package db

import (
	"fmt"
	"hp-server-lib/entity"
	"hp-server-lib/log"
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB
var err error

func init() {
	EnsureDirExists("./data", 0755, true)
	DB, err = gorm.Open(sqlite.Open("./data/hp-lite.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // 设置日志级别为 Info
	})
	if err != nil {
		fmt.Println(err)
	}
	// 获取底层的 sql.DB 实例
	sqlDB, err := DB.DB()
	if err != nil {
		log.Errorf("failed to get sql.DB instance", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(10)         // 设置最大空闲连接数
	sqlDB.SetMaxOpenConns(50)         // 设置最大打开连接数
	sqlDB.SetConnMaxLifetime(30 * 60) // 设置连接的最大生命周期（单位：秒）

	//自动创建表
	DB.AutoMigrate(&entity.UserCustomEntity{})
	DB.AutoMigrate(&entity.UserDeviceEntity{})
	DB.AutoMigrate(&entity.UserConfigEntity{})
	DB.AutoMigrate(&entity.UserStatisticsEntity{})
	DB.AutoMigrate(&entity.UserDomainEntity{})
	DB.AutoMigrate(&entity.UserWafEntity{})
	DB.AutoMigrate(&entity.UserFwdEntity{})
	DB.AutoMigrate(&entity.UserReverseEntity{})
}

func EnsureDirExists(dirPath string, perm os.FileMode, createParent bool) error {
	// 规范化路径
	dirPath = filepath.Clean(dirPath)
	// 检查目录是否存在
	stat, err := os.Stat(dirPath)
	// 目录存在
	if err == nil {
		if !stat.IsDir() {
			return fmt.Errorf("路径存在但不是目录: %s", dirPath)
		}
		return nil
	}
	// 目录不存在，创建它
	if os.IsNotExist(err) {
		if createParent {
			return os.MkdirAll(dirPath, perm)
		}
		return os.Mkdir(dirPath, perm)
	}
	// 其他错误（如权限不足）
	return err
}
