// Port of modeset.c example to Go
// Source: https://github.com/dvdhrm/docs/blob/master/drm-howto/modeset.c
package mode

import (
	"fmt"
	_ "image/jpeg"
	"os"
)

type (
	Modeset struct {
		Width, Height uint16

		Mode Info
		Conn uint32
		Crtc uint32
	}

	SimpleModeset struct {
		Modesets []Modeset
		driFile  *os.File
	}
)

func (mset *SimpleModeset) prepare() error {
	res, err := GetResources(mset.driFile)
	if err != nil {
		return fmt.Errorf("Cannot retrieve resources: %s", err.Error())
	}

	for i := 0; i < len(res.Connectors); i++ {
		conn, err := GetConnector(mset.driFile, res.Connectors[i])
		if err != nil {
			return fmt.Errorf("Cannot retrieve connector: %s", err.Error())
		}

		dev := Modeset{}
		dev.Conn = conn.ID
		ok, err := mset.setupDev(res, conn, &dev)
		if err != nil {
			return err
		}

		if !ok {
			continue
		}

		mset.Modesets = append(mset.Modesets, dev)
	}

	return nil
}

func (mset *SimpleModeset) setupDev(res *Resources, conn *Connector, dev *Modeset) (bool, error) {
	// check if a monitor is connected
	if conn.Connection != Connected {
		return false, nil
	}

	// check if there is at least one valid mode
	if len(conn.Modes) == 0 {
		return false, fmt.Errorf("no valid mode for connector %d\n", conn.ID)
	}
	dev.Mode = conn.Modes[0]
	dev.Width = conn.Modes[0].Hdisplay
	dev.Height = conn.Modes[0].Vdisplay

	err := mset.findCrtc(res, conn, dev)
	if err != nil {
		return false, fmt.Errorf("no valid crtc for connector %u: %s", conn.ID, err.Error())
	}

	return true, nil
}

func (mset *SimpleModeset) findCrtc(res *Resources, conn *Connector, dev *Modeset) error {
	var (
		encoder *Encoder
		err     error
	)

	if conn.EncoderID != 0 {
		encoder, err = GetEncoder(mset.driFile, conn.EncoderID)
		if err != nil {
			return err
		}
	}

	if encoder != nil {
		if encoder.CrtcID != 0 {
			crtcid := encoder.CrtcID
			found := false

			for i := 0; i < len(mset.Modesets); i++ {
				if mset.Modesets[i].Crtc == crtcid {
					found = true
					break
				}
			}

			if crtcid >= 0 && !found {
				dev.Crtc = crtcid
				return nil
			}
		}
	}

	// If the connector is not currently bound to an encoder or if the
	// encoder+crtc is already used by another connector (actually unlikely
	// but lets be safe), iterate all other available encoders to find a
	// matching CRTC.
	for i := 0; i < len(conn.Encoders); i++ {
		encoder, err := GetEncoder(mset.driFile, conn.Encoders[i])
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
			for k := 0; k < len(mset.Modesets); k++ {
				if mset.Modesets[k].Crtc == crtcid {
					found = true
					break
				}
			}

			// we have found a CRTC, so save it and return
			if crtcid >= 0 && !found {
				dev.Crtc = crtcid
				return nil
			}
		}
	}

	return fmt.Errorf("Cannot find a suitable CRTC for connector %d", conn.ID)
}

func (mset *SimpleModeset) SetCrtc(dev *Modeset, savedCrtc *Crtc) error {
	err := SetCrtc(mset.driFile, savedCrtc.ID,
		savedCrtc.BufferID,
		savedCrtc.X, savedCrtc.Y,
		&dev.Conn,
		1,
		&savedCrtc.Mode,
	)
	if err != nil {
		return fmt.Errorf("Failed to restore CRTC: %s\n", err.Error())
	}

	return nil
}

func NewSimpleModeset(file *os.File) (*SimpleModeset, error) {
	var err error

	mset := &SimpleModeset{
		driFile: file,
	}
	err = mset.prepare()
	if err != nil {
		return nil, err
	}

	return mset, nil
}
