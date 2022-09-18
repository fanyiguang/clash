package config

import (
	"encoding/json"
	"errors"

	"github.com/Dreamacro/clash/adapter/otherinbound"
	C "github.com/Dreamacro/clash/constant"

	"gopkg.in/yaml.v3"
)

type OtherInbound struct {
	Name         string                    `json:"name"`
	Type         C.InboundType             `json:"type"`
	HttpOption   otherinbound.HttpOption   `json:"-"`
	SocksOption  otherinbound.SocksOption  `json:"-"`
	DirectOption otherinbound.DirectOption `json:"-"`
}

// _OtherInbound 是一个辅助结构体，用来避免 json.Unmarshal 循环
type _OtherInbound OtherInbound

func (i *OtherInbound) UnmarshalJSON(b []byte) error {
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

func (i *OtherInbound) UnmarshalYAML(node *yaml.Node) error {
	return UnmarshalYAML(node, i)
}

func (i *OtherInbound) MarshalJSON() ([]byte, error) {
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
