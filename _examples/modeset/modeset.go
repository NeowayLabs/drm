// Port of modeset.c example to Go
// Source: https://github.com/dvdhrm/docs/blob/master/drm-howto/modeset.c
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/tiago4orion/drm"
	"github.com/tiago4orion/drm/mode"
)

type modeset struct {
	width, height uint16
	stride        uint32
	size          uint32
	handle        uint32
	data          []byte

	mode mode.Info
	fb   uint32
	conn uint32
	crtc uint32
	//	savedCtrc *drm.ModeCrtc
}

var modesetlist []modeset

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
	return true, nil
}

func findCrtc(file *os.File, res *mode.Resources, conn *mode.Connector, dev *modeset) {
	// TODO
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
}
