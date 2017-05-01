package mode

import (
	"os"
	"unsafe"

	"github.com/tiago4orion/drm"
	"github.com/tiago4orion/drm/ioctl"
)

const (
	DisplayInfoLen   = 32
	ConnectorNameLen = 32
	DisplayModeLen   = 32
	PropNameLen      = 32

	Connected         = 1
	Disconnected      = 2
	UnknownConnection = 3
)

type (
	sysResources struct {
		fbIdPtr              uintptr
		crtcIdPtr            uintptr
		connectorIdPtr       uintptr
		encoderIdPtr         uintptr
		CountFbs             uint32
		CountCrtcs           uint32
		CountConnectors      uint32
		CountEncoders        uint32
		MinWidth, MaxWidth   uint32
		MinHeight, MaxHeight uint32
	}

	sysGetConnector struct {
		encodersPtr   uintptr
		modesPtr      uintptr
		propsPtr      uintptr
		propValuesPtr uintptr

		countModes    uint32
		countProps    uint32
		countEncoders uint32

		encoderID       uint32 // current encoder
		ID              uint32
		connectorType   uint32
		connectorTypeID uint32

		connection        uint32
		mmWidth, mmHeight uint32 // HxW in millimeters
		subpixel          uint32
	}

	Info struct {
		Clock                                         uint32
		Hdisplay, HsyncStart, HsyncEnd, Htotal, Hskew uint16
		Vdisplay, VsyncStart, VsyncEnd, Vtotal, Vscan uint16

		Vrefresh uint32

		Flags uint32
		Type  uint32
		Name  [DisplayModeLen]uint8
	}

	Resources struct {
		sysResources

		Fbs        []uint32
		Crtcs      []uint32
		Connectors []uint32
		Encoders   []uint32
	}

	Connector struct {
		sysGetConnector

		ID            uint32
		EncoderID     uint32
		Type          uint32
		TypeID        uint32
		Connection    uint8
		Width, Height uint32
		Subpixel      uint8

		Modes []Info

		Props      []uint32
		PropValues []uint64

		Encoders []uint32
	}
)

var (
	// DRM_IOWR(0xA0, struct drm_mode_card_res)
	IOCTLModeResources = ioctl.NewCode(ioctl.Read|ioctl.Write,
		uint16(unsafe.Sizeof(sysResources{})), drm.IOCTLBase, 0xA0)

	// DRM_IOWR(0xA7, struct drm_mode_get_connector)
	IOCTLModeGetConnector = ioctl.NewCode(ioctl.Read|ioctl.Write,
		uint16(unsafe.Sizeof(sysGetConnector{})), drm.IOCTLBase, 0xA7)
)

func GetResources(file *os.File) (*Resources, error) {
	mres := &sysResources{}
	err := ioctl.Do(uintptr(file.Fd()), uintptr(IOCTLModeResources),
		uintptr(unsafe.Pointer(mres)))
	if err != nil {
		return nil, err
	}

	var (
		fbids, crtcids, connectorids, encoderids []uint32
	)

	if mres.CountFbs > 0 {
		fbids = make([]uint32, mres.CountFbs)
		mres.fbIdPtr = uintptr(unsafe.Pointer(&fbids[0]))
	}
	if mres.CountCrtcs > 0 {
		crtcids = make([]uint32, mres.CountCrtcs)
		mres.crtcIdPtr = uintptr(unsafe.Pointer(&crtcids[0]))
	}
	if mres.CountEncoders > 0 {
		encoderids = make([]uint32, mres.CountEncoders)
		mres.encoderIdPtr = uintptr(unsafe.Pointer(&encoderids[0]))
	}
	if mres.CountConnectors > 0 {
		connectorids = make([]uint32, mres.CountConnectors)
		mres.connectorIdPtr = uintptr(unsafe.Pointer(&connectorids[0]))
	}

	err = ioctl.Do(uintptr(file.Fd()), uintptr(IOCTLModeResources),
		uintptr(unsafe.Pointer(mres)))
	if err != nil {
		return nil, err
	}

	// TODO(i4k): handle hotplugging in-between the ioctls above

	return &Resources{
		sysResources: *mres,
		Fbs:          fbids,
		Crtcs:        crtcids,
		Encoders:     encoderids,
		Connectors:   connectorids,
	}, nil
}

func GetConnector(file *os.File, connid uint32) (*Connector, error) {
	conn := &sysGetConnector{}
	conn.ID = connid
	err := ioctl.Do(uintptr(file.Fd()), uintptr(IOCTLModeGetConnector),
		uintptr(unsafe.Pointer(conn)))
	if err != nil {
		return nil, err
	}

	var (
		props, encoders []uint32
		propValues      []uint64
		modes           []Info
	)

	if conn.countProps > 0 {
		props = make([]uint32, conn.countProps)
		conn.propsPtr = uintptr(unsafe.Pointer(&props[0]))

		propValues = make([]uint64, conn.countProps)
		conn.propValuesPtr = uintptr(unsafe.Pointer(&propValues[0]))
	}

	if conn.countModes == 0 {
		conn.countModes = 1
	}

	modes = make([]Info, conn.countModes)
	conn.modesPtr = uintptr(unsafe.Pointer(&modes[0]))

	if conn.countEncoders > 0 {
		encoders = make([]uint32, conn.countEncoders)
		conn.encodersPtr = uintptr(unsafe.Pointer(&encoders[0]))
	}

	err = ioctl.Do(uintptr(file.Fd()), uintptr(IOCTLModeGetConnector),
		uintptr(unsafe.Pointer(conn)))
	if err != nil {
		return nil, err
	}

	ret := &Connector{
		sysGetConnector: *conn,
		ID:              conn.ID,
		EncoderID:       conn.encoderID,
		Connection:      uint8(conn.connection),
		Width:           conn.mmWidth,
		Height:          conn.mmHeight,

		// convert subpixel from kernel to userspace */
		Subpixel: uint8(conn.subpixel + 1),
		Type:     conn.connectorType,
		TypeID:   conn.connectorTypeID,
	}

	ret.Props = make([]uint32, len(props))
	copy(ret.Props, props)
	ret.PropValues = make([]uint64, len(propValues))
	copy(ret.PropValues, propValues)
	ret.Modes = make([]Info, len(modes))
	copy(ret.Modes, modes)
	ret.Encoders = make([]uint32, len(encoders))
	copy(ret.Encoders, encoders)

	return ret, nil
}
