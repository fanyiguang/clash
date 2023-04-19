package rules

import (
	"path/filepath"
	"strings"

	C "github.com/Dreamacro/clash/constant"
)

type NotProcess struct {
	adapter   string
	payload   string
	processes []string
	nameOnly  bool
}

func (ps *NotProcess) RuleType() C.RuleType {
	if ps.nameOnly {
		return C.RuleTypeProcess
	}

	return C.RuleTypeProcessPath
}

func (ps *NotProcess) Match(metadata *C.Metadata) bool {
	if ps.nameOnly {
		for _, proc := range ps.processes {
			if strings.EqualFold(filepath.Base(metadata.ProcessPath), proc) {
				return false
			}
		}
		return true
	}
	for _, proc := range ps.processes {
		if strings.EqualFold(metadata.ProcessPath, proc) {
			return false
		}
	}
	return true
}

func (ps *NotProcess) Adapter() string {
	return ps.adapter
}

func (ps *NotProcess) Payload() string {
	return ps.payload
}

func (ps *NotProcess) ShouldResolveIP() bool {
	return false
}

func (ps *NotProcess) ShouldFindProcess() bool {
	return true
}

func NewNotProcess(process string, adapter string, nameOnly bool) (*NotProcess, error) {
	processes := strings.Split(process, "|")
	return &NotProcess{
		adapter:   adapter,
		processes: processes,
		payload:   process,
		nameOnly:  nameOnly,
	}, nil
}
