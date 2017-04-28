package drm

import (
	"unsafe"

	"github.com/tiago4orion/drm/ioctl"
)

const base = 'd'

var (
	// DRM_IOR(0x0, struct drm_version)
	IOCTLVersion = ioctl.NewCode(ioctl.Read,
		uint16(unsafe.Sizeof(version{})), base, 0)

	// DRM_IOWR(0x0c, struct drm_get_cap)
	IOCTLGetCap = ioctl.NewCode(ioctl.Read|ioctl.Write,
		uint16(unsafe.Sizeof(capability{})), base, 0x0c)

	// DRM_IOWR(0xA0, struct drm_mode_card_res)
	IOCTLModeResources = ioctl.NewCode(ioctl.Read|ioctl.Write,
		uint16(unsafe.Sizeof(modeRes{})), base, 0xA0)
)
