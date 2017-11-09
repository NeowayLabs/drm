package drm

import (
	"unsafe"

	"github.com/NeowayLabs/drm/ioctl"
)

const IOCTLBase = 'd'

var (
	// DRM_IOWR(0x00, struct drm_version)
	IOCTLVersion = ioctl.NewCode(ioctl.Read|ioctl.Write,
		uint16(unsafe.Sizeof(version{})), IOCTLBase, 0)

	// DRM_IOWR(0x0c, struct drm_get_cap)
	IOCTLGetCap = ioctl.NewCode(ioctl.Read|ioctl.Write,
		uint16(unsafe.Sizeof(capability{})), IOCTLBase, 0x0c)
)
