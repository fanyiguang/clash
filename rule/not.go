package rules

import (
	"fmt"
	"strings"

	C "github.com/Dreamacro/clash/constant"
)

type Not struct {
	rule C.Rule
}

func (n *Not) RuleType() C.RuleType {
	return C.RuleTypeNot
}

func (n *Not) Match(metadata *C.Metadata) bool {
	return !n.rule.Match(metadata)
}

func (n *Not) Adapter() string {
	return n.rule.Adapter()
}

func (n *Not) Payload() string {
	return n.rule.Payload()
}

func (n *Not) ShouldResolveIP() bool {
	return n.rule.ShouldResolveIP()
}

func (n *Not) ShouldFindProcess() bool {
	return n.rule.ShouldFindProcess()
}

func NewNot(payload string, adapter string, params []string) (*Not, error) {
	payloads := strings.SplitN(payload, ":", 1)
	if len(payload) != 2 {
		return nil, fmt.Errorf("invalid NOT rule payload: %s", payload)
	}
	r, err := ParseRule(C.RuleType(0).FormatString(payloads[0]), payloads[1], adapter, params)
	if err != nil {
		return nil, fmt.Errorf("parse NOT rule failed: %v", err)
	}
	return &Not{rule: r}, nil
}
