package outboundgroup

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"

	"github.com/Dreamacro/clash/adapter/outbound"
	"github.com/Dreamacro/clash/common/singledo"
	"github.com/Dreamacro/clash/component/dialer"
	C "github.com/Dreamacro/clash/constant"
	"github.com/Dreamacro/clash/constant/provider"
)

type Relay struct {
	*outbound.Base
	single    *singledo.Single
	providers []provider.ProxyProvider
	hasSsh    bool
}

// DialContext implements C.ProxyAdapter
func (r *Relay) DialContext(ctx context.Context, metadata *C.Metadata, opts ...dialer.Option) (C.Conn, error) {
	var proxies []C.Proxy
	for _, proxy := range r.proxies(metadata, true) {
		if proxy.Type() != C.Direct {
			proxies = append(proxies, proxy)
		}
	}

	switch len(proxies) {
	case 0:
		return outbound.NewDirect().DialContext(ctx, metadata, r.Base.DialOptions(opts...)...)
	case 1:
		return proxies[0].DialContext(ctx, metadata, r.Base.DialOptions(opts...)...)
	}

	// 线路存在ssh,尝试复用连接
	if r.hasSsh {
		return r.dialWithSsh(ctx, metadata, proxies, opts...)
	}

	first := proxies[0]
	last := proxies[len(proxies)-1]

	firstMeta, err := addrToMetadata(proxies[1].Addr())
	if err != nil {
		return nil, fmt.Errorf("%s addrToMetadata failed", proxies[1].Name())
	}

	var c net.Conn
	c, err = first.DialContext(ctx, firstMeta, r.Base.DialOptions(opts...)...)
	if err != nil {
		return nil, fmt.Errorf("relay DialContext failed, first jump = %s, %v", first.Name(), err)
	}

	if len(proxies) > 2 {
		var currentMeta *C.Metadata
		current := proxies[1]
		for _, next := range proxies[2:] {
			currentMeta, err = addrToMetadata(next.Addr())
			if err != nil {
				return nil, err
			}

			c, err = current.StreamConn(c, currentMeta)
			if err != nil {
				return nil, fmt.Errorf("%s connect error: %w", first.Addr(), err)
			}

			current = next
		}
	}

	c, err = last.StreamConn(c, metadata)
	if err != nil {
		return nil, fmt.Errorf("%s connect error: %w", last.Addr(), err)
	}

	return outbound.NewConn(c, r), nil
}

// 进入到该方法时 len(proxies) > 1
func (r *Relay) dialWithSsh(ctx context.Context, metadata *C.Metadata, proxies []C.Proxy, opts ...dialer.Option) (C.Conn, error) {
	var (
		conn net.Conn
		tmp  int
		err  error
	)
	// 从后往前
	for i := len(proxies) - 1; i >= 0; i-- {
		proxy := proxies[i]
		if proxy.Type() == C.Ssh {
			var meta = metadata
			if i < len(proxies)-1 {
				meta, err = addrToMetadata(proxies[i+1].Addr())
				if err != nil {
					return nil, err
				}
			}
			conn, err = proxy.StreamConn(nil, meta)
			if err != nil {
				if !errors.Is(err, outbound.ErrEmptyConnection) {
					return nil, err
				}
				continue
			}
			tmp = i
			break
		}
	}
	// ssh为最后一跳,并且stream成功,直接返回
	if tmp == len(proxies)-1 {
		return outbound.NewConn(conn, r), nil
	}
	if conn == nil {
		meta, err := addrToMetadata(proxies[1].Addr())
		if err != nil {
			return nil, err
		}
		conn, err = proxies[0].DialContext(ctx, meta, opts...)
		if err != nil {
			return nil, err
		}
		tmp = 1
	}
	for i := tmp; i < len(proxies)-1; i++ {
		meta, err := addrToMetadata(proxies[i+1].Addr())
		if err != nil {
			return nil, err
		}
		conn, err = proxies[i].StreamConn(conn, meta)
		if err != nil {
			return nil, err
		}
	}
	conn, err = proxies[len(proxies)-1].StreamConn(conn, metadata)
	if err != nil {
		return nil, err
	}
	return outbound.NewConn(conn, r), nil
}

func (r *Relay) ListenPacketContext(ctx context.Context, metadata *C.Metadata, opts ...dialer.Option) (C.PacketConn, error) {
	proxies := r.proxies(metadata, true)
	if len(proxies) > 1 {
		return nil, errors.New("unsupported relay udp")
	}
	if len(proxies) == 1 {
		if !proxies[0].SupportUDP() {
			return nil, errors.New("relay first jump not support udp")
		}
		return proxies[0].ListenPacketContext(ctx, metadata, opts...)
	}
	return outbound.NewDirect().ListenPacketContext(ctx, metadata, opts...)
}

// MarshalJSON implements C.ProxyAdapter
func (r *Relay) MarshalJSON() ([]byte, error) {
	var all []string
	for _, proxy := range r.rawProxies(false) {
		all = append(all, proxy.Name())
	}
	return json.Marshal(map[string]any{
		"type": r.Type().String(),
		"all":  all,
	})
}

func (r *Relay) rawProxies(touch bool) []C.Proxy {
	elm, _, _ := r.single.Do(func() (any, error) {
		return getProvidersProxies(r.providers, touch), nil
	})

	return elm.([]C.Proxy)
}

func (r *Relay) proxies(metadata *C.Metadata, touch bool) []C.Proxy {
	proxies := r.rawProxies(touch)

	for n, proxy := range proxies {
		subproxy := proxy.Unwrap(metadata)

		if subproxy != nil && subproxy.Name() == proxy.Name() {
			continue
		}

		for subproxy != nil {
			proxies[n] = subproxy
			subproxy = subproxy.Unwrap(metadata)
		}
	}

	return proxies
}

func NewRelay(option *GroupCommonOption, providers []provider.ProxyProvider) *Relay {
	var supportUDP bool
	proxies := getProvidersProxies(providers, false)
	if len(proxies) == 0 {
		supportUDP = true
	} else if len(proxies) == 1 {
		supportUDP = proxies[0].SupportUDP()
	}
	var hasSsh bool
	for _, proxy := range proxies {
		if proxy.Type() == C.Ssh {
			hasSsh = true
			break
		}
	}
	return &Relay{
		Base: outbound.NewBase(outbound.BaseOption{
			Name:        option.Name,
			Type:        C.Relay,
			Interface:   option.Interface,
			RoutingMark: option.RoutingMark,
			UDP:         supportUDP,
		}),
		single:    singledo.NewSingle(defaultGetProxiesDuration),
		providers: providers,
		hasSsh:    hasSsh,
	}
}
