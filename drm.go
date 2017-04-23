package drm

import (
	"fmt"
	"os"
	"unsafe"

	"github.com/tiago4orion/drm/ioctl"
)

//
// Driver version information.
//
// drm.GetVersion() and drm.SetVersion().
type Version struct {
	Major   int32
	Minor   int32
	Patch   int32
	Namelen int64
	Name    unsafe.Pointer
	Datelen int64
	Date    unsafe.Pointer
	Desclen int64
	Desc    unsafe.Pointer
}

const (
	Primary = iota
	Control
	Render
)

var (
	v               Version
	IOCTLVersion, _ = ioctl.NewCode(ioctl.Read, uint16(unsafe.Sizeof(v)), 'd', 0)
)

func Available() (*Version, error) {
	f, err := openMinor(0, Primary)
	if err != nil {
		// handle backward linux compat?
		// check /proc/dri/0 ?
		return nil, err
	}
	defer f.Close()
	return GetVersion(f)
}

func openMinor(minor int, typ int) (*os.File, error) {
	var (
		devname string
		devfmt  string
	)

	switch typ {
	case Primary:
		devfmt = "%s/card%d"
	case Control:
		devfmt = "%s/controlD%d"
	case Render:
		devfmt = "%s/renderD%d"
	default:
		return nil, fmt.Errorf("invalid DRM type: %d", typ)
	}

	devname = fmt.Sprintf(devfmt, "/dev/dri", minor)
	return os.OpenFile(devname, os.O_RDWR, 0)
}

func GetVersion(file *os.File) (*Version, error) {
	version := &Version{}

	err := ioctl.Do(uintptr(file.Fd()), uintptr(IOCTLVersion),
		uintptr(unsafe.Pointer(version)))
	if err != nil {
		return nil, err
	}
	if version.Namelen > 0 {
		name := make([]byte, version.Namelen+1)
		version.Name = unsafe.Pointer(&name[0])
	}
	if version.Datelen > 0 {
		var date []byte = make([]byte, version.Datelen+1)

		version.Date = unsafe.Pointer(&date[0])
	}
	if version.Desclen > 0 {
		var desc []byte = make([]byte, version.Desclen+1)
		version.Desc = unsafe.Pointer(&desc[0])
	}
	err = ioctl.Do(uintptr(file.Fd()), uintptr(IOCTLVersion),
		uintptr(unsafe.Pointer(version)))
	if err != nil {
		return nil, err
	}

	return version, nil
}
