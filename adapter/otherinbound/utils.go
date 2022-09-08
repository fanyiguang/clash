package otherinbound

import (
	"net"

	"github.com/Dreamacro/clash/component/auth"

	"github.com/Dreamacro/clash/common/pool"
	"github.com/Dreamacro/clash/common/sockopt"
	"github.com/Dreamacro/clash/log"
)

type Listener struct {
	listener net.Listener
	addr     string
	closed   bool
}

// RawAddress implements C.Listener
func (l *Listener) RawAddress() string {
	return l.addr
}

// Address implements C.Listener
func (l *Listener) Address() string {
	return l.listener.Addr().String()
}

// Close implements C.Listener
func (l *Listener) Close() error {
	l.closed = true
	return l.listener.Close()
}

func NewTCP(addr string, handleSocks func(conn net.Conn)) (*Listener, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	sl := &Listener{
		listener: l,
		addr:     addr,
	}
	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				if sl.closed {
					break
				}
				continue
			}
			go handleSocks(c)
		}
	}()

	return sl, nil
}

func NewUDP(addr string, handleSocksUDP func(pc net.PacketConn, buf []byte, addr net.Addr)) (*UDPListener, error) {
	l, err := net.ListenPacket("udp", addr)
	if err != nil {
		return nil, err
	}

	if err := sockopt.UDPReuseaddr(l.(*net.UDPConn)); err != nil {
		log.Warnln("Failed to Reuse UDP Address: %s", err)
	}

	sl := &UDPListener{
		packetConn: l,
		addr:       addr,
	}
	go func() {
		for {
			buf := pool.Get(pool.UDPBufferSize)
			n, remoteAddr, err := l.ReadFrom(buf)
			if err != nil {
				pool.Put(buf)
				if sl.closed {
					break
				}
				continue
			}
			handleSocksUDP(l, buf[:n], remoteAddr)
		}
	}()

	return sl, nil
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewAuthenticator(users []User) auth.Authenticator {
	authUsers := make([]auth.AuthUser, 0, len(users))
	for _, user := range users {
		authUsers = append(authUsers, auth.AuthUser{
			User: user.Username,
			Pass: user.Password,
		})
	}

	return auth.NewAuthenticator(authUsers)
}
