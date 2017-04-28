package ioctl

import (
	"strconv"
	"testing"
)

func getbits(n uint32) string {
	return strconv.FormatUint(uint64(n), 2)
}

func TestNewCode(t *testing.T) {
	code := NewCode(Read, 0x218, 'r', 1)
	expected := uint32(0x82187201)
	if code != expected {
		t.Errorf("Expected %s but got %s", getbits(expected),
			getbits(code))
		return
	}
}
