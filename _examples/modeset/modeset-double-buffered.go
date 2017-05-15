// Port of modeset.c example to Go
// Source: https://github.com/dvdhrm/docs/blob/master/drm-howto/modeset-double-buffered.c
package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"launchpad.net/gommap"

	"github.com/NeowayLabs/drm"
	"github.com/NeowayLabs/drm/mode"
)

type modesetBuf struct {
	width, height uint16
	stride        uint32
	size          uint64
	handle        uint32
	data          []byte
	fb            uint32
}

type modeset struct {
	frontBuf  uint
	bufs      [2]modesetBuf

	mode      mode.Info
	conn      uint32
	crtc      uint32
	savedCtrc *mode.Crtc
}

var modesetlist []*modeset

func prepare(file *os.File) error {
	res, err := mode.GetResources(file)
	if err != nil {
		return fmt.Errorf("Cannot retrieve resources: %s", err.Error())
	}

	for i := 0; i < len(res.Connectors); i++ {
		conn, err := mode.GetConnector(file, res.Connectors[i])
		if err != nil {
			return fmt.Errorf("Cannot retrieve connector: %s", err.Error())
		}

		dev := &modeset{}
		dev.conn = conn.ID
		ok, err := setupDev(file, res, conn, dev)
		if err != nil {
			return err
		}

		if !ok {
			continue
		}

		modesetlist = append(modesetlist, dev)
		fmt.Printf("%#v\n", conn)
	}

	return nil
}

func setupDev(file *os.File, res *mode.Resources, conn *mode.Connector, dev *modeset) (bool, error) {
	// check if a monitor is connected
	if conn.Connection != mode.Connected {
		log.Printf("Ignoring unused connector %d: %d", conn.ID, conn.Connection)
		return false, nil
	}

	// check if there is at least one valid mode
	if len(conn.Modes) == 0 {
		return false, fmt.Errorf("no valid mode for connector %d\n", conn.ID)
	}
	dev.mode = conn.Modes[0]
	dev.bufs[0].width = conn.Modes[0].Hdisplay
	dev.bufs[0].height = conn.Modes[0].Vdisplay
	dev.bufs[1].width = conn.Modes[0].Hdisplay
	dev.bufs[1].height = conn.Modes[0].Vdisplay

	log.Printf("mode for connector %d is %dx%d\n", conn.ID, conn.Modes[0].Hdisplay, conn.Modes[0].Vdisplay)

	err := findCrtc(file, res, conn, dev)
	if err != nil {
		return false, fmt.Errorf("no valid crtc for connector %u: %s", conn.ID, err.Error())
	}

	err = createFramebuffer(file, &dev.bufs[0])
	if err != nil {
		return false, err
	}

	err = createFramebuffer(file, &dev.bufs[1])
	if err != nil {
		return false, err
	}

	return true, nil
}

func findCrtc(file *os.File, res *mode.Resources, conn *mode.Connector, dev *modeset) error {
	var (
		encoder *mode.Encoder
		err     error
	)

	if conn.EncoderID != 0 {
		encoder, err = mode.GetEncoder(file, conn.EncoderID)
		if err != nil {
			return err
		}
	}

	if encoder != nil {
		if encoder.CrtcID != 0 {
			crtcid := encoder.CrtcID
			found := false

			for i := 0; i < len(modesetlist); i++ {
				if modesetlist[i].crtc == crtcid {
					found = true
					break
				}
			}

			if crtcid >= 0 && !found {
				dev.crtc = crtcid
				return nil
			}
		}
	}

	// If the connector is not currently bound to an encoder or if the
	// encoder+crtc is already used by another connector (actually unlikely
	// but lets be safe), iterate all other available encoders to find a
	// matching CRTC.
	for i := 0; i < len(conn.Encoders); i++ {
		encoder, err := mode.GetEncoder(file, conn.Encoders[i])
		if err != nil {
			return fmt.Errorf("Cannot retrieve encoder: %s", err.Error())
		}
		// iterate all global CRTCs
		for j := 0; j < len(res.Crtcs); j++ {
			// check whether this CRTC works with the encoder
			if (encoder.PossibleCrtcs & (1 << uint(j))) != 0 {
				continue
			}

			// check that no other device already uses this CRTC
			crtcid := res.Crtcs[j]
			found := false
			for k := 0; k < len(modesetlist); k++ {
				if modesetlist[k].crtc == crtcid {
					found = true
					break
				}
			}

			// we have found a CRTC, so save it and return
			if crtcid >= 0 && !found {
				dev.crtc = crtcid
				return nil
			}
		}
	}

	return fmt.Errorf("Cannot find a suitable CRTC for connector %d", conn.ID)
}

