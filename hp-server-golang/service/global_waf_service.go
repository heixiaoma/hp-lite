package service

import (
	"hp-server-lib/bean"
	"hp-server-lib/db"
	"hp-server-lib/entity"
	"hp-server-lib/log"
	"strings"
	"sync/atomic"

	"github.com/corazawaf/coraza/v3"
)

/* 全局 Web 安全规则（后台「Web 安全」菜单）的增删改。

   这里只负责「规则怎么存」，不管「规则怎么生效」：
   编译成 coraza WAF、挂在哪个入口是 net/http 那边的事 ——
   controller 反过来依赖 service，这层不能再 import 挂载方，两边只用
   GlobalWafVersion 这个自增版本号同步：规则一改就 +1，
   挂载方发现版本变了就重新编译（见 net/http/global_waf.go）。 */

// globalWafVer 规则版本号。初值给 1，好让挂载方用 -1 初值触发首次编译
var globalWafVer int64 = 1

// GlobalWafVersion 当前规则版本号，每次增删改 +1
func GlobalWafVersion() int64 {
	return atomic.LoadInt64(&globalWafVer)
}

// GlobalWafDirectives 拼出当前生效的 seclang 指令（只含启用中的规则）。
// 第二个返回值是参与编译的规则条数，0 表示全局防护当前没有启用。
func GlobalWafDirectives() (string, int) {
	return joinWafDirectives(loadEnabledWafRules())
}

// loadEnabledWafRules 按 id 升序取启用中的规则。
// 固定顺序是因为 SecMarker / skipAfter 这类指令之间有先后依赖，
// 每次编译出来的顺序必须一致，否则同一份规则可能这次生效、下次不生效
func loadEnabledWafRules() []entity.GlobalWafEntity {
	var rules []entity.GlobalWafEntity
	if err := db.DB.Where("enabled = ?", true).Order("id asc").Find(&rules).Error; err != nil {
		log.Errorf("全局Web安全规则读取失败：%v", err)
		return nil
	}
	return rules
}

// joinWafDirectives 把多条规则拼成一份完整指令，每条之间补换行，
// 免得上一条末尾没换行时被和下一条粘成一行
func joinWafDirectives(rules []entity.GlobalWafEntity) (string, int) {
	var b strings.Builder
	n := 0
	for _, r := range rules {
		rule := strings.TrimSpace(r.Rule)
		if rule == "" {
			continue
		}
		b.WriteString(rule)
		b.WriteString("\n")
		n++
	}
	return b.String(), n
}

type GlobalWafService struct {
}

// AddData 新增或更新一条全局规则。
//
// 落库前先把「改完之后」的整套规则编译一遍，编译不过就不写库：
// 一条语法错误或者规则 ID 和已有规则撞车的规则，会让整个全局 WAF 起不来，
// 那样整站防护会一起失效，必须在保存时就拦下来。
func (receiver *GlobalWafService) AddData(msg entity.GlobalWafEntity) error {
	directives, _ := joinWafDirectives(previewWafRules(msg))
	if directives != "" {
		if _, err := coraza.NewWAF(coraza.NewWAFConfig().WithDirectives(directives)); err != nil {
			return err
		}
	}
	if err := db.DB.Save(&msg).Error; err != nil {
		return err
	}
	atomic.AddInt64(&globalWafVer, 1)
	return nil
}

// previewWafRules 算出「改完 msg 之后」会生效的规则列表，专门给落库前试编译用。
// msg.Id 为空是新增，否则替换掉同 id 的那条。
func previewWafRules(msg entity.GlobalWafEntity) []entity.GlobalWafEntity {
	rules := loadEnabledWafRules()
	if msg.Id == nil {
		if msg.Enabled {
			return append(rules, msg)
		}
		return rules
	}
	out := make([]entity.GlobalWafEntity, 0, len(rules)+1)
	hit := false
	for _, r := range rules {
		if r.Id != nil && *r.Id == *msg.Id {
			out = append(out, msg)
			hit = true
			continue
		}
		out = append(out, r)
	}
	// 原来是禁用状态、这次改成启用：它不在启用列表里，补到后面
	if !hit && msg.Enabled {
		out = append(out, msg)
	}
	return out
}

func (receiver *GlobalWafService) ListData(page int, pageSize int) *bean.ResPage {
	var results []*entity.GlobalWafEntity
	var total int64
	db.DB.Model(&entity.GlobalWafEntity{}).Order("id desc").Count(&total).
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&results)
	return bean.PageOk(total, results)
}

func (receiver *GlobalWafService) RemoveData(id int) {
	db.DB.Where("id = ?", id).Delete(&entity.GlobalWafEntity{})
	atomic.AddInt64(&globalWafVer, 1)
}
