package drm_test

import (
	"testing"

	"github.com/tiago4orion/drm"
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
	if drm.HasDumbBuffer(file) != cardInfo.capabilities[drm.CapDumbBuffer] {
		t.Errorf("Card '%s' should support dumb buffers...", version.Name)
		return
	}
}
