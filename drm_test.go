package drm_test

import (
	"testing"

	"github.com/NeowayLabs/drm"
	"github.com/NeowayLabs/drm/mode"
)

func TestDRIOpen(t *testing.T) {
	file, err := drm.OpenCard(0)
	if err != nil {
		t.Fatal(err)
	}
	file.Close()
}

func TestAvailableCard(t *testing.T) {
	v, err := drm.Available()
	if err != nil {
		t.Fatal(err)
	}
	if v.Major == 0 && v.Minor == 0 && v.Patch == 0 {
		t.Fatalf("failed to get driver version: %#v", v)
	}
	if v.Major != cardInfo.version.Major && v.Minor != cardInfo.version.Minor &&
		v.Patch != cardInfo.version.Patch {
		t.Logf("Unknow driver version: %d.%d.%d", v.Major, v.Minor, v.Patch)
	}

	t.Logf("Driver name: %s", v.Name)
	t.Logf("Driver version: %d.%d.%d", v.Major, v.Minor, v.Patch)
	t.Logf("Driver date: %s", v.Date)
	t.Logf("Driver description: %s", v.Desc)
}

func TestModeRes(t *testing.T) {
	file, err := drm.OpenCard(0)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	mres, err := mode.GetResources(file)
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Number of framebuffers: %d", mres.CountFbs)
	t.Logf("Number of CRTCs: %d", mres.CountCrtcs)
	t.Logf("Number of connectors: %d", mres.CountConnectors)
	t.Logf("Number of encoders: %d", mres.CountEncoders)
	t.Logf("Framebuffers ids: %v", mres.Fbs)
	t.Logf("CRTC ids: %v", mres.Crtcs)
	t.Logf("Connector ids: %v", mres.Connectors)
	t.Logf("Encoder ids: %v", mres.Encoders)
}
