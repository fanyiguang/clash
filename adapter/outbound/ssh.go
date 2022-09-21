package outbound

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/Dreamacro/clash/component/dialer"
	C "github.com/Dreamacro/clash/constant"
	"github.com/Dreamacro/clash/log"
	"go.uber.org/atomic"
	"golang.org/x/crypto/ssh"
)

type Ssh struct {
	*Base
	client    *ssh.Client
	cfg       *ssh.ClientConfig
	connected *atomic.Bool
	mu        sync.Mutex
}

type SshOption struct {
	BasicOption
	Name       string `proxy:"name" json:"name"`
	Server     string `proxy:"server" json:"server"`
	Port       int    `proxy:"port" json:"port"`
	UserName   string `proxy:"username" json:"username"`
	Password   string `proxy:"password,omitempty" json:"password,omitempty"`     // 密码
	KeyPath    string `proxy:"key_path,omitempty" json:"key_path,omitempty"`     // 私钥地址
	Passphrase string `proxy:"passphrase,omitempty" json:"passphrase,omitempty"` // 私钥密码
}

// StreamConn implements C.ProxyAdapter
func (s *Ssh) StreamConn(c net.Conn, metadata *C.Metadata) (net.Conn, error) {
	if !s.connected.Load() {
		s.mu.Lock()
		if !s.connected.Load() {
			conn, chans, reqs, err := ssh.NewClientConn(c, s.addr, s.cfg)
			if err != nil {
				s.mu.Unlock()
				return nil, err
			}
			s.client = ssh.NewClient(conn, chans, reqs)
			s.connected.Store(true)
			go func() {
				err = s.client.Wait()
				s.connected.Store(false)
				log.Errorln("ssh client wait: %s", err)
			}()
		}
		s.mu.Unlock()
	}

	return s.client.Dial("tcp", metadata.RemoteAddress())
}

func (s *Ssh) DialContext(ctx context.Context, metadata *C.Metadata, opts ...dialer.Option) (_ C.Conn, err error) {
	c, err := dialer.DialContext(ctx, "tcp", s.addr, opts...)
	if err != nil {
		return nil, err
	}
	c, err = s.StreamConn(c, metadata)
	if err != nil {
		return nil, err
	}
	return NewConn(c, s), nil
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
			name:  option.Name,
			addr:  net.JoinHostPort(option.Server, strconv.Itoa(option.Port)),
			tp:    C.Ssh,
			iface: option.Interface,
			rmark: option.RoutingMark,
		},
		cfg:       cfg,
		connected: atomic.NewBool(false),
	}, nil
}
