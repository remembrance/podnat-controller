package firewall

import "github.com/remembrance/podnat-controller/internal/api"

type Processor interface {
	Apply(event *api.PodInfo) error
}
