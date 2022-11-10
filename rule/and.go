package rules

import (
	"fmt"
	"strings"

	C "github.com/Dreamacro/clash/constant"
	"github.com/Dreamacro/clash/log"
)

type And struct {
	rules             []C.Rule
	adapter           string
	payload           string
	shouldResolveIP   bool
	shouldFindProcess bool
}

func (d *And) RuleType() C.RuleType {
	return C.RuleTypeAnd
}

func (d *And) Match(metadata *C.Metadata) bool {
	for _, rule := range d.rules {
		if !rule.Match(metadata) {
			return false
		}
	}
	return true
}

func (d *And) Adapter() string {
	return d.adapter
}

func (d *And) Payload() string {
	return d.payload
}

func (d *And) ShouldResolveIP() bool {
	return d.shouldResolveIP
}

func (d *And) ShouldFindProcess() bool {
	return d.shouldFindProcess
}

func NewAnd(payload string, adapter string) (*And, error) {
	and := &And{
		payload: payload,
		adapter: adapter,
	}
	rules := strings.Split(payload, "&&")
	for _, rule := range rules {
		r, err := parseAndRules(rule)
		if err != nil {
			return nil, err
		}
		if r.ShouldFindProcess() {
			and.shouldFindProcess = true
		}
		if r.ShouldResolveIP() {
			and.shouldResolveIP = true
		}
		and.rules = append(and.rules, r)
		log.Infoln("%s , %s, %s", r.RuleType().String(), r.Payload(), r.Adapter())
	}
	return and, nil
}

func parseAndRules(str string) (C.Rule, error) {
	rule := trimArr(strings.Split(str, ":"))
	var (
		payload string
		params  []string
	)

	switch l := len(rule); {
	case l == 2:
		payload = rule[1]
	case l >= 3:
		payload = rule[1]
		params = rule[2:]
	default:
		return nil, fmt.Errorf("rules [%s] error: format invalid", str)
	}

	rule = trimArr(rule)
	params = trimArr(params)
	return ParseRule(C.RuleType(0).FormatString(rule[0]), payload, "", params)
}

func trimArr(arr []string) (r []string) {
	for _, e := range arr {
		r = append(r, strings.Trim(e, " "))
	}
	return
}
