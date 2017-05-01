package drm

import (
	"os"
	"unsafe"

	"github.com/tiago4orion/drm/ioctl"
)

type (
	capability struct {
		id  uint64
		val uint64
	}
)

const (
	CapDumbBuffer uint64 = iota + 1
	CapVBlankHighCRTC
	CapDumbPreferredDepth
	CapDumbPreferShadow
	CapPrime
	CapTimestampMonotonic
	CapAsyncPageFlip
	CapCursorWidth
	CapCursorHeight

	CapAddFB2Modifiers = 0x10
)

func HasDumbBuffer(file *os.File) bool {
	cap, err := GetCap(file, CapDumbBuffer)
	if err != nil {
		return false
	}
	return cap != 0
}

func GetCap(file *os.File, capid uint64) (uint64, error) {
	cap := &capability{}
	cap.id = capid
	err := ioctl.Do(uintptr(file.Fd()), uintptr(IOCTLGetCap), uintptr(unsafe.Pointer(cap)))
	if err != nil {
		return 0, err
	}
	return cap.val, nil
}
