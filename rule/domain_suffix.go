package rules

import (
	"strings"

	C "github.com/Dreamacro/clash/constant"
)

type DomainSuffix struct {
	suffix  string
	adapter string
}

func (ds *DomainSuffix) RuleType() C.RuleType {
	return C.RuleTypeDomainSuffix
}

func (ds *DomainSuffix) Match(metadata *C.Metadata) bool {
	domain := metadata.Host
	return strings.HasSuffix(domain, "."+ds.suffix) || domain == ds.suffix
}

func (ds *DomainSuffix) Adapter() string {
	return ds.adapter
}

func (ds *DomainSuffix) Payload() string {
	return ds.suffix
}

func (ds *DomainSuffix) ShouldResolveIP() bool {
	return false
}

func (ds *DomainSuffix) ShouldFindProcess() bool {
	return false
}

func NewDomainSuffix(suffix string, adapter string) *DomainSuffix {
	return &DomainSuffix{
		suffix:  strings.ToLower(suffix),
		adapter: adapter,
	}
}
