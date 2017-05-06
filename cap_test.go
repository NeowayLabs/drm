package drm_test

import (
	"testing"

	"github.com/NeowayLabs/drm"
)

func TestHasDumbBuffer(t *testing.T) {
	file, err := drm.OpenCard(0)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	version, err := drm.GetVersion(file)
	if err != nil {
		t.Error(err)
		return
	}
	if hasDumb := drm.HasDumbBuffer(file); hasDumb != (cardInfo.capabilities[drm.CapDumbBuffer] != 0) {
		t.Errorf("Card '%s' should support dumb buffers...Got %v but %d", version.Name, hasDumb, cardInfo.capabilities[drm.CapDumbBuffer])
		return
	}
}

func TestGetCap(t *testing.T) {
	file, err := drm.OpenCard(0)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	for cap, capval := range cardInfo.capabilities {
		ccap, err := drm.GetCap(file, cap)
		if err != nil {
			t.Error(err)
			return
		}
		if ccap != capval {
			t.Errorf("Capability %d differs: %d != %d", cap, ccap, capval)
			return
		}

	}
}
