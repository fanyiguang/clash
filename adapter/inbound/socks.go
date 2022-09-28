package inbound

import (
	"io"
	"net"
	"strconv"
	"strings"

	"github.com/Dreamacro/clash/adapter/defaultinbound"
	N "github.com/Dreamacro/clash/common/net"
	"github.com/Dreamacro/clash/common/pool"
	"github.com/Dreamacro/clash/component/auth"
	C "github.com/Dreamacro/clash/constant"
	"github.com/Dreamacro/clash/context"
	"github.com/Dreamacro/clash/log"
	"github.com/Dreamacro/clash/transport/socks4"
	"github.com/Dreamacro/clash/transport/socks5"
)

type SocksOption struct {
	Listen string `yaml:"listen" json:"listen"`
	Port   int    `yaml:"port" json:"port"`
	Users  []User `json:"users,omitempty"`
}

type Socks struct {
	Base
	Listener      *Listener
	UDPListener   *UDPListener
	tcpIn         chan<- C.ConnContext
	udpIn         chan<- *defaultinbound.PacketAdapter
	Authenticator auth.Authenticator
}

func (s *Socks) run() error {
	tcpListener, err := NewTCP(s.addr, s.handleSocks)
	if err != nil {
		return err
	}
	log.Infoln("SOCKS OtherInbound %s listening at: %s", s.Base.inboundName, tcpListener.Address())

	udpListener, err := NewUDP(s.addr, s.handleSocksUDP)
	if err != nil {
		_ = tcpListener.Close()
		return err
	}

	s.Listener = tcpListener
	s.UDPListener = udpListener

	return nil
}

func (s *Socks) Close() {
	_ = s.Listener.Close()
	_ = s.UDPListener.Close()
	log.Infoln("SOCKS OtherInbound %s closed", s.Base.inboundName)
}

func (s *Socks) handleSocks(conn net.Conn) {
	conn.(*net.TCPConn).SetKeepAlive(true)
	bufConn := N.NewBufferedConn(conn)
	head, err := bufConn.Peek(1)
	if err != nil {
		_ = conn.Close()
		return
	}

	switch head[0] {
	case socks4.Version:
		s.HandleSocks4(bufConn)
	case socks5.Version:
		s.HandleSocks5(bufConn)
	default:
		_ = conn.Close()
	}
}

func (s *Socks) HandleSocks4(conn net.Conn) {
	addr, _, err := socks4.ServerHandshake(conn, s.Authenticator)
	if err != nil {
		_ = conn.Close()
		return
	}
	connContext := NewSocket(socks5.ParseAddr(addr), conn, C.SOCKS4, s.newMetadata())
	s.tcpIn <- connContext
}

func (s *Socks) HandleSocks5(conn net.Conn) {
	target, command, err := socks5.ServerHandshake(conn, s.Authenticator)
	if err != nil {
		_ = conn.Close()
		return
	}
	if command == socks5.CmdUDPAssociate {
		defer conn.Close()
		io.Copy(io.Discard, conn)
		return
	}

	connContext := NewSocket(target, conn, C.SOCKS5, s.newMetadata())
	s.tcpIn <- connContext
}

func (s *Socks) handleSocksUDP(pc net.PacketConn, buf []byte, addr net.Addr) {
	target, payload, err := socks5.DecodeUDPPacket(buf)
	if err != nil {
		// Unresolved UDP packet, return buffer to the pool
		pool.Put(buf)
		return
	}
	packet := &packet{
		pc:      pc,
		rAddr:   addr,
		payload: payload,
		bufRef:  buf,
	}

	packetAdapter := NewPacket(target, packet, C.SOCKS5, s.newMetadata())

	select {
	case s.udpIn <- packetAdapter:
	default:
	}
}

// NewPacket is PacketAdapter generator
func NewPacket(target socks5.Addr, packet C.UDPPacket, source C.Type, meta *C.Metadata) *defaultinbound.PacketAdapter {
	parseSocksAddr(target, meta)
	meta.NetWork = C.UDP
	meta.Type = source
	if ip, port, err := parseAddr(packet.LocalAddr().String()); err == nil {
		meta.SrcIP = ip
		meta.SrcPort = port
	}

	return &defaultinbound.PacketAdapter{
		UDPPacket: packet,
		Meta:      meta,
	}
}

func NewSocks(option SocksOption, name string, tcpIn chan<- C.ConnContext, udpIn chan<- *defaultinbound.PacketAdapter) (*Socks, error) {
	addr := net.JoinHostPort(option.Listen, strconv.Itoa(option.Port))

	var newAuth auth.Authenticator
	if len(option.Users) > 0 {
		newAuth = NewAuthenticator(option.Users)
	}
	s := &Socks{
		Base: Base{
			inboundName: name,
			inboundType: C.InboundTypeSOCKS,
			addr:        addr,
		},
		tcpIn:         tcpIn,
		udpIn:         udpIn,
		Authenticator: newAuth,
	}

	if err := s.run(); err != nil {
		return nil, err
	}
	return s, nil
}

// NewSocket receive TCP inbound and return ConnContext
func NewSocket(target socks5.Addr, conn net.Conn, source C.Type, meta *C.Metadata) *context.ConnContext {
	parseSocksAddr(target, meta)
	meta.NetWork = C.TCP
	meta.Type = source
	if ip, port, err := parseAddr(conn.RemoteAddr().String()); err == nil {
		meta.SrcIP = ip
		meta.SrcPort = port
	}

	return context.NewConnContext(conn, meta)
}

func parseSocksAddr(target socks5.Addr, meta *C.Metadata) *C.Metadata {
	switch target[0] {
	case socks5.AtypDomainName:
		// trim for FQDN
		meta.Host = strings.TrimRight(string(target[2:2+target[1]]), ".")
		meta.DstPort = strconv.Itoa((int(target[2+target[1]]) << 8) | int(target[2+target[1]+1]))
	case socks5.AtypIPv4:
		ip := net.IP(target[1 : 1+net.IPv4len])
		meta.DstIP = ip
		meta.DstPort = strconv.Itoa((int(target[1+net.IPv4len]) << 8) | int(target[1+net.IPv4len+1]))
	case socks5.AtypIPv6:
		ip := net.IP(target[1 : 1+net.IPv6len])
		meta.DstIP = ip
		meta.DstPort = strconv.Itoa((int(target[1+net.IPv6len]) << 8) | int(target[1+net.IPv6len+1]))
	}

	return meta
}

type UDPListener struct {
	packetConn net.PacketConn
	addr       string
	closed     bool
}

// RawAddress implements C.Listener
func (l *UDPListener) RawAddress() string {
	return l.addr
}

// Address implements C.Listener
func (l *UDPListener) Address() string {
	return l.packetConn.LocalAddr().String()
}

// Close implements C.Listener
func (l *UDPListener) Close() error {
	l.closed = true
	return l.packetConn.Close()
}

type packet struct {
	pc      net.PacketConn
	rAddr   net.Addr
	payload []byte
	bufRef  []byte
}

func (c *packet) Data() []byte {
	return c.payload
}

// WriteBack write UDP packet with source(ip, port) = `addr`
func (c *packet) WriteBack(b []byte, addr net.Addr) (n int, err error) {
	packet, err := socks5.EncodeUDPPacket(socks5.ParseAddrToSocksAddr(addr), b)
	if err != nil {
		return
	}
	return c.pc.WriteTo(packet, c.rAddr)
}

// LocalAddr returns the source IP/Port of UDP Packet
func (c *packet) LocalAddr() net.Addr {
	return c.rAddr
}

func (c *packet) Drop() {
	pool.Put(c.bufRef)
}
