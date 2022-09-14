package inbound

import (
	C "github.com/Dreamacro/clash/constant"
	"github.com/Dreamacro/clash/transport/socks5"
)

// PacketAdapter is a UDP Packet adapter for socks/redir/tun
type PacketAdapter struct {
	C.UDPPacket
	Meta *C.Metadata
}

// Metadata returns destination Meta
func (s *PacketAdapter) Metadata() *C.Metadata {
	return s.Meta
}

// NewPacket is PacketAdapter generator
func NewPacket(target socks5.Addr, packet C.UDPPacket, source C.Type) *PacketAdapter {
	metadata := parseSocksAddr(target)
	metadata.NetWork = C.UDP
	metadata.Type = source
	if ip, port, err := parseAddr(packet.LocalAddr().String()); err == nil {
		metadata.SrcIP = ip
		metadata.SrcPort = port
	}

	return &PacketAdapter{
		UDPPacket: packet,
		Meta:      metadata,
	}
}
