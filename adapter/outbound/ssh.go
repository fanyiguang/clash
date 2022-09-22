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
	"golang.org/x/crypto/ssh"
)

type Ssh struct {
	*Base
	cfg    *ssh.ClientConfig
	client *ssh.Client
	mu     sync.Mutex
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
// relay会调用该方法,传入net.Conn,由于该net.Conn每次都是随机,新建的,无法复用ssh.Client
// TODO: 未关闭该client和connection并每次新建,不知道会不会有其它问题
func (s *Ssh) StreamConn(c net.Conn, metadata *C.Metadata) (net.Conn, error) {
	conn, chans, reqs, err := ssh.NewClientConn(c, s.addr, s.cfg)
	if err != nil {
		return nil, err
	}
	client := ssh.NewClient(conn, chans, reqs)
	return client.Dial("tcp", metadata.RemoteAddress())
}

func (s *Ssh) DialContext(ctx context.Context, metadata *C.Metadata, opts ...dialer.Option) (_ C.Conn, err error) {
	client, err := s.connect(ctx, opts...)
	if err != nil {
		return nil, err
	}
	c, err := client.Dial("tcp", metadata.RemoteAddress())
	if err != nil {
		return nil, err
	}
	return NewConn(c, s), nil
}

func (s *Ssh) connect(ctx context.Context, opts ...dialer.Option) (*ssh.Client, error) {
	if s.client != nil {
		return s.client, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	conn, err := dialer.DialContext(ctx, "tcp", s.addr, opts...)
	if err != nil {
		return nil, err
	}

	c, chans, reqs, err := ssh.NewClientConn(conn, s.addr, s.cfg)
	if err != nil {
		conn.Close()
		return nil, err
	}

	s.client = ssh.NewClient(c, chans, reqs)
	go func() {
		err = s.client.Wait()
		c.Close()

		log.Warnln("ssh client wait: %s", err)
		s.mu.Lock()
		s.client = nil
		s.mu.Unlock()
	}()

	return s.client, nil
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
		cfg: cfg,
	}, nil
}
