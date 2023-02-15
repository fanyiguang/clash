package dns

import (
	"context"
	"errors"
	D "github.com/miekg/dns"
	"net"
	"net/netip"
	"sort"
	"strings"
)

type local struct {
	resolver net.Resolver
}

func newLocal() *local {
	return &local{
		resolver: net.Resolver{},
	}
}

func (l *local) Exchange(m *D.Msg) (msg *D.Msg, err error) {
	var network string
	switch m.Question[0].Qtype {
	case D.TypeA:
		network = "ip4"
	case D.TypeAAAA:
		network = "ip6"
	default:
		return nil, errors.New("unknown Qtype")
	}
	addrs, err := l.resolver.LookupNetIP(context.Background(), network, strings.TrimSuffix(m.Question[0].Name, "."))
	if err != nil {
		return nil, err
	}
	addrs = Map(addrs, func(it netip.Addr) netip.Addr {
		if it.Is4In6() {
			return netip.AddrFrom4(it.As4())
		}
		return it
	})
	switch m.Question[0].Qtype {
	case D.TypeA:
		sort.Slice(addrs, func(i, j int) bool {
			return addrs[i].Is4() && addrs[j].Is6()
		})
	case D.TypeAAAA:
		sort.Slice(addrs, func(i, j int) bool {
			return addrs[i].Is6() && addrs[j].Is4()
		})
	default:
		return nil, errors.New("unknown Qtype")
	}
	return addressesToMessage(addrs, m.Question[0], 300), nil
}

func (l *local) ExchangeContext(ctx context.Context, m *D.Msg) (msg *D.Msg, err error) {
	var network string
	switch m.Question[0].Qtype {
	case D.TypeA:
		network = "ip4"
	case D.TypeAAAA:
		network = "ip6"
	default:
		return nil, errors.New("unknown Qtype")
	}
	addrs, err := l.resolver.LookupNetIP(ctx, network, strings.TrimSuffix(m.Question[0].Name, "."))
	if err != nil {
		return nil, err
	}
	addrs = Map(addrs, func(it netip.Addr) netip.Addr {
		if it.Is4In6() {
			return netip.AddrFrom4(it.As4())
		}
		return it
	})
	switch m.Question[0].Qtype {
	case D.TypeA:
		sort.Slice(addrs, func(i, j int) bool {
			return addrs[i].Is4() && addrs[j].Is6()
		})
	case D.TypeAAAA:
		sort.Slice(addrs, func(i, j int) bool {
			return addrs[i].Is6() && addrs[j].Is4()
		})
	default:
		return nil, errors.New("unknown Qtype")
	}
	return addressesToMessage(addrs, m.Question[0], 300), nil
}

func Map[T any, N any](arr []T, block func(it T) N) []N {
	retArr := make([]N, 0, len(arr))
	for index := range arr {
		retArr = append(retArr, block(arr[index]))
	}
	return retArr
}

func addressesToMessage(addresses []netip.Addr, question D.Question, ttl uint32) *D.Msg {
	response := D.Msg{
		MsgHdr: D.MsgHdr{
			Response: true,
		},
		Question: []D.Question{question},
	}
	if ttl < 0 {
		ttl = 300
	}
	for _, address := range addresses {
		if address.Is4In6() {
			address = netip.AddrFrom4(address.As4())
		}
		if address.Is4() {
			response.Answer = append(response.Answer, &D.A{
				Hdr: D.RR_Header{
					Name:   question.Name,
					Rrtype: D.TypeA,
					Class:  D.ClassINET,
					Ttl:    ttl,
				},
				A: address.AsSlice(),
			})
		} else {
			response.Answer = append(response.Answer, &D.AAAA{
				Hdr: D.RR_Header{
					Name:   question.Name,
					Rrtype: D.TypeAAAA,
					Class:  D.ClassINET,
					Ttl:    ttl,
				},
				AAAA: address.AsSlice(),
			})
		}
	}
	return &response
}