func createFramebuffer(file *os.File, buf *modesetBuf) error {
	fb, err := mode.CreateFB(file, buf.width, buf.height, 32)
	if err != nil {
		return fmt.Errorf("Failed to create framebuffer: %s", err.Error())
	}
	buf.stride = fb.Pitch
	buf.size = fb.Size
	buf.handle = fb.Handle
	fbID, err := mode.AddFB(file, buf.width, buf.height, 24, 32, buf.stride, buf.handle)
	if err != nil {
		return fmt.Errorf("Cannot create dumb buffer: %s", err.Error())
	}
	buf.fb = fbID

	offset, err := mode.MapDumb(file, buf.handle)
	if err != nil {
		return err
	}

	mmap, err := gommap.MapAt(0, uintptr(file.Fd()), int64(offset), int64(buf.size), gommap.PROT_READ|gommap.PROT_WRITE, gommap.MAP_SHARED)
	if err != nil {
		return fmt.Errorf("Failed to mmap framebuffer: %s", err.Error())
	}
	for i := uint64(0); i < buf.size; i++ {
		mmap[i] = 0
	}
	buf.data = mmap
	return nil
}

func destroyFramebuffer(file *os.File, buf *modesetBuf) error {
	for i := 0; i < len(buf.data); i++ {
		buf.data[i] = 0
	}

	err := gommap.MMap(buf.data).UnsafeUnmap()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to munmap memory: %s\n", err.Error())
		return err
	}
	err = mode.RmFB(file, buf.fb)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to remove frame buffer: %s\n", err.Error())
		return err
	}

	err = mode.DestroyDumb(file, buf.handle)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to destroy dumb buffer: %s\n", err.Error())
		return err
	}

	return nil
}

func draw(file *os.File) {
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

		for j := 0; j < len(modesetlist); j++ {
			iter := modesetlist[j]
			buf := &iter.bufs[iter.frontBuf ^ 1]
			for k := uint16(0); k < buf.height; k++ {
				for s := uint16(0); s < buf.width; s++ {
					off = (buf.stride * uint32(k)) + (uint32(s) * 4)
					buf.data[off] = (r << 16) | (g << 8) | b
				}
			}

			err := mode.SetCrtc(file, iter.crtc, buf.fb, 0, 0, &iter.conn, 1, &iter.mode)
			if err != nil {
				log.Printf("[error] Cannot flip CRTC for connector %d: %s", iter.conn, err.Error())
				return
			} else {
				iter.frontBuf ^= 1
			}

		}

		time.Sleep(100 * time.Millisecond)
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

func cleanup(file *os.File) {
	for _, dev := range modesetlist {
		err := mode.SetCrtc(file, dev.savedCtrc.ID,
			dev.savedCtrc.BufferID,
			dev.savedCtrc.X, dev.savedCtrc.Y,
			&dev.conn,
			1,
			&dev.savedCtrc.Mode,
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to restore CRTC: %s\n", err.Error())
			continue
		}

		// destroy framebuffers
		destroyFramebuffer(file, &dev.bufs[1])
		destroyFramebuffer(file, &dev.bufs[0])
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
	err = prepare(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
		return
	}

	for i := 0; i < len(modesetlist); i++ {
		var err error

		mset := modesetlist[i]
		mset.savedCtrc, err = mode.GetCrtc(file, mset.crtc)
		if err != nil {
			log.Printf("[error] Cannot get CRTC for connector %d: %s", mset.conn, err.Error())
			return
		}
		fmt.Printf("crtc = %d, conn = %d, mode = %#v\n", mset.crtc, mset.conn, mset.mode)
		buf := &mset.bufs[mset.frontBuf]
		fmt.Printf("fb = %d\n", buf.fb)
		err = mode.SetCrtc(file, mset.crtc, buf.fb, 0, 0, &mset.conn, 1, &mset.mode)
		if err != nil {
			log.Printf("[error] Cannot set CRTC for connector %d: %s", mset.conn, err.Error())
			return
		}
	}

	draw(file)
	cleanup(file)
}
