package service

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"hp-server-lib/bean"
	"hp-server-lib/db"
	"hp-server-lib/entity"
	"hp-server-lib/net/acme"
	"log"
	"strconv"
	"strings"
	"time"
)

type DomainService struct {
}

func (receiver *DomainService) DomainList(userId int, page int, pageSize int) *bean.ResPage {
	var results []*entity.UserDomainEntity
	var total int64
	if userId < 0 {
		db.DB.Model(&entity.UserDomainEntity{}).Order("id desc").Count(&total).Offset((page - 1) * pageSize).Limit(pageSize).Find(&results)
	} else {
		db.DB.Model(&entity.UserDomainEntity{}).Where("user_id = ?", userId).Order("id desc").Count(&total).Offset((page - 1) * pageSize).Limit(pageSize).Find(&results)
	}

	for _, data := range results {
		if len(data.CertificateKey) > 0 && len(data.CertificateContent) > 0 {
			// 解码 PEM 数据
			block, _ := pem.Decode([]byte(data.CertificateContent))
			if block == nil {
				data.Tips = "证书格式错误.."
				continue
			}
			// 解析证书
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				log.Printf(err.Error())
				data.Tips = "证书格式错误."
				continue
			} else {
				// 获取当前时间
				now := time.Now()
				// 检查证书是否过期
				if now.After(cert.NotAfter) {
					data.Tips = "证书已经过期"
				} else {
					duration := cert.NotAfter.Sub(now)
					daysRemaining := int(duration.Hours() / 24)
					data.Tips = "证书还有" + strconv.Itoa(daysRemaining) + "天过期"
				}
			}
		} else {
			data.Tips = "无证书"
		}
	}
	// 计算总记录数并执行分页查询
	return bean.PageOk(total, results)
}

func (receiver *DomainService) DomainListByKey(userId int, keyword string) *bean.ResData {
	var results []entity.UserDomainEntity
	if userId < 0 {
		tx := db.DB.Model(&entity.UserDomainEntity{})
		if len(keyword) > 0 {
			tx.Where("domain like ?", "%"+keyword+"%")
		}
		tx.Order("id desc").Find(&results)
	} else {
		model := db.DB.Model(&entity.UserDomainEntity{})
		if len(keyword) > 0 {
			model.Where("domain like ? and user_id = ? ", "%"+keyword+"%", userId)
		} else {
			model.Where("user_id = ?", userId)
		}
		model.Order("id desc").Find(&results)
	}
	return bean.ResOk(results)
}

func (receiver *DomainService) RemoveData(id int) bool {
	userQuery := &entity.UserDomainEntity{}
	db.DB.Where("id = ? ", id).First(userQuery)
	if userQuery != nil {
		var results entity.UserDomainEntity
		db.DB.Where("id = ?", id).Delete(&results)
		return true
	}
	return false
}

func (receiver *DomainService) AddData(userDomain entity.UserDomainEntity) error {
	userDomain.CertificateContent = strings.TrimSpace(userDomain.CertificateContent)
	userDomain.CertificateKey = strings.TrimSpace(userDomain.CertificateKey)
	if userDomain.Id == nil {
		var total int64
		db.DB.Model(&entity.UserDomainEntity{}).Where("domain = ?", strings.TrimSpace(*userDomain.Domain)).Count(&total)
		if total > 0 {
			return errors.New("域名已存在")
		}
		db.DB.Save(&userDomain)
	} else {
		db.DB.Model(&entity.UserDomainEntity{}).Where("id = ?", userDomain.Id).Update("certificate_content", userDomain.CertificateContent).Update("certificate_key", userDomain.CertificateKey).Update("desc", userDomain.Desc)
	}
	return nil
}

func (receiver *DomainService) GenSsl(id int) bool {
	userQuery := &entity.UserDomainEntity{}
	db.DB.Where("id = ? ", id).First(userQuery)
	if userQuery != nil {
		receiver.UpdateStatus(id, "证书获取中...")
		go func() {
			cert, err := acme.ConfigAcme.GenCert(*userQuery.Domain)
			if err != nil {
				receiver.UpdateStatus(id, "证书获取失败:"+err.Error())
			} else {
				receiver.UpdateData(id, string(cert.PrivateKey), string(cert.Certificate))
				receiver.UpdateStatus(id, "证书获取完成")
			}
		}()
		return true
	}
	return false
}

func (receiver *DomainService) UpdateData(id int, key, content string) error {
	db.DB.Model(&entity.UserDomainEntity{}).Where("id = ?", id).Update("certificate_key", key).Update("certificate_content", content)
	return nil
}

func (receiver *DomainService) UpdateStatus(id int, status string) error {
	db.DB.Model(&entity.UserDomainEntity{}).Where("id = ?", id).Update("status", status)
	return nil
}
