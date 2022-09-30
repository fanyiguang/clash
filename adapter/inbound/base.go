package inbound

import (
	"encoding/json"

	C "github.com/Dreamacro/clash/constant"
	"github.com/Dreamacro/clash/log"
)

type Base struct {
	addr        string
	inboundName string
	inboundType C.InboundType

	rawConfig any
}

func (b Base) Name() string {
	return b.inboundName
}

func (b Base) Type() C.InboundType {
	return b.inboundType
}

func (b Base) RawAddress() string {
	return b.addr
}

func (b Base) newMetadata() *C.Metadata {
	return &C.Metadata{
		Inbound: b.inboundName,
	}
}

func (b Base) GetRawConfigString() string {
	if b.rawConfig == nil {
		return ""
	}

	s, err := json.MarshalIndent(b.rawConfig, "", "  ")
	if err != nil {
		log.Warnln("Marshal config error: %s", err.Error())
		return ""
	}
	return string(s)
}
