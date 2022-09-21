package controller

// 增删改查 inbounds,outbounds,outbound-groups,rules

import (
	"context"
	"io"
	"time"

	"github.com/Dreamacro/clash/adapter/outboundgroup"
	"github.com/Dreamacro/clash/config"
	C "github.com/Dreamacro/clash/constant"
	"github.com/Dreamacro/clash/hub/route"
	"github.com/Dreamacro/clash/log"

	P "github.com/Dreamacro/clash/listener"
	T "github.com/Dreamacro/clash/tunnel"
)

// GetInbounds 获取全部inbounds
func GetInbounds() map[string]C.Inbound {
	return P.GetInbounds()
}

// AddInbounds 新增inbound
func AddInbounds(params []config.InboundConfig) error {
	return P.AddInbounds(params)
}

// DeleteInbounds 删除inbound
func DeleteInbounds(inboundNames []string) {
	P.DeleteInbounds(inboundNames)
}

// GetProxies 获取全部outbound (proxy和proxy-group)
func GetProxies() map[string]C.Proxy {
	return T.Proxies()
}

// AddProxies 新增outbound (proxy)
func AddProxies(params []config.ProxyConfig) error {
	var ps []C.Proxy
	for _, param := range params {
		proxy, err := config.ParseProxy(param)
		if err != nil {
			return err
		}
		ps = append(ps, proxy)

	}

	return T.AddOutbounds(ps)
}

// DeleteProxies 删除outbound/outbound-group 需注意:删除outbound不会影响到outbound-group
func DeleteProxies(proxyNames []string) {
	T.DeleteOutbounds(proxyNames)
}

// AddProxyGroups 新增outbound-group
func AddProxyGroups(groups []outboundgroup.GroupCommonOption) error {
	return T.AddOutboundGroups(groups)
}

// DeleteProxyGroups 删除outbound-group
func DeleteProxyGroups(groupNames []string) {
	T.DeleteOutbounds(groupNames)
}

// GetRules 获取全部规则
func GetRules() []C.Rule {
	return T.Rules()
}

// UpdateRules 更新全部rules
func UpdateRules(params []config.RuleConfig) error {
	rules, err := config.ParseRules(params, T.Proxies())
	if err != nil {
		return err
	}
	T.UpdateRules(rules)
	return nil
}

// StartApi 开启api服务,需go StartApi
func StartApi(addr, secret string) error {
	return route.Start(addr, secret)
}

// StopApi 关闭
func StopApi() {
	route.Shutdown()
}

// SetLog 设置日志输出
func SetLog(out io.Writer, level log.LogLevel) {
	log.SetOutput(out)
	log.SetLevel(level)
}

func SpeedUrl(proxyName string, url string, timeout time.Duration) (uint16, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return T.Proxies()[proxyName].URLTest(ctx, url)
}
