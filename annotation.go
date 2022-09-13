package main

import (
	"encoding/json"
	"errors"
	"fmt"
)

type PodNatAnnotation struct {
	PublicInterface bool   `json:"pubif"`
	SourcePort      int    `json:"src"`
	DestinationPort int    `json:"dst"`
	Protocol        string `json:"proto"`
}

func parseAnnotation(data string) (*PodNatAnnotation, error) {

	// default values can be overridden by unmarshaling
	pa := PodNatAnnotation{
		PublicInterface: true,
		Protocol: "tcp",
	}

	err := json.Unmarshal([]byte(data), &pa)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error unmarshaling data into annotation json format: %v", data))
	}

	// sanity checks for data
	if pa.SourcePort == 0 || pa.DestinationPort == 0 {
		return nil, errors.New("port 0 is reserved and cannot be assigned")
	}

	if pa.SourcePort >= 65535 || pa.DestinationPort >= 65535 {
		return nil, errors.New("port values must be lower than 65536")
	}

	if pa.Protocol != "tcp" && pa.Protocol != "udp" {
		return nil, errors.New("supported protocols for NAT entries are 'tcp' and 'udp'")
	}

	return &pa, nil
}