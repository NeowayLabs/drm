package drm_test

import (
	"fmt"

	"github.com/NeowayLabs/drm"
)

func ExampleHasDumbBuffer() {
	// This example shows how to test if your graphics card
	// supports 'dumb buffers' capability. With this capability
	// you can create simple memory-mapped buffers without any
	// driver-dependent code.

	file, err := drm.OpenCard(0)
	if err != nil {
		fmt.Printf("error: %s", err.Error())
		return
	}
	defer file.Close()
	if !drm.HasDumbBuffer(file) {
		fmt.Printf("drm device does not support dumb buffers")
		return
	}
	fmt.Printf("ok")

	// Output: ok
}
