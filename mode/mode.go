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

	sysGetEncoder struct {
		id  uint32
		typ uint32

		crtcID uint32

		possibleCrtcs  uint32
		possibleClones uint32
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

	Encoder struct {
		ID   uint32
		Type uint32

		CrtcID uint32

		PossibleCrtcs  uint32
		PossibleClones uint32
	}

	sysCreateDumb struct {
		height, width uint32
		bpp           uint32
		flags         uint32

		// returned values
		handle uint32
		pitch  uint32
		size   uint64
	}

	sysMapDumb struct {
		handle uint32 // Handle for the object being mapped
		pad    uint32

		// Fake offset to use for subsequent mmap call
		// This is a fixed-size type for 32/64 compatibility.
		offset uint64
	}

	sysFBCmd struct {
		fbID          uint32
		width, height uint32
		pitch         uint32
		bpp           uint32
		depth         uint32

		/* driver specific handle */
		handle uint32
	}

	sysRmFB struct {
		handle uint32
	}

	sysCrtc struct {
		setConnectorsPtr uintptr
		countConnectors  uint32

		id   uint32
		fbID uint32 // Id of framebuffer

		x, y uint32 // Position on the frameuffer

		gammaSize uint32
		modeValid uint32
		mode      Info
	}

	sysDestroyDumb struct {
		handle uint32
	}

	Crtc struct {
		ID       uint32
		BufferID uint32 // FB id to connect to 0 = disconnect

		X, Y          uint32 // Position on the framebuffer
		Width, Height uint32
		ModeValid     int
		Mode          Info

		GammaSize int // Number of gamma stops
	}

	FB struct {
		Height, Width, BPP, Flags uint32
		Handle                    uint32
		Pitch                     uint32
		Size                      uint64
	}
)

