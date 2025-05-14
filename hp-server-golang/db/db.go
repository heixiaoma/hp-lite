package db

import (
	"fmt"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"hp-server-lib/entity"
	"log"
)

var DB *gorm.DB
var err error

func init() {
	DB, err = gorm.Open(sqlite.Open("hp-lite.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // 设置日志级别为 Info
	})
	if err != nil {
		fmt.Println(err)
	}
	// 获取底层的 sql.DB 实例
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("failed to get sql.DB instance", err)
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
}
