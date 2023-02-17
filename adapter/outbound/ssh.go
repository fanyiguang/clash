package outbound

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/Dreamacro/clash/component/dialer"
	C "github.com/Dreamacro/clash/constant"
	"github.com/Dreamacro/clash/log"

	"golang.org/x/crypto/ssh"
)

type Ssh struct {
	*Base
	cfg *ssh.ClientConfig
	// client *ssh.Client
	// mu     sync.Mutex

	// relay 所用的client
	rClient *ssh.Client
	rMu     sync.RWMutex
}

type SshOption struct {
	BasicOption
	Name       string `proxy:"name" json:"name"`
	Server     string `proxy:"server" json:"server"`
	Port       int    `proxy:"port" json:"port"`
	UserName   string `proxy:"username" json:"username,omitempty"`
	Password   string `proxy:"password,omitempty" json:"password,omitempty"`     // 密码
	KeyPath    string `proxy:"key_path,omitempty" json:"key_path,omitempty"`     // 私钥地址
	Passphrase string `proxy:"passphrase,omitempty" json:"passphrase,omitempty"` // 私钥密码
}

// StreamConn implements C.ProxyAdapter
// note: relay 专用, 复用连接,但每个 Ssh 对象只能为一条线路
func (s *Ssh) StreamConn(c net.Conn, metadata *C.Metadata) (net.Conn, error) {
	client, err := s.rConnect(c)
	if err != nil {
		return nil, err
	}
	return client.Dial("tcp", metadata.RemoteAddress())
}

func (s *Ssh) DialContext(ctx context.Context, metadata *C.Metadata, opts ...dialer.Option) (_ C.Conn, err error) {
	client, err := s.rConnect(nil)
	if err != nil {
		if !errors.Is(err, ErrEmptyConnection) {
			return nil, err
		}
		conn, err := dialer.DialContext(ctx, "tcp", s.addr, opts...)
		if err != nil {
			return nil, err
		}
		client, err = s.rConnect(conn)
		if err != nil {
			return nil, err
		}
	}

	c, err := client.Dial("tcp", metadata.RemoteAddress())
	if err != nil {
		return nil, err
	}
	return NewConn(c, s), nil
}

// func (s *Ssh) connect(ctx context.Context, opts ...dialer.Option) (*ssh.Client, error) {
// 	if s.client != nil {
// 		return s.client, nil
// 	}
// 	s.mu.Lock()
// 	defer s.mu.Unlock()
// 	if s.client != nil {
// 		return s.client, nil
// 	}

// 	conn, err := dialer.DialContext(ctx, "tcp", s.addr, opts...)
// 	if err != nil {
// 		return nil, err
// 	}

// 	c, chans, reqs, err := ssh.NewClientConn(conn, s.addr, s.cfg)
// 	if err != nil {
// 		conn.Close()
// 		return nil, err
// 	}

// 	s.client = ssh.NewClient(c, chans, reqs)
// 	go func() {
// 		err = s.client.Wait()
// 		c.Close()

// 		log.Warnln("ssh client wait: %v", err)
// 		s.mu.Lock()
// 		s.client = nil
// 		s.mu.Unlock()
// 	}()
// 	return s.client, nil
// }

func (s *Ssh) rConnect(c net.Conn) (*ssh.Client, error) {
	if s.rClient != nil {
		return s.rClient, nil
	}
	if c == nil {
		return nil, ErrEmptyConnection
	}
	s.rMu.Lock()
	defer s.rMu.Unlock()
	if s.rClient != nil {
		return s.rClient, nil
	}
	cc, chans, reqs, err := ssh.NewClientConn(c, s.addr, s.cfg)
	if err != nil {
		return nil, err
	}
	s.rClient = ssh.NewClient(cc, chans, reqs)
	go func() {
		err = s.rClient.Wait()
		log.Warnln("ssh client wait: %v", err)
		s.rMu.Lock()
		s.rClient = nil
		s.rMu.Unlock()
	}()
	return s.rClient, nil
}

func (s *Ssh) Close() error {
	s.rMu.Lock()
	defer s.rMu.Unlock()
	if s.rClient != nil {
		log.Infoln("ssh closed - " + s.name)
		return s.rClient.Close()
	}
	return nil
}

func NewSsh(option SshOption) (*Ssh, error) {
	cfg := &ssh.ClientConfig{
		User:            option.UserName,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Second * 15,
	}

	if option.Password != "" {
		cfg.Auth = append(cfg.Auth, ssh.Password(option.Password))
	}

	if option.KeyPath != "" {
		buffer, err := ioutil.ReadFile(option.KeyPath)
		if err != nil {
			return nil, fmt.Errorf("read from keyPath '%s' failed", option.KeyPath)
		}

		var signer ssh.Signer
		if option.Passphrase != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase(buffer, []byte(option.Passphrase))
		} else {
			signer, err = ssh.ParsePrivateKey(buffer)
		}
		if err != nil {
			return nil, err
		}
		cfg.Auth = append(cfg.Auth, ssh.PublicKeys(signer))
	}

	return &Ssh{
		Base: &Base{
			name:           option.Name,
			addr:           net.JoinHostPort(option.Server, strconv.Itoa(option.Port)),
			tp:             C.Ssh,
			iface:          option.Interface,
			rmark:          option.RoutingMark,
			originalConfig: &option,
		},
		cfg: cfg,
	}, nil
}

var ErrEmptyConnection = errors.New("empty connection")
