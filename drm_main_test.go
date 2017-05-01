package drm_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/tiago4orion/drm"
)

type (
	cardDetail struct {
		version      drm.Version
		capabilities map[uint64]uint64
	}
)

var (
	card, errCard = drm.Available()
	cards         = map[string]cardDetail{
		"i915": cardDetail{
			version: drm.Version{
				Major: 1,
				Minor: 6,
				Patch: 1,
				Name:  "i915",
				Desc:  "i915",
				Date:  "20160425",
			},
			capabilities: map[uint64]uint64{
				drm.CapDumbBuffer:         1,
				drm.CapVBlankHighCRTC:     1,
				drm.CapDumbPreferredDepth: 24,
				drm.CapDumbPreferShadow:   1,
				drm.CapPrime:              3,
				drm.CapTimestampMonotonic: 1,
				drm.CapAsyncPageFlip:      0,
				drm.CapCursorWidth:        256,
				drm.CapCursorHeight:       256,

				drm.CapAddFB2Modifiers: 1,
			},
		},
	}
	cardInfo cardDetail
)

func TestMain(m *testing.M) {
	if errCard != nil {
		fmt.Fprintf(os.Stderr, "No graphics card available to test")
		os.Exit(1)
	}
	if _, ok := cards[card.Name]; !ok {
		fmt.Fprintf(os.Stderr, "No tests for card %s", card.Name)
		os.Exit(1)
	}
	cardInfo = cards[card.Name]
	os.Exit(m.Run())
}
