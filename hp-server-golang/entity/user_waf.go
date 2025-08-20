package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// IPList 自定义类型，用于存储IP列表
type IPList []string

// Value 实现 driver.Valuer 接口，将IPList转为JSON字符串
func (i IPList) Value() (driver.Value, error) {
	if len(i) == 0 {
		return "[]", nil
	}
	return json.Marshal(i)
}

// Scan 实现 sql.Scanner 接口，将JSON字符串转为IPList
func (i *IPList) Scan(value interface{}) error {
	if value == nil {
		*i = []string{}
		return nil
	}

	var strValue string
	switch v := value.(type) {
	case string:
		strValue = v
	case []byte:
		strValue = string(v)
	default:
		return fmt.Errorf("unsupported Scan type for IPList: %T", value)
	}

	if strValue == "" || strValue == "[]" {
		*i = []string{}
		return nil
	}

	return json.Unmarshal([]byte(strValue), i)
}

type UserWafEntity struct {
	/**
	 * 主键
	 */
	Id *int `json:"id" gorm:"primaryKey;autoIncrement"`
	/**
	 * 套餐ID
	 */
	ConfigId int `json:"configId"`

	/**
	 * 用户ID
	 */
	UserId *int `json:"userId"`

	AllowedIPs IPList `json:"allowedIps" gorm:"type:text"` // 使用自定义类型

	BlockedIPs IPList `json:"blockedIps" gorm:"type:text"`

	RateLimit int `json:"rateLimit"`

	OutLimit int `json:"outLimit"`

	InLimit int `json:"inLimit"`

	ConfigDesc string `json:"configDesc"  gorm:"-"`
	Username   string `json:"username" gorm:"-"`

	UserDesc string `json:"userDesc"  gorm:"-"`
}

func (UserWafEntity) TableName() string {
	return "user_waf"
}