var (
	// DRM_IOWR(0xA0, struct drm_mode_card_res)
	IOCTLModeResources = ioctl.NewCode(ioctl.Read|ioctl.Write,
		uint16(unsafe.Sizeof(sysResources{})), drm.IOCTLBase, 0xA0)

	// DRM_IOWR(0xA1, struct drm_mode_crtc)
	IOCTLModeGetCrtc = ioctl.NewCode(ioctl.Read|ioctl.Write,
		uint16(unsafe.Sizeof(sysCrtc{})), drm.IOCTLBase, 0xA1)

	// DRM_IOWR(0xA2, struct drm_mode_crtc)
	IOCTLModeSetCrtc = ioctl.NewCode(ioctl.Read|ioctl.Write,
		uint16(unsafe.Sizeof(sysCrtc{})), drm.IOCTLBase, 0xA2)

	// DRM_IOWR(0xA6, struct drm_mode_get_encoder)
	IOCTLModeGetEncoder = ioctl.NewCode(ioctl.Read|ioctl.Write,
		uint16(unsafe.Sizeof(sysGetEncoder{})), drm.IOCTLBase, 0xA6)

	// DRM_IOWR(0xA7, struct drm_mode_get_connector)
	IOCTLModeGetConnector = ioctl.NewCode(ioctl.Read|ioctl.Write,
		uint16(unsafe.Sizeof(sysGetConnector{})), drm.IOCTLBase, 0xA7)

	// DRM_IOWR(0xAE, struct drm_mode_fb_cmd)
	IOCTLModeAddFB = ioctl.NewCode(ioctl.Read|ioctl.Write,
		uint16(unsafe.Sizeof(sysFBCmd{})), drm.IOCTLBase, 0xAE)

	// DRM_IOWR(0xAF, unsigned int)
	IOCTLModeRmFB = ioctl.NewCode(ioctl.Read|ioctl.Write,
		uint16(unsafe.Sizeof(uint32(0))), drm.IOCTLBase, 0xAF)

	// DRM_IOWR(0xB2, struct drm_mode_create_dumb)
	IOCTLModeCreateDumb = ioctl.NewCode(ioctl.Read|ioctl.Write,
		uint16(unsafe.Sizeof(sysCreateDumb{})), drm.IOCTLBase, 0xB2)

	// DRM_IOWR(0xB3, struct drm_mode_map_dumb)
	IOCTLModeMapDumb = ioctl.NewCode(ioctl.Read|ioctl.Write,
		uint16(unsafe.Sizeof(sysMapDumb{})), drm.IOCTLBase, 0xB3)

	// DRM_IOWR(0xB4, struct drm_mode_destroy_dumb)
	IOCTLModeDestroyDumb = ioctl.NewCode(ioctl.Read|ioctl.Write,
		uint16(unsafe.Sizeof(sysDestroyDumb{})), drm.IOCTLBase, 0xB4)
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

func GetEncoder(file *os.File, id uint32) (*Encoder, error) {
	encoder := &sysGetEncoder{}
	encoder.id = id

	err := ioctl.Do(uintptr(file.Fd()), uintptr(IOCTLModeGetEncoder),
		uintptr(unsafe.Pointer(encoder)))
	if err != nil {
		return nil, err
	}

	return &Encoder{
		ID:             encoder.id,
		CrtcID:         encoder.crtcID,
		Type:           encoder.typ,
		PossibleCrtcs:  encoder.possibleCrtcs,
		PossibleClones: encoder.possibleClones,
	}, nil
}

func CreateFB(file *os.File, width, height uint16, bpp uint32) (*FB, error) {
	fb := &sysCreateDumb{}
	fb.width = uint32(width)
	fb.height = uint32(height)
	fb.bpp = bpp
	err := ioctl.Do(uintptr(file.Fd()), uintptr(IOCTLModeCreateDumb),
		uintptr(unsafe.Pointer(fb)))
	if err != nil {
		return nil, err
	}
	return &FB{
		Height: fb.height,
		Width:  fb.width,
		BPP:    fb.bpp,
		Handle: fb.handle,
		Pitch:  fb.pitch,
		Size:   fb.size,
	}, nil
}

func AddFB(file *os.File, width, height uint16,
	depth, bpp uint8, pitch, boHandle uint32) (uint32, error) {
	f := &sysFBCmd{}
	f.width = uint32(width)
	f.height = uint32(height)
	f.pitch = pitch
	f.bpp = uint32(bpp)
	f.depth = uint32(depth)
	f.handle = boHandle
	err := ioctl.Do(uintptr(file.Fd()), uintptr(IOCTLModeAddFB),
		uintptr(unsafe.Pointer(f)))
	if err != nil {
		return 0, err
	}
	return f.fbID, nil
}

func RmFB(file *os.File, bufferid uint32) error {
	return ioctl.Do(uintptr(file.Fd()), uintptr(IOCTLModeRmFB),
		uintptr(unsafe.Pointer(&sysRmFB{bufferid})))
}

func MapDumb(file *os.File, boHandle uint32) (uint64, error) {
	mreq := &sysMapDumb{}
	mreq.handle = boHandle
	err := ioctl.Do(uintptr(file.Fd()), uintptr(IOCTLModeMapDumb),
		uintptr(unsafe.Pointer(mreq)))
	if err != nil {
		return 0, err
	}
	return mreq.offset, nil
}

func DestroyDumb(file *os.File, handle uint32) error {
	return ioctl.Do(uintptr(file.Fd()), uintptr(IOCTLModeDestroyDumb),
		uintptr(unsafe.Pointer(&sysDestroyDumb{handle})))
}

func GetCrtc(file *os.File, id uint32) (*Crtc, error) {
	crtc := &sysCrtc{}
	crtc.id = id
	err := ioctl.Do(uintptr(file.Fd()), uintptr(IOCTLModeGetCrtc),
		uintptr(unsafe.Pointer(crtc)))
	if err != nil {
		return nil, err
	}
	ret := &Crtc{
		ID:        crtc.id,
		X:         crtc.x,
		Y:         crtc.y,
		ModeValid: int(crtc.modeValid),
		BufferID:  crtc.fbID,
		GammaSize: int(crtc.gammaSize),
	}

	ret.Mode = crtc.mode
	ret.Width = uint32(crtc.mode.Hdisplay)
	ret.Height = uint32(crtc.mode.Vdisplay)
	return ret, nil
}

func SetCrtc(file *os.File, crtcid, bufferid, x, y uint32, connectors *uint32, count int, mode *Info) error {
	crtc := &sysCrtc{}
	crtc.x = x
	crtc.y = y
	crtc.id = crtcid
	crtc.fbID = bufferid
	if connectors != nil {
		crtc.setConnectorsPtr = uintptr(unsafe.Pointer(connectors))
	}
	crtc.countConnectors = uint32(count)
	if mode != nil {
		crtc.mode = *mode
		crtc.modeValid = 1
	}
	return ioctl.Do(uintptr(file.Fd()), uintptr(IOCTLModeSetCrtc),
		uintptr(unsafe.Pointer(crtc)))
}
