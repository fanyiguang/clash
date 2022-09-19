package config

import (
	"encoding/json"
	"errors"

	"github.com/Dreamacro/clash/adapter/inbound"
	C "github.com/Dreamacro/clash/constant"

	"gopkg.in/yaml.v3"
)

type InboundConfig struct {
	Name         string               `json:"name"`
	Type         C.InboundType        `json:"type"`
	HttpOption   inbound.HttpOption   `json:"-"`
	SocksOption  inbound.SocksOption  `json:"-"`
	DirectOption inbound.DirectOption `json:"-"`
}

// _OtherInbound 是一个辅助结构体，用来避免 json.Unmarshal 循环
type _OtherInbound InboundConfig

func (i *InboundConfig) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, (*_OtherInbound)(i))
	if err != nil {
		return err
	}

	var v any
	switch i.Type {
	case C.HTTPInbound:
		v = &i.HttpOption
	case C.SOCKSInbound:
		v = &i.SocksOption
	case C.DIRECTInbound:
		v = &i.DirectOption
	default:
		return errors.New("unknown inbound type")
	}

	return json.Unmarshal(b, v)
}

func (i *InboundConfig) UnmarshalYAML(node *yaml.Node) error {
	return UnmarshalYAML(node, i)
}

func (i *InboundConfig) MarshalJSON() ([]byte, error) {
	var v any
	switch i.Type {
	case C.HTTPInbound:
		v = &i.HttpOption
	case C.SOCKSInbound:
		v = &i.SocksOption
	case C.DIRECTInbound:
		v = &i.DirectOption
	default:
		return nil, errors.New("unknown inbound type")
	}

	return MarshalObjects((*_OtherInbound)(i), v)
}
