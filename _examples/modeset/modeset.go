// Port of modeset.c example to Go
// Source: https://github.com/dvdhrm/docs/blob/master/drm-howto/modeset.c
package main

import (
	"fmt"
	"log"
	"os"

	"launchpad.net/gommap"

	"github.com/tiago4orion/drm"
	"github.com/tiago4orion/drm/mode"
)

type modeset struct {
	width, height uint16
	stride        uint32
	size          uint64
	handle        uint32
	data          []byte

	mode      mode.Info
	fb        uint32
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
	dev.width = conn.Modes[0].Hdisplay
	dev.height = conn.Modes[0].Vdisplay

	log.Println("mode for connector %u is %ux%u\n", conn.ID, dev.width, dev.height)

	err := findCrtc(file, res, conn, dev)
	if err != nil {
		return false, fmt.Errorf("no valid crtc for connector %u: %s", conn.ID, err.Error())
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

func createFramebuffer(file *os.File, dev *modeset) error {
	fb, err := mode.CreateFB(file, dev.width, dev.height, 32)
	if err != nil {
		return fmt.Errorf("Failed to create framebuffer: %s", err.Error())
	}
	dev.stride = fb.Pitch
	dev.size = fb.Size
	dev.handle = fb.Handle
	fbID, err := mode.AddFB(file, dev.width, dev.height, 24, 32, dev.stride, dev.handle)
	if err != nil {
		return fmt.Errorf("Cannot create dumb buffer: %s", err.Error())
	}
	dev.fb = fbID
	mmap, err := gommap.Map(file.Fd(), gommap.PROT_READ|gommap.PROT_WRITE, gommap.MAP_SHARED)
	if err != nil {
		return fmt.Errorf("Failed to mmap framebuffer: %s", err.Error())
	}
	for i := uint64(0); i < dev.size; i++ {
		mmap[i] = 0
	}
	return nil
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
			log.Printf("[error] 1- Cannot set CRTC for connector %d", mset.conn)
			return
		}
		err = mode.SetCrtc(file, mset.crtc, mset.fb, 0, 0, &mset.conn, 1, &mset.mode)
		if err != nil {
			log.Printf("[error] Cannot set CRTC for connector %d", mset.conn)
			return
		}
	}
}
