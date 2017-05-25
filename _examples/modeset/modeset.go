// Port of modeset.c example to Go
// Source: https://github.com/dvdhrm/docs/blob/master/drm-howto/modeset.c
package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"
	"unsafe"

	"launchpad.net/gommap"

	"github.com/NeowayLabs/drm"
	"github.com/NeowayLabs/drm/mode"
)

type (
	framebuffer struct {
		id     uint32
		handle uint32
		data   []byte
		fb     *mode.FB
		size   uint64
		stride uint32
	}

	// msetData just store the pair (mode, fb) and the saved CRTC of the mode.
	msetData struct {
		mode      *mode.Modeset
		fb        framebuffer
		savedCrtc *mode.Crtc
	}
)

func createFramebuffer(file *os.File, dev *mode.Modeset) (framebuffer, error) {
	fb, err := mode.CreateFB(file, dev.Width, dev.Height, 32)
	if err != nil {
		return framebuffer{}, fmt.Errorf("Failed to create framebuffer: %s", err.Error())
	}
	stride := fb.Pitch
	size := fb.Size
	handle := fb.Handle

	fbID, err := mode.AddFB(file, dev.Width, dev.Height, 24, 32, stride, handle)
	if err != nil {
		return framebuffer{}, fmt.Errorf("Cannot create dumb buffer: %s", err.Error())
	}

	offset, err := mode.MapDumb(file, handle)
	if err != nil {
		return framebuffer{}, err
	}

	mmap, err := gommap.MapAt(0, uintptr(file.Fd()), int64(offset), int64(size), gommap.PROT_READ|gommap.PROT_WRITE, gommap.MAP_SHARED)
	if err != nil {
		return framebuffer{}, fmt.Errorf("Failed to mmap framebuffer: %s", err.Error())
	}
	for i := uint64(0); i < size; i++ {
		mmap[i] = 0
	}
	framebuf := framebuffer{
		id:     fbID,
		handle: handle,
		data:   mmap,
		fb:     fb,
		size:   size,
		stride: stride,
	}
	return framebuf, nil
}

func draw(msets []msetData) {
	var (
		r, g, b       uint8
		rUp, gUp, bUp = true, true, true
		off           uint32
	)

	rand.Seed(int64(time.Now().Unix()))
	r = uint8(rand.Intn(256))
	g = uint8(rand.Intn(256))
	b = uint8(rand.Intn(256))

	for i := 0; i < 50; i++ {
		r = nextColor(&rUp, r, 20)
		g = nextColor(&gUp, g, 10)
		b = nextColor(&bUp, b, 5)

		for j := 0; j < len(msets); j++ {
			mset := msets[j]
			for k := uint16(0); k < mset.mode.Height; k++ {
				for s := uint16(0); s < mset.mode.Width; s++ {
					off = (mset.fb.stride * uint32(k)) + (uint32(s) * 4)
					val := uint32((uint32(r) << 16) | (uint32(g) << 8) | uint32(b))
					*(*uint32)(unsafe.Pointer(&mset.fb.data[off])) = val
				}
			}
		}

		time.Sleep(150 * time.Millisecond)
	}
}

func nextColor(up *bool, cur uint8, mod int) uint8 {
	var next uint8

	if *up {
		next = cur + 1
	} else {
		next = cur - 1
	}
	next = next * uint8(rand.Intn(mod))
	if (*up && next < cur) || (!*up && next > cur) {
		*up = !*up
		next = cur
	}
	return next
}

func destroyFramebuffer(modeset *mode.SimpleModeset, mset msetData, file *os.File) error {
	handle := mset.fb.handle
	data := mset.fb.data
	fb := mset.fb

	err := gommap.MMap(data).UnsafeUnmap()
	if err != nil {
		return fmt.Errorf("Failed to munmap memory: %s\n", err.Error())
	}
	err = mode.RmFB(file, fb.id)
	if err != nil {
		return fmt.Errorf("Failed to remove frame buffer: %s\n", err.Error())
	}

	err = mode.DestroyDumb(file, handle)
	if err != nil {
		return fmt.Errorf("Failed to destroy dumb buffer: %s\n", err.Error())
	}
	return modeset.SetCrtc(mset.mode, mset.savedCrtc)
}

func cleanup(modeset *mode.SimpleModeset, msets []msetData, file *os.File) {
	for _, mset := range msets {
		destroyFramebuffer(modeset, mset, file)
	}

}

func main() {
	file, err := drm.OpenCard(0)
	if err != nil {
		fmt.Printf("error: %s", err.Error())
		return
	}
	defer file.Close()
	if !drm.HasDumbBuffer(file) {
		fmt.Printf("drm device does not support dumb buffers")
		return
	}
	modeset, err := mode.NewSimpleModeset(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
		os.Exit(1)
	}

	var msets []msetData
	for _, mod := range modeset.Modesets {
		framebuf, err := createFramebuffer(file, &mod)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
			cleanup(modeset, msets, file)
			return
		}

		// save current CRTC of this mode to restore at exit
		savedCrtc, err := mode.GetCrtc(file, mod.Crtc)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: Cannot get CRTC for connector %d: %s", mod.Conn, err.Error())
			cleanup(modeset, msets, file)
			return
		}
		// change the mode
		err = mode.SetCrtc(file, mod.Crtc, framebuf.id, 0, 0, &mod.Conn, 1, &mod.Mode)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot set CRTC for connector %d: %s", mod.Conn, err.Error())
			cleanup(modeset, msets, file)
			return
		}
		msets = append(msets, msetData{
			mode:      &mod,
			fb:        framebuf,
			savedCrtc: savedCrtc,
		})
	}

	draw(msets)
	cleanup(modeset, msets, file)
}
