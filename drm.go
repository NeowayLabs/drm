package drm

import (
	"fmt"
	"os"
	"unsafe"

	"github.com/tiago4orion/drm/ioctl"
)

type (
	version struct {
		Major   int32
		Minor   int32
		Patch   int32
		namelen int64
		name    uintptr
		datelen int64
		date    uintptr
		desclen int64
		desc    uintptr
	}

	modeRes struct {
		fbIdPtr              uintptr
		crtcIdPtr            uintptr
		connectorIdPtr       uintptr
		encoderIdPtr         uintptr
		CountFbs             uint32
		CountCrtcs           uint32
		CountConnectors      uint32
		CountEncoders        uint32
		MinWidth, MaxWidth   uint32
		MinHeight, MaxHeight uint32
	}

	// Version of DRM driver
	Version struct {
		version
		Name string // Name of the driver (eg.: i915)
		Date string
		Desc string
	}

	ModeRes struct {
		modeRes
		Fbs        []uint32
		Crtcs      []uint32
		Connectors []uint32
		Encoders   []uint32
	}
)

const (
	driPath = "/dev/dri"
)

func Available() (Version, error) {
	f, err := OpenCard(0)
	if err != nil {
		// handle backward linux compat?
		// check /proc/dri/0 ?
		return Version{}, err
	}
	defer f.Close()
	return GetVersion(f)
}

func OpenCard(n int) (*os.File, error) {
	return open(fmt.Sprintf("%s/card%d", driPath, n))
}

func OpenControlDev(n int) (*os.File, error) {
	return open(fmt.Sprintf("%s/controlD%d", driPath, n))
}

func OpenRenderDev(n int) (*os.File, error) {
	return open(fmt.Sprintf("%s/renderD%d", driPath, n))
}

func open(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDWR, 0)
}

func GetVersion(file *os.File) (Version, error) {
	var date []byte
	var name []byte
	var desc []byte

	version := &version{}

	err := ioctl.Do(uintptr(file.Fd()), uintptr(IOCTLVersion),
		uintptr(unsafe.Pointer(version)))
	if err != nil {
		return Version{}, err
	}
	if version.namelen > 0 {
		name = make([]byte, version.namelen+1)
		version.name = uintptr(unsafe.Pointer(&name[0]))
	}
	if version.datelen > 0 {
		date = make([]byte, version.datelen+1)
		version.date = uintptr(unsafe.Pointer(&date[0]))
	}
	if version.desclen > 0 {
		desc = make([]byte, version.desclen+1)
		version.desc = uintptr(unsafe.Pointer(&desc[0]))
	}
	err = ioctl.Do(uintptr(file.Fd()), uintptr(IOCTLVersion),
		uintptr(unsafe.Pointer(version)))
	if err != nil {
		return Version{}, err
	}

	// remove C null byte at end
	name = name[:version.namelen]
	date = date[:version.datelen]
	desc = desc[:version.desclen]

	v := Version{
		version: *version,
		Name:    string(name),
		Date:    string(date),
		Desc:    string(desc),
	}

	return v, nil
}

func ModeResources(file *os.File) (*ModeRes, error) {
	mres := &modeRes{}
	err := ioctl.Do(uintptr(file.Fd()), uintptr(IOCTLModeResources),
		uintptr(unsafe.Pointer(mres)))
	if err != nil {
		return nil, err
	}

	var (
		fbids, crtcids, connectorids, encoderids []uint32
	)

	if mres.CountFbs > 0 {
		fbids = make([]uint32, mres.CountFbs)
		mres.fbIdPtr = uintptr(unsafe.Pointer(&fbids[0]))
	}
	if mres.CountCrtcs > 0 {
		crtcids = make([]uint32, mres.CountCrtcs)
		mres.crtcIdPtr = uintptr(unsafe.Pointer(&crtcids[0]))
	}
	if mres.CountEncoders > 0 {
		encoderids = make([]uint32, mres.CountEncoders)
		mres.encoderIdPtr = uintptr(unsafe.Pointer(&encoderids[0]))
	}
	if mres.CountConnectors > 0 {
		connectorids = make([]uint32, mres.CountConnectors)
		mres.connectorIdPtr = uintptr(unsafe.Pointer(&connectorids[0]))
	}

	err = ioctl.Do(uintptr(file.Fd()), uintptr(IOCTLModeResources),
		uintptr(unsafe.Pointer(mres)))
	if err != nil {
		return nil, err
	}

	// TODO(i4k): handle hotplugging in-between the ioctls above

	return &ModeRes{
		modeRes:    *mres,
		Fbs:        fbids,
		Crtcs:      crtcids,
		Encoders:   encoderids,
		Connectors: connectorids,
	}, nil
}
