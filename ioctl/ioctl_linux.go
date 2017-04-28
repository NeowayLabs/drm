package ioctl

import (
	"fmt"
	"syscall"
)

// To decode a hex IOCTL code:
//
// Most architectures use this generic format, but check
// include/ARCH/ioctl.h for specifics, e.g. powerpc
// uses 3 bits to encode read/write and 13 bits for size.
//
//  bits    meaning
//  31-30	00 - no parameters: uses _IO macro
// 	10 - read: _IOR
// 	01 - write: _IOW
// 	11 - read/write: _IOWR
//
//  29-16	size of arguments
//
//  15-8	ascii character supposedly
// 	unique to each driver
//
//  7-0	function #
//
// So for example 0x82187201 is a read with arg length of 0x218,
// character 'r' function 1. Grepping the source reveals this is:
//
// #define VFAT_IOCTL_READDIR_BOTH         _IOR('r', 1, struct dirent [2])
// source: https://www.kernel.org/doc/Documentation/ioctl/ioctl-decoding.txt

type Code struct {
	typ  uint8  // type of ioctl call (read, write, both or none)
	sz   uint16 // size of arguments (only 13bits usable)
	uniq uint8  // unique ascii character for this device
	fn   uint8  // function code
}

const (
	None  = uint8(0x0)
	Write = uint8(0x1)
	Read  = uint8(0x2)
)

func NewCode(typ uint8, sz uint16, uniq, fn uint8) uint32 {
	var code uint32
	if typ > Write|Read {
		panic(fmt.Errorf("invalid ioctl code value: %d\n", typ))
	}

	if sz > 2<<14 {
		panic(fmt.Errorf("invalid ioctl size value: %d\n", sz))
	}

	code = code | (uint32(typ) << 30)
	code = code | (uint32(sz) << 16) // sz has 14bits
	code = code | (uint32(uniq) << 8)
	code = code | uint32(fn)
	return code
}

func Do(fd, cmd, ptr uintptr) error {
	_, _, errcode := syscall.Syscall(syscall.SYS_IOCTL, fd, cmd, ptr)
	if errcode != 0 {
		return errcode
	}
	return nil
}
