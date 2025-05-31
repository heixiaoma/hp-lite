package service

import (
	"hp-server-lib/db"
	"hp-server-lib/entity"
)

type MonitorService struct {
}

func (receiver MonitorService) ListData(userId int) []entity.UserConfigEntity {
	var results2 []int
	if userId > 0 {
		var configIds []int
		var results []entity.UserConfigEntity
		db.DB.Model(entity.UserConfigEntity{}).Where("user_id = ?", userId).Find(&results)
		for _, result := range results {
			configIds = append(configIds, *result.Id)
		}
		db.DB.Model(&entity.UserStatisticsEntity{}).Distinct("config_id").Where("config_id in ?", configIds).Pluck("config_id", &results2)
	} else {
		db.DB.Model(&entity.UserStatisticsEntity{}).Distinct("config_id").Pluck("config_id", &results2)
	}

	var resultsConfig []entity.UserConfigEntity
	db.DB.Model(entity.UserConfigEntity{}).Where("id in ?", results2).Find(&resultsConfig)
	return resultsConfig
}

func (receiver MonitorService) DetailData(id string) []entity.UserStatisticsEntity {
	var results2 []entity.UserStatisticsEntity
	db.DB.Model(&entity.UserStatisticsEntity{}).Where("config_id = ?", id).Find(&results2)
	return results2
}
