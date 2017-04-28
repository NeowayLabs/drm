package drm

import (
	"os"
	"unsafe"

	"github.com/tiago4orion/drm/ioctl"
)

type (
	capability struct {
		cap uint64
		val uint64
	}
)

const (
	CapDumbBuffer = iota + 1
	CapVBlankHighCRTC
	CapDumbPreferredDepth
	CapDumbPreferShadow
	CapPrime
	CapTimestampMonotonic
	CapAsyncPageFlip

	CapAddFB2Modifiers = 0x10
)

func HasDumbBuffer(file *os.File) bool {
	cap := &capability{}
	cap.cap = CapDumbBuffer
	err := ioctl.Do(uintptr(file.Fd()), uintptr(IOCTLGetCap), uintptr(unsafe.Pointer(cap)))
	if err != nil {
		return false
	}
	return cap.val != 0
}
