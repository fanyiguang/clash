package inbound

import (
	"fmt"
	"net"
	"strconv"

	"github.com/Dreamacro/clash/adapter/defaultinbound"
	C "github.com/Dreamacro/clash/constant"
	"github.com/Dreamacro/clash/context"
	"github.com/Dreamacro/clash/log"
)

type Direct struct {
	Base

	Listener    *Listener
	UDPListener *UDPListener
	tcpIn       chan<- C.ConnContext
	udpIn       chan<- *defaultinbound.PacketAdapter

	dstPort string
	host    string
	destIP  net.IP
}

type DirectOption struct {
	Listen       string `yaml:"listen" json:"listen"`
	Port         int    `yaml:"port" json:"port"`
	RedirectAddr string `json:"redirect-addr"` // 重定向地址，把从监听地址收到的数据转发到这个地址
}

func NewDirect(option DirectOption, name string, tcpIn chan<- C.ConnContext, udpIn chan<- *defaultinbound.PacketAdapter) (*Direct, error) {
	addr := net.JoinHostPort(option.Listen, strconv.Itoa(option.Port))

	h, p, err := net.SplitHostPort(option.RedirectAddr)
	if err != nil {
		return nil, fmt.Errorf("address error:%w", err)
	}

	s := &Direct{
		Base: Base{
			inboundName: name,
			inboundType: C.InboundTypeDirect,
			addr:        addr,
		},
		tcpIn:   tcpIn,
		udpIn:   udpIn,
		host:    h,
		dstPort: p,
		destIP:  net.ParseIP(h),
	}

	if err := s.run(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Direct) handleTCP(conn net.Conn) {
	_ = conn.(*net.TCPConn).SetKeepAlive(true)
	s.tcpIn <- s.NewRedirectTCP(conn)
}

func parseAddr(addr string) (net.IP, string, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, "", err
	}

	ip := net.ParseIP(host)
	return ip, port, nil
}

func (s *Direct) NewRedirectTCP(conn net.Conn) *context.ConnContext {
	meta := s.newMetadata()
	meta.DstPort = s.dstPort
	meta.Host = s.host
	meta.DstIP = s.destIP
	meta.Type = C.REDIR
	meta.NetWork = C.TCP

	if ip, port, err := parseAddr(conn.RemoteAddr().String()); err == nil {
		meta.SrcIP = ip
		meta.SrcPort = port
	}
	return context.NewConnContext(conn, meta)
}

func (s *Direct) handleUDP(pc net.PacketConn, buf []byte, addr net.Addr) {
	packet := &packet{
		pc:     pc,
		rAddr:  addr,
		bufRef: buf,
	}
	meta := s.newMetadata()
	meta.DstPort = s.dstPort
	meta.Host = s.host
	meta.DstIP = s.destIP
	meta.Type = C.REDIR
	meta.NetWork = C.UDP

	p := &defaultinbound.PacketAdapter{
		UDPPacket: packet,
		Meta:      meta,
	}
	select {
	case s.udpIn <- p:
	default:
	}
}

func (s *Direct) run() error {
	tcpListener, err := NewTCP(s.addr, s.handleTCP)
	if err != nil {
		return err
	}
	log.Infoln("Direct OtherInbound %s listening at: %s", s.Base.inboundName, tcpListener.Address())

	udpListener, err := NewUDP(s.addr, s.handleUDP)
	if err != nil {
		_ = tcpListener.Close()
		return err
	}

	s.Listener = tcpListener
	s.UDPListener = udpListener

	return nil
}

// Close 关闭监听
func (s *Direct) Close() {
	_ = s.Listener.Close()
	_ = s.UDPListener.Close()
	log.Infoln("Direct OtherInbound %s closed", s.Base.inboundName)
}
