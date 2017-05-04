package drm

import (
	"bytes"
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

	// Version of DRM driver
	Version struct {
		version

		Major, Minor, Patch int32
		Name                string // Name of the driver (eg.: i915)
		Date                string
		Desc                string
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
	var (
		name, date, desc []byte
	)

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

	nozero := func(r rune) bool {
		if r == 0 {
			return true
		} else {
			return false
		}
	}

	v := Version{
		version: *version,
		Major:   version.Major,
		Minor:   version.Minor,
		Patch:   version.Patch,
		Name:    string(bytes.TrimFunc(name, nozero)),
		Date:    string(bytes.TrimFunc(date, nozero)),
		Desc:    string(bytes.TrimFunc(desc, nozero)),
	}

	return v, nil
}
