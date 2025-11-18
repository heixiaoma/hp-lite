package task

import (
	"crypto/x509"
	"encoding/pem"
	"hp-server-lib/db"
	"hp-server-lib/entity"
	"hp-server-lib/log"
	"hp-server-lib/service"
	"strconv"
	"time"

	"github.com/gogf/gf/v2/os/gcron"
	"golang.org/x/net/context"
)

// StartSslTask 每天检查下证书过期问题，过期的就自动续签
func StartSslTask() {
	gcron.AddSingleton(context.Background(), "@daily", updateCertificate, "update_certificate")
	//gcron.AddSingleton(context.Background(), "*/30 * * * * *", updateCertificate, "update_certificate")
}

func updateCertificate(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			// 记录异常信息
			log.Errorf("捕获到异常: %v\n", r)
		}
	}()
	log.Info("开始更新证书:", time.Now().Format("2006-01-02 15:04:05"))
	domainService := service.DomainService{}
	pageSiz := 10
	page := 1
	for {
		_, results := getDomainList(page, pageSiz)
		for _, data := range results {
			if len(data.CertificateKey) > 0 && len(data.CertificateContent) > 0 {
				// 解码 PEM 数据
				block, _ := pem.Decode([]byte(data.CertificateContent))
				if block == nil {
					continue
				}
				// 解析证书
				cert, err := x509.ParseCertificate(block.Bytes)
				if err != nil {
					continue
				} else {
					// 获取当前时间
					now := time.Now()
					//提前两天开始检查
					twoDaysBefore := cert.NotAfter.AddDate(0, 0, -2)
					// 检查证书是否过期
					if now.After(twoDaysBefore) {
						ssl := domainService.GenSsl(true, *data.Id)
						if ssl {
							log.Info("证书检查->" + *data.Domain + ",生成成功")
						} else {
							log.Info("证书检查->" + *data.Domain + ",生成失败")
						}
					} else {
						duration := cert.NotAfter.Sub(now)
						daysRemaining := int(duration.Hours() / 24)
						log.Info("证书检查->" + *data.Domain + ",证书还有" + strconv.Itoa(daysRemaining) + "天过期")
					}
				}
			}
		}
		if pageSiz == len(results) {
			page++
		} else {
			break
		}

	}
}

func getDomainList(page, pageSize int) (int64, []*entity.UserDomainEntity) {
	var results []*entity.UserDomainEntity
	var total int64
	db.DB.Model(&entity.UserDomainEntity{}).Order("id desc").Count(&total).Offset((page - 1) * pageSize).Limit(pageSize).Find(&results)
	return total, results
}
