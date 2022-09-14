package adapter

import (
	"fmt"

	"github.com/Dreamacro/clash/adapter/otherinbound"
	"github.com/Dreamacro/clash/adapter/outbound"
	"github.com/Dreamacro/clash/common/structure"
	C "github.com/Dreamacro/clash/constant"
	"github.com/Dreamacro/clash/tunnel"
)

func ParseInbound(mapping map[string]any) (C.OtherInbound, error) {
	decoder := structure.NewDecoder(structure.Option{TagName: "json", WeaklyTypedInput: true})
	otherInbound, existType := mapping["type"].(string)
	if !existType {
		return nil, fmt.Errorf("missing type")
	}
	var (
		inbound C.OtherInbound
		err     error
	)
	switch otherInbound {
	case "socks":
		socksOption := &otherinbound.SocksOption{}
		err = decoder.Decode(mapping, socksOption)
		if err != nil {
			break
		}
		inbound, err = otherinbound.NewSocks(*socksOption, tunnel.TCPIn(), tunnel.UDPIn())
	case "http":
		httpOption := &otherinbound.HttpOption{}
		err = decoder.Decode(mapping, httpOption)
		if err != nil {
			break
		}
		inbound, err = otherinbound.NewHttp(*httpOption, tunnel.TCPIn())
	case "direct":
		// direct 代理直接将流量转发给目标地址
		directOption := &otherinbound.DirectOption{}
		err = decoder.Decode(mapping, directOption)
		if err != nil {
			break
		}
		inbound, err = otherinbound.NewDirect(*directOption, tunnel.TCPIn(), tunnel.UDPIn())
	default:
		return nil, fmt.Errorf("unsupport proxy type: %s", otherInbound)
	}
	return inbound, err
}

func ParseProxy(mapping map[string]any) (C.Proxy, error) {
	decoder := structure.NewDecoder(structure.Option{TagName: "proxy", WeaklyTypedInput: true})
	proxyType, existType := mapping["type"].(string)
	if !existType {
		return nil, fmt.Errorf("missing type")
	}

	var (
		proxy C.ProxyAdapter
		err   error
	)
	switch proxyType {
	case "ss":
		ssOption := &outbound.ShadowSocksOption{}
		err = decoder.Decode(mapping, ssOption)
		if err != nil {
			break
		}
		proxy, err = outbound.NewShadowSocks(*ssOption)
	case "ssr":
		ssrOption := &outbound.ShadowSocksROption{}
		err = decoder.Decode(mapping, ssrOption)
		if err != nil {
			break
		}
		proxy, err = outbound.NewShadowSocksR(*ssrOption)
	case "socks5":
		socksOption := &outbound.Socks5Option{}
		err = decoder.Decode(mapping, socksOption)
		if err != nil {
			break
		}
		proxy = outbound.NewSocks5(*socksOption)
	case "http":
		httpOption := &outbound.HttpOption{}
		err = decoder.Decode(mapping, httpOption)
		if err != nil {
			break
		}
		proxy = outbound.NewHttp(*httpOption)
	case "vmess":
		vmessOption := &outbound.VmessOption{
			HTTPOpts: outbound.HTTPOptions{
				Method: "GET",
				Path:   []string{"/"},
			},
		}
		err = decoder.Decode(mapping, vmessOption)
		if err != nil {
			break
		}
		proxy, err = outbound.NewVmess(*vmessOption)
	case "snell":
		snellOption := &outbound.SnellOption{}
		err = decoder.Decode(mapping, snellOption)
		if err != nil {
			break
		}
		proxy, err = outbound.NewSnell(*snellOption)
	case "trojan":
		trojanOption := &outbound.TrojanOption{}
		err = decoder.Decode(mapping, trojanOption)
		if err != nil {
			break
		}
		proxy, err = outbound.NewTrojan(*trojanOption)
	default:
		return nil, fmt.Errorf("unsupport proxy type: %s", proxyType)
	}

	if err != nil {
		return nil, err
	}

	return NewProxy(proxy), nil
}
