package service

import (
	"hash/fnv"
	"strings"
)

const (
	GrayscaleModeWhitelist  = "whitelist"
	GrayscaleModePercentage = "percentage"
	GrayscaleModeAll        = "all"
)

// GrayscaleConfig holds grayscale switch configuration.
type GrayscaleConfig struct {
	Enabled    bool
	Mode       string
	Whitelist  []string
	Percentage int
}

// GrayscaleService decides whether a tracking number should be processed by Go.
type GrayscaleService struct {
	config GrayscaleConfig
}

func NewGrayscaleService(config GrayscaleConfig) *GrayscaleService {
	return &GrayscaleService{config: config}
}

// ShouldProcessByGo returns true if this tracking number should be handled by Go.
func (s *GrayscaleService) ShouldProcessByGo(trackingNumber string) bool {
	if !s.config.Enabled {
		return false
	}

	switch s.config.Mode {
	case GrayscaleModeWhitelist:
		return s.isInWhitelist(trackingNumber)
	case GrayscaleModePercentage:
		return s.isInPercentage(trackingNumber)
	case GrayscaleModeAll:
		return true
	default:
		return false
	}
}

func (s *GrayscaleService) isInWhitelist(trackingNumber string) bool {
	for _, id := range s.config.Whitelist {
		if strings.TrimSpace(id) == trackingNumber {
			return true
		}
	}
	return false
}

func (s *GrayscaleService) isInPercentage(trackingNumber string) bool {
	if s.config.Percentage <= 0 {
		return false
	}
	if s.config.Percentage >= 100 {
		return true
	}
	h := fnv.New32a()
	h.Write([]byte(trackingNumber))
	return int(h.Sum32()%100) < s.config.Percentage
}

// WhitelistSize returns the number of entries in the whitelist.
func (s *GrayscaleService) WhitelistSize() int {
	return len(s.config.Whitelist)
}
