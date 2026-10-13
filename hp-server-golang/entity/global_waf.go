package entity

import "time"

// GlobalWafEntity 全局 Web 安全规则（后台「Web 安全」菜单）。
//
// 和 user_safe 的区别：user_safe 的规则要绑定到某条穿透 / 反向代理上才生效，
// 这里的规则对所有进站流量生效 —— HTTP(80) / HTTPS(443) 入口各挂一份，
// 不针对某个域名或某个用户，也只开放给管理员改（见 web_server.go 的 middleware.Admin）。
//
// 生效内容是「所有 enabled 规则的集合」：它们拼成一份 seclang 指令，
// 编译出单个 coraza WAF 实例，改完即时重建，不需要重启服务。
type GlobalWafEntity struct {
	/**
	 * 主键
	 */
	Id *int `json:"id" gorm:"primaryKey;autoIncrement"`

	RuleName string `json:"ruleName"`

	Rule string `json:"rule"`

	// Enabled 关掉的规则仍然留在库里，只是不参与编译。
	// 临时摘掉一条规则不用「删掉 → 想起来再重写一遍」
	Enabled bool `json:"enabled"`

	// Remark 给管理员自己看的备注：这条规则挡的是什么
	Remark string `json:"remark"`

	UpdatedAt time.Time `json:"updatedAt"`
}

func (GlobalWafEntity) TableName() string {
	return "global_waf"
}
