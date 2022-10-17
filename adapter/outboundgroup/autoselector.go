package outboundgroup

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/Dreamacro/clash/adapter/outbound"
	"github.com/Dreamacro/clash/common/singledo"
	"github.com/Dreamacro/clash/component/dialer"
	C "github.com/Dreamacro/clash/constant"
	"github.com/Dreamacro/clash/constant/provider"
)

var defaultFailedTimeout = time.Second * 5

//// AutoSelector 封装一层，用于处理gc的时候停止定时器
//type AutoSelector struct {
//	*AutoSelector
//
//
//}

type AutoSelector struct {
	*outbound.Base
	providers []provider.ProxyProvider
	single    *singledo.Single

	// failedProxies 存储所有近期失败过的代理，当所有代理近期都失败过时，逐一尝试所有代理
	failedProxies sync.Map

	// failedTimeout 代理失败后重新尝试的时间间隔
	failedTimeout time.Duration

	// 记录当前代理
	currentProxyName string
}

func (a *AutoSelector) Alive() bool {
	return true
}

func (a *AutoSelector) DelayHistory() []C.DelayHistory {
	return make([]C.DelayHistory, 0)
}

func (a *AutoSelector) LastDelay() uint16 {
	proxies := a.findCandidatesProxy()
	if len(proxies) == 0 {
		return 0
	}
	for _, proxy := range proxies {
		d := proxy.LastDelay()
		return d
	}

	return 0
}

func (a *AutoSelector) URLTest(ctx context.Context, url string) (uint16, error) {
	proxies := a.findCandidatesProxy()
	if len(proxies) == 0 {
		return 0, errors.New("no available proxies")
	}
	for _, proxy := range proxies {
		t, err := proxy.URLTest(ctx, url)
		if err == nil {
			a.currentProxyName = proxy.Name()
			return t, nil
		}
		a.failedProxies.Store(proxy.Name(), time.Now())
	}

	return 0, errors.New("no available proxies")
}

func (a *AutoSelector) Dial(metadata *C.Metadata) (C.Conn, error) {
	proxies := a.findCandidatesProxy()
	if len(proxies) == 0 {
		return nil, errors.New("no available proxies")
	}
	for _, proxy := range proxies {
		conn, err := proxy.Dial(metadata)
		if err == nil {
			a.currentProxyName = proxy.Name()
			return conn, nil
		}
		a.failedProxies.Store(proxy.Name(), time.Now())
	}

	return nil, errors.New("no available proxies")
}

func (a *AutoSelector) DialUDP(metadata *C.Metadata) (C.PacketConn, error) {
	proxies := a.findCandidatesProxy()
	if len(proxies) == 0 {
		return nil, errors.New("no available proxies")
	}
	for _, proxy := range proxies {
		conn, err := proxy.DialUDP(metadata)
		if err == nil {
			a.currentProxyName = proxy.Name()
			return conn, nil
		}
		a.failedProxies.Store(proxy.Name(), time.Now())
	}

	return nil, errors.New("no available proxies")
}

func (a *AutoSelector) DialContext(ctx context.Context, metadata *C.Metadata, opts ...dialer.Option) (C.Conn, error) {
	proxies := a.findCandidatesProxy()
	if len(proxies) == 0 {
		return nil, errors.New("no available proxies")
	}
	for _, proxy := range proxies {
		conn, err := proxy.DialContext(ctx, metadata, a.Base.DialOptions(opts...)...)
		if err == nil {
			a.currentProxyName = proxy.Name()
			return conn, nil
		}
		a.failedProxies.Store(proxy.Name(), time.Now())
	}

	return nil, errors.New("no available proxies")
}

// ListenPacketContext implements C.ProxyAdapter
func (a *AutoSelector) ListenPacketContext(ctx context.Context, metadata *C.Metadata, opts ...dialer.Option) (C.PacketConn, error) {
	proxies := a.findCandidatesProxy()
	if len(proxies) == 0 {
		return nil, errors.New("no available proxies")
	}

	for _, proxy := range proxies {
		pc, err := proxy.ListenPacketContext(ctx, metadata, a.Base.DialOptions(opts...)...)
		if err == nil {
			a.currentProxyName = proxy.Name()
			return pc, nil
		}
		a.failedProxies.Store(proxy.Name(), time.Now())
	}

	return nil, nil
}

func (a *AutoSelector) Now() string {
	if a.currentProxyName != "" {
		return a.currentProxyName
	}

	c := a.findCandidatesProxy()
	if len(c) > 0 {
		return c[0].Name()
	}

	return ""
}

func (a *AutoSelector) findCandidatesProxy() []C.Proxy {
	// 获取所有可用代理
	aliveProxies := a.proxies(true)

	// 获取所有被阻止的代理
	blockedProxies := a.getBlockedProxiesName()

	p := make([]C.Proxy, 0, len(aliveProxies))

	// 删除所有被阻止的代理
	for _, proxy := range aliveProxies {
		blocked := false
		for _, bn := range blockedProxies {
			if bn == proxy.Name() {
				blocked = true
				break
			}
		}
		if !blocked {
			p = append(p, proxy)
		}
	}

	// 如果所有的代理都被阻止，返回所有代理
	if len(p) == 0 {
		p = aliveProxies
	}

	// 如果当前代理在可用列表中，把当前代理放到第一个
	if len(p) > 1 && a.currentProxyName != "" {
		currentIndex := 0
		for id, proxy := range p {
			if proxy.Name() == a.currentProxyName {
				currentIndex = id
				break
			}
		}
		if currentIndex != 0 {
			p = append(p[currentIndex:], p[:currentIndex]...)
		}
	}

	return p
}

func (a *AutoSelector) getBlockedProxiesName() []string {
	timeout := time.Now().Add(-a.failedTimeout)

	list := make([]string, 0, 8)
	a.failedProxies.Range(func(key, value interface{}) bool {
		if value.(time.Time).Before(timeout) {
			a.failedProxies.Delete(key)
		}
		list = append(list, key.(string))
		return true
	})

	return list
}

func (a *AutoSelector) proxies(touch bool) []C.Proxy {
	elm, _, _ := a.single.Do(func() (any, error) {
		return getProvidersProxies(a.providers, touch), nil
	})

	return elm.([]C.Proxy)
}

// Unwrap 返回可用代理列表
func (a *AutoSelector) Unwrap(metadata *C.Metadata) C.Proxy {
	return a
}

// MarshalJSON implements C.ProxyAdapter
func (a *AutoSelector) MarshalJSON() ([]byte, error) {
	var all []string
	for _, proxy := range a.proxies(false) {
		all = append(all, proxy.Name())
	}
	return json.Marshal(map[string]any{
		"type": a.Type().String(),
		"now":  a.Now(),
		"all":  all,
	})
}

func NewAutoSelector(option *GroupCommonOption, providers []provider.ProxyProvider) *AutoSelector {
	as := &AutoSelector{
		Base: outbound.NewBase(outbound.BaseOption{
			Name:        option.Name,
			Type:        C.AutoSelector,
			Interface:   option.Interface,
			RoutingMark: option.RoutingMark,
		}),
		providers: providers,
		single:    singledo.NewSingle(defaultGetProxiesDuration),
	}

	return as
	// wrapper := &AutoSelector{as}
	// runtime.SetFinalizer(wrapper, stopAutoSelector)
	// return wrapper
}
