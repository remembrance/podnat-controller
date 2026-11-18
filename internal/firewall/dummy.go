/*
  no-op template for a firewall processor
  use this as example for nftables, etc.
*/

package firewall

import (
	"github.com/remembrance/podnat-controller/internal/api"
	"github.com/remembrance/podnat-controller/internal/common"
	"k8s.io/klog/v2"
)

type DummyProcessor struct{}

func NewDummyProcessor() *DummyProcessor {
	klog.Warningf("firewall flavor '%s' not implemented, please use a supported firewall", common.FirewallFlavor)
	proc := &DummyProcessor{}
	return proc
}

func (p *DummyProcessor) Apply(event *api.PodInfo) error {
	klog.Warningf("firewall flavor '%s' not implemented, please use a supported firewall", common.FirewallFlavor)
	return nil
}
