package drm

import (
	"fmt"
	"testing"
)

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
	fmt.Printf("%#v\n", v)
}
