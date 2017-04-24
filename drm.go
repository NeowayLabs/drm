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

	Version struct {
		version
		Name string
		Date string
		Desc string
	}
)

const (
	Primary = iota
	Control
	Render
)

var (
	IOCTLVersion, _ = ioctl.NewCode(ioctl.Read,
		uint16(unsafe.Sizeof(version{})), 'd', 0)
)

func Available() (Version, error) {
	f, err := openMinor(0, Primary)
	if err != nil {
		// handle backward linux compat?
		// check /proc/dri/0 ?
		return Version{}, err
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
