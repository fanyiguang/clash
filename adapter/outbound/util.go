package outbound

import (
	"bytes"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/Dreamacro/clash/common/buf"
	"github.com/Dreamacro/clash/common/bufio"
	"github.com/Dreamacro/clash/component/resolver"
	C "github.com/Dreamacro/clash/constant"
	"github.com/Dreamacro/clash/transport/socks5"
)

func tcpKeepAlive(c net.Conn) {
	if tcp, ok := c.(*net.TCPConn); ok {
		tcp.SetKeepAlive(true)
		tcp.SetKeepAlivePeriod(30 * time.Second)
	}
}

func serializesSocksAddr(metadata *C.Metadata) []byte {
	var buf [][]byte
	addrType := metadata.AddrType()
	aType := uint8(addrType)
	p, _ := strconv.ParseUint(metadata.DstPort, 10, 16)
	port := []byte{uint8(p >> 8), uint8(p & 0xff)}
	switch addrType {
	case socks5.AtypDomainName:
		len := uint8(len(metadata.Host))
		host := []byte(metadata.Host)
		buf = [][]byte{{aType, len}, host, port}
	case socks5.AtypIPv4:
		host := metadata.DstIP.To4()
		buf = [][]byte{{aType}, host, port}
	case socks5.AtypIPv6:
		host := metadata.DstIP.To16()
		buf = [][]byte{{aType}, host, port}
	}
	return bytes.Join(buf, nil)
}

func resolveUDPAddr(network, address string) (*net.UDPAddr, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}

	ip, err := resolver.ResolveIP(host)
	if err != nil {
		return nil, err
	}
	return net.ResolveUDPAddr(network, net.JoinHostPort(ip.String(), port))
}

func safeConnClose(c net.Conn, err error) {
	if err != nil {
		c.Close()
	}
}

// refer to https://pkg.go.dev/net/http@master#pkg-constants
var methods = [...]string{"get", "post", "head", "put", "delete", "options", "connect", "patch", "trace"}

var (
	snifferReadTimeOut = 300 * time.Millisecond
	// bufferSize         = 16 * 1024
	bufferSize = 8
)

func beginWithHTTPMethod(b []byte) bool {
	for _, m := range &methods {
		if len(b) >= len(m) && strings.EqualFold(string(b[:len(m)]), m) {
			return true
		}
	}

	return false
}

func SniffHTTP(b []byte) bool {
	return beginWithHTTPMethod(b)
}

func SniffHTTPFromConn(c net.Conn) (bool, net.Conn) {
	buffer := buf.NewSize(bufferSize)
	buffer.FullReset()

	// 读取 300ms
	err := c.SetReadDeadline(time.Now().Add(snifferReadTimeOut))
	defer c.SetReadDeadline(time.Time{})
	if err != nil {
		return false, c
	}

	var isHttp bool
	_, err = buffer.ReadOnceFrom(c)

	if err != nil {
		isHttp = false
	}

	isHttp = SniffHTTP(buffer.Bytes())

	if !buffer.IsEmpty() {
		c = bufio.NewCachedConn(c, buffer)
	} else {
		buffer.Release()
	}
	return isHttp, c
}
