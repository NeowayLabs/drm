package drm

import "testing"

func TestDRIOpen(t *testing.T) {
	file, err := openMinor(0, Primary)
	if err != nil {
		t.Fatal(err)
	}
	file.Close()
}

func TestAvailable(t *testing.T) {
	v, err := Available()
	if err != nil {
		t.Fatal(err)
	}
	if v.Major == 0 && v.Minor == 0 && v.Patch == 0 {
		t.Fatalf("Doesn't got driver version: %d.%d.%d",
			v.Major, v.Minor, v.Patch)
	}

	t.Logf("Driver name: %s", v.Name)
	t.Logf("Driver version: %d.%d.%d", v.Major, v.Minor, v.Patch)
	t.Logf("Driver date: %s", v.Date)
	t.Logf("Driver description: %s", v.Desc)
}
