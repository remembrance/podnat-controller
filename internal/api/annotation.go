package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gutmensch/podnat-controller/internal/common"
	"strings"

	"golang.org/x/exp/slices"
)

// inject default values with custom unmarshaler
func (c *NATDefinition) UnmarshalJSON(data []byte) error {
	pd := &struct {
		InterfaceAutoDetect bool    `json:"ifaceAuto"`
		SourceIP            *string `json:"srcIP"`
		SourcePort          uint16  `json:"srcPort"`
		DestinationPort     uint16  `json:"dstPort"`
		Protocol            string  `json:"proto"`
	}{
		InterfaceAutoDetect: true,
		SourceIP:            nil,
		Protocol:            "tcp",
	}
	if err := json.Unmarshal(data, pd); err != nil {
		return err
	}
	c.InterfaceAutoDetect = pd.InterfaceAutoDetect
	c.SourceIP = pd.SourceIP
	c.SourcePort = pd.SourcePort
	c.DestinationPort = pd.DestinationPort
	c.Protocol = pd.Protocol

	return nil
}

func ParseAnnotation(data string) (*PodNATAnnotation, error) {
	pa := &PodNATAnnotation{}

	err := json.Unmarshal([]byte(data), pa)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error unmarshaling data into annotation json format: %v", data))
	}

	// sanity checks for IPs and ports
	for _, def := range pa.TableEntries {
		if def.SourceIP == nil && def.InterfaceAutoDetect == false {
			return nil, errors.New("need either InterfaceAutoDetect enabled or provided SourceIP for entry")
		}
		if def.SourceIP != nil && def.InterfaceAutoDetect == true {
			return nil, errors.New("SourceIP provided but InterfaceAutoDetect still enabled, please disable")
		}

		if def.SourcePort == 0 || def.DestinationPort == 0 {
			return nil, errors.New("port 0 is reserved and cannot be used")
		}

		if def.Protocol != "tcp" && def.Protocol != "udp" {
			return nil, errors.New("supported protocols for NAT entries are 'tcp' and 'udp'")
		}

		if common.RestrictedPorts == "" {
			continue
		}

		_restrictedPorts, _ := common.SliceAtoi(strings.Split(common.RestrictedPorts, ","))
		if slices.Contains(_restrictedPorts, def.SourcePort) || slices.Contains(_restrictedPorts, def.DestinationPort) {
			return nil, errors.New(
				fmt.Sprintf(
					"restricted ports %v are not allowed by default",
					common.RestrictedPorts,
				),
			)
		}
	}

	return pa, nil
}
