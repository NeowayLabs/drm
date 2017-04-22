package drm

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

//
// Driver version information.
//
// drm.GetVersion() and drm.SetVersion().
// dowsn't works.. requires cgo
// type Version = C.struct_drmversion
// :(
type Version struct {
	Major      int            // Major version
	Minor      int            // Minor version
	PatchLevel int            // Patch level
	NameLen    int            // Length of name buffer
	Name       unsafe.Pointer // Name of driver
	DateLen    int            // Length of date buffer
	date       unsafe.Pointer // User-space buffer to hold date
	DescLen    int            // Length of desc buffer
	Desc       unsafe.Pointer // User-space buffer to hold desc
}

const (
	Primary = iota
	Control
	Render
)

const (
	_IOC_NONE  = 0x0
	_IOC_WRITE = 0x1
	_IOC_READ  = 0x2

	_IOC_NRBITS   = 8
	_IOC_TYPEBITS = 8
	_IOC_SIZEBITS = 14
	_IOC_DIRBITS  = 2
	_IOC_NRSHIFT  = 0

	_IOC_TYPESHIFT = _IOC_NRSHIFT + _IOC_NRBITS
	_IOC_SIZESHIFT = _IOC_TYPESHIFT + _IOC_TYPEBITS
	_IOC_DIRSHIFT  = _IOC_SIZESHIFT + _IOC_SIZEBITS
)

var (
	IOCTLVersion = uintptr(_IOWR('d', 0x00, int(unsafe.Sizeof(Version{}))))
)

func _IOC(dir int, t int, nr int, size int) int {
	return (dir << _IOC_DIRSHIFT) | (t << _IOC_TYPESHIFT) |
		(nr << _IOC_NRSHIFT) | (size << _IOC_SIZESHIFT)
}

func _IOR(t int, nr int, size int) int {
	return _IOC(_IOC_READ, t, nr, size)
}

func _IOWR(t int, nr int, size int) int {
	return _IOC(_IOC_READ|_IOC_WRITE, t, nr, size)
}

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
	var version Version
	err := ioctl(uintptr(file.Fd()), IOCTLVersion, uintptr(unsafe.Pointer(&version)))
	if err != nil {
		return nil, err
	}
	return &version, nil
}

func ioctl(fd, cmd, ptr uintptr) error {
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd, cmd, ptr)
	if err != 0 {
		return err
	}
	return nil
}
