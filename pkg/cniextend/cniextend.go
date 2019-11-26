package cniextend

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/containernetworking/cni/pkg/invoke"
	"github.com/containernetworking/cni/pkg/types/current"
	"os"
	"path/filepath"
)

type VlanResult struct {
	current.Result
	VlanNum int `json:"vlan,omitempty"`
}

func (r *VlanResult) Vlan() int {
	return r.VlanNum
}

func (r *VlanResult) ConvertToResult() (current.Result, error) {
	// version convert not implemented
	return r.Result, nil
}

func (r *VlanResult) Print() error {
	data, err := json.MarshalIndent(r, "", "    ")
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(data)
	return err
}

func ExecAdd(plugin string, netconf []byte) (*VlanResult, error) {
	return DelegateAdd(plugin, netconf)
}

// DelegateAdd calls the given delegate plugin with the CNI ADD action and
// JSON configuration
func DelegateAdd(delegatePlugin string, netconf []byte) (*VlanResult, error) {
	exec := &invoke.DefaultExec{
		RawExec: &invoke.RawExec{Stderr: os.Stderr},
	}

	if os.Getenv("CNI_COMMAND") != "ADD" {
		return nil, fmt.Errorf("CNI_COMMAND is not ADD")
	}

	paths := filepath.SplitList(os.Getenv("CNI_PATH"))

	pluginPath, err := exec.FindInPath(delegatePlugin, paths)
	if err != nil {
		return nil, err
	}

	return ExecPluginWithVlanResult(pluginPath, netconf, invoke.ArgsFromEnv(), exec)
}

func ExecPluginWithVlanResult(pluginPath string, netconf []byte, args invoke.CNIArgs, exec invoke.Exec) (*VlanResult, error) {
	ctx := context.TODO()
	stdoutBytes, err := exec.ExecPlugin(ctx, pluginPath, netconf, args.AsEnv())
	if err != nil {
		return nil, err
	}
	vlanResult := &VlanResult{}
	err = json.Unmarshal(stdoutBytes, vlanResult)
	if err != nil {
		return nil, err
	}
	return vlanResult, nil
}
