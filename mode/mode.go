package mode

import (
	"bytes"
	"os"
	"slices"
	"unsafe"

	"github.com/NeowayLabs/drm"
	"github.com/NeowayLabs/drm/ioctl"
)

const (
	DisplayInfoLen   = 32
	ConnectorNameLen = 32
	DisplayModeLen   = 32
	PropNameLen      = 32

	Connected         = 1
	Disconnected      = 2
	UnknownConnection = 3

	// deprecated
	PropPending = 1 << 0
	// legacy types
	PropRagen     = 1 << 1
	PropImmutable = 1 << 2
	PropEnum      = 1 << 3
	PropBlob      = 1 << 4
	PropBitmask   = 1 << 5
	// extended types
	PropExtended    = 0x0000ffc0
	PropObject      = 1 << 6
	PropSignedrange = 2 << 6
	// atomic flag
	PropAtomic = 0x80000000

	// Client Capabilities
	ClientCapStereo3D            = 1
	ClientCapUniversalPlanes     = 2
	ClientCapAtomic              = 3
	ClientCapAspectRatio         = 4
	ClientCapWritebackConnectors = 5

	// Object Types
	ObjectCRTC      = 0xcccccccc
	ObjectConnector = 0xc0c0c0c0
	ObjectEncoder   = 0xe0e0e0e0
	ObjectMode      = 0xdededede
	ObjectProperty  = 0xb0b0b0b0
	ObjectFB        = 0xfbfbfbfb
	ObjectBLOB      = 0xbbbbbbbb
	ObjectPlane     = 0xeeeeeeee
	ObjectAny       = 0

	// Atomic Flags
	PageFlipEvent      = 0x01
	PageFlipAsync      = 0x02
	AtomicTestOnly     = 0x0100
	AtomicNonBlock     = 0x0200
	AtomicAllowModeSet = 0x0400
)

type (
	sysResources struct {
		fbIdPtr              uint64
		crtcIdPtr            uint64
		connectorIdPtr       uint64
		encoderIdPtr         uint64
		CountFbs             uint32
		CountCrtcs           uint32
		CountConnectors      uint32
		CountEncoders        uint32
		MinWidth, MaxWidth   uint32
		MinHeight, MaxHeight uint32
	}

	sysGetConnector struct {
		encodersPtr   uint64
		modesPtr      uint64
		propsPtr      uint64
		propValuesPtr uint64

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

	sysGetPlaneResources struct {
		planeIdPtr  uint64
		countPlanes uint32
	}

	sysGetPlane struct {
		planeId          uint32
		crtcId           uint32
		fbId             uint32
		possibleCrtcs    uint32
		gammaSize        uint32
		countFormatTypes uint32
		formatTypePtr    uint64
	}

	sysSetPlane struct {
		planeId uint32
		crtcId  uint32
		fbId    uint32
		flags   uint32
		crtcX   int32
		crtcY   int32
		crtcW   uint32
		crtcH   uint32
		srcX    uint32
		srcY    uint32
		srcH    uint32
		srcW    uint32
	}

	sysGetProperty struct {
		valuesPtr      uint64
		enumBlobPtr    uint64
		propId         uint32
		flags          uint32
		name           [PropNameLen]uint8
		countValues    uint32
		countEnumBlobs uint32
	}

	sysPropertyEnum struct {
		value uint64
		name  [PropNameLen]uint8
	}

	sysGetBlob struct {
		data   uint64
		length uint32
		blobId uint32
	}

	sysCreateBlob struct {
		data   uint64
		length uint32
		blobId uint32
	}

	sysDestroyBlob struct {
		blobId uint32
	}

	sysSetClientCap struct {
		capability uint64
		value      uint64
	}

	sysObjGetProperties struct {
		propsPtr      uint64
		propValuesPtr uint64
		countProps    uint32
		objID         uint32
		objType       uint32
	}

	sysAtomic struct {
		flags         uint32
		countObjs     uint32
		objsPtr       uint64
		countPropsPtr uint64
		propsPtr      uint64
		propValuesPtr uint64
		reserved      uint64
		userData      uint64
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

	PlaneResources struct {
		sysGetPlaneResources

		Planes []uint32
	}

	Plane struct {
		sysGetPlane

		ID            uint32
		CrtcID        uint32
		FbID          uint32
		PossibleCrtcs uint32
		GammaSize     uint32
		FormatTypes   []uint32
	}

	Property struct {
		ID        uint32
		Values    []uint64
		EnumBlobs []PropertyEnum
		Flags     uint32
		Name      string
	}

	PropertyEnum struct {
		Value uint64
		Name  string
	}

	Blob struct {
		ID   uint32
		Data []byte
	}

	Properties struct {
		ObjectID   uint32
		ObjectType uint32

		Props      []uint32
		PropValues []uint64
	}

	AtomicProperty struct {
		ObjectID   uint32
		PropertyID uint32
		Value      uint64
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

	sysFBCmd2 struct {
		fbID          uint32
		width, height uint32
		pixelFormat   uint32
		flags         uint32
		handles       [4]uint32
		pitches       [4]uint32
		offsets       [4]uint32
		modifier      [4]uint64
	}

	sysRmFB struct {
		handle uint32
	}

	sysCrtc struct {
		setConnectorsPtr uint64
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

	// DRM_IOWR(0xAE, struct drm_mode_fb_cmd)
	IOCTLModeAddFB2 = ioctl.NewCode(ioctl.Read|ioctl.Write,
		uint16(unsafe.Sizeof(sysFBCmd2{})), drm.IOCTLBase, 0xB8)

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

	// DRM_IOWR(0xB5, struct drm_mode_get_plane_res)
	IOCTLModeGetPlaneResources = ioctl.NewCode(ioctl.Read|ioctl.Write,
		uint16(unsafe.Sizeof(sysGetPlaneResources{})), drm.IOCTLBase, 0xB5)

	// DRM_IOWR(0xB6, struct drm_mode_get_plane)
	IOCTLModeGetPlane = ioctl.NewCode(ioctl.Read|ioctl.Write,
		uint16(unsafe.Sizeof(sysGetPlane{})), drm.IOCTLBase, 0xB6)

	// DRM_IOWR(0xB7, struct drm_mode_set_plane)
	IOCTLModeSetPlane = ioctl.NewCode(ioctl.Read|ioctl.Write,
		uint16(unsafe.Sizeof(sysSetPlane{})), drm.IOCTLBase, 0xB7)

	// DRM_IOWR(0xAA, struct drm_mode_get_property)
	IOCTLModeGetProperty = ioctl.NewCode(ioctl.Read|ioctl.Write,
		uint16(unsafe.Sizeof(sysGetProperty{})), drm.IOCTLBase, 0xAA)

	// DRM_IOWR(0xAC, struct drm_mode_get_blob)
	IOCTLModeGetPropBlob = ioctl.NewCode(ioctl.Read|ioctl.Write,
		uint16(unsafe.Sizeof(sysGetBlob{})), drm.IOCTLBase, 0xAC)

	// DRM_IOW(0x0D, struct drm_set_client_cap)
	IOCTLSetClientCap = ioctl.NewCode(ioctl.Write,
		uint16(unsafe.Sizeof(sysSetClientCap{})), drm.IOCTLBase, 0x0D)

	// DRM_IOWR(0xB9, struct drm_mode_obj_get_properties)
	IOCTLModeObjGetProperties = ioctl.NewCode(ioctl.Read|ioctl.Write,
		uint16(unsafe.Sizeof(sysObjGetProperties{})), drm.IOCTLBase, 0xB9)

	// DRM_IOWR(0xBC, struct drm_mode_atomic)
	IOCTLModeAtomic = ioctl.NewCode(ioctl.Read|ioctl.Write,
		uint16(unsafe.Sizeof(sysAtomic{})), drm.IOCTLBase, 0xBC)

	// DRM_IOWR(0xBD, struct drm_mode_create_blob)
	IOCTLModeCreateBlob = ioctl.NewCode(ioctl.Read|ioctl.Write,
		uint16(unsafe.Sizeof(sysCreateBlob{})), drm.IOCTLBase, 0xBD)

	// DRM_IOWR(0xBE, struct drm_mode_destroy_blob)
	IOCTLModeDestroyBlob = ioctl.NewCode(ioctl.Read|ioctl.Write,
		uint16(unsafe.Sizeof(sysDestroyBlob{})), drm.IOCTLBase, 0xBE)
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
		mres.fbIdPtr = uint64(uintptr(unsafe.Pointer(&fbids[0])))
	}
	if mres.CountCrtcs > 0 {
		crtcids = make([]uint32, mres.CountCrtcs)
		mres.crtcIdPtr = uint64(uintptr(unsafe.Pointer(&crtcids[0])))
	}
	if mres.CountEncoders > 0 {
		encoderids = make([]uint32, mres.CountEncoders)
		mres.encoderIdPtr = uint64(uintptr(unsafe.Pointer(&encoderids[0])))
	}
	if mres.CountConnectors > 0 {
		connectorids = make([]uint32, mres.CountConnectors)
		mres.connectorIdPtr = uint64(uintptr(unsafe.Pointer(&connectorids[0])))
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
		conn.propsPtr = uint64(uintptr(unsafe.Pointer(&props[0])))

		propValues = make([]uint64, conn.countProps)
		conn.propValuesPtr = uint64(uintptr(unsafe.Pointer(&propValues[0])))
	}

	if conn.countModes == 0 {
		conn.countModes = 1
	}

	modes = make([]Info, conn.countModes)
	conn.modesPtr = uint64(uintptr(unsafe.Pointer(&modes[0])))

	if conn.countEncoders > 0 {
		encoders = make([]uint32, conn.countEncoders)
		conn.encodersPtr = uint64(uintptr(unsafe.Pointer(&encoders[0])))
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

func GetPlaneResources(file *os.File) (*PlaneResources, error) {
	mPlaneRes := &sysGetPlaneResources{}
	err := ioctl.Do(uintptr(file.Fd()), uintptr(IOCTLModeGetPlaneResources),
		uintptr(unsafe.Pointer(mPlaneRes)))
	if err != nil {
		return nil, err
	}

	var (
		planeIds []uint32
	)

	if mPlaneRes.countPlanes > 0 {
		planeIds = make([]uint32, mPlaneRes.countPlanes)
		mPlaneRes.planeIdPtr = uint64(uintptr(unsafe.Pointer(&planeIds[0])))
	}

	err = ioctl.Do(uintptr(file.Fd()), uintptr(IOCTLModeGetPlaneResources),
		uintptr(unsafe.Pointer(mPlaneRes)))
	if err != nil {
		return nil, err
	}

	return &PlaneResources{
		sysGetPlaneResources: *mPlaneRes,
		Planes:               planeIds,
	}, nil
}

func GetPlane(file *os.File, id uint32) (*Plane, error) {
	mPlaneRes := &sysGetPlane{planeId: id}
	err := ioctl.Do(uintptr(file.Fd()), uintptr(IOCTLModeGetPlane),
		uintptr(unsafe.Pointer(mPlaneRes)))
	if err != nil {
		return nil, err
	}

	var (
		formatTypes []uint32
	)

	if mPlaneRes.countFormatTypes > 0 {
		formatTypes = make([]uint32, mPlaneRes.countFormatTypes)
		mPlaneRes.formatTypePtr = uint64(uintptr(unsafe.Pointer(&formatTypes[0])))
	}

	err = ioctl.Do(uintptr(file.Fd()), uintptr(IOCTLModeGetPlane),
		uintptr(unsafe.Pointer(mPlaneRes)))
	if err != nil {
		return nil, err
	}

	return &Plane{
		sysGetPlane:   *mPlaneRes,
		ID:            mPlaneRes.planeId,
		CrtcID:        mPlaneRes.crtcId,
		FbID:          mPlaneRes.fbId,
		PossibleCrtcs: mPlaneRes.possibleCrtcs,
		GammaSize:     mPlaneRes.gammaSize,
		FormatTypes:   formatTypes,
	}, nil
}

func SetPlane(file *os.File, planeId, crtcId, fbId uint32, flags uint32, crtcX, crtcY int32, crtcW, crtcH, srcX, srcY, srcH, srcW uint32) error {
	mPlaneRes := &sysSetPlane{
		planeId: planeId,
		crtcId:  crtcId,
		fbId:    fbId,
		flags:   flags,
		crtcX:   crtcX,
		crtcY:   crtcY,
		crtcW:   crtcW,
		crtcH:   crtcH,
		srcX:    srcX,
		srcY:    srcY,
		srcW:    srcW,
		srcH:    srcH,
	}
	err := ioctl.Do(uintptr(file.Fd()), uintptr(IOCTLModeSetPlane),
		uintptr(unsafe.Pointer(mPlaneRes)))
	return err
}

func GetProperty(file *os.File, id uint32) (*Property, error) {
	propertyRes := &sysGetProperty{propId: id, countValues: 0, countEnumBlobs: 0}
	err := ioctl.Do(uintptr(file.Fd()), uintptr(IOCTLModeGetProperty),
		uintptr(unsafe.Pointer(propertyRes)))
	if err != nil {
		return nil, err
	}

	// Create arrays to store the values and enums.
	var (
		values    []uint64
		enumBlobs []sysPropertyEnum
	)

	if propertyRes.countValues > 0 {
		values = make([]uint64, propertyRes.countValues)
		propertyRes.valuesPtr = uint64(uintptr(unsafe.Pointer(&values[0])))
	}

	if propertyRes.countEnumBlobs > 0 {
		enumBlobs = make([]sysPropertyEnum, propertyRes.countEnumBlobs)
		propertyRes.enumBlobPtr = uint64(uintptr(unsafe.Pointer(&enumBlobs[0])))
	}

	// Repeat the ioctl command to fill the arrays.
	err = ioctl.Do(uintptr(file.Fd()), uintptr(IOCTLModeGetProperty),
		uintptr(unsafe.Pointer(propertyRes)))
	if err != nil {
		return nil, err
	}

	// Create enum value array for output.
	enums := make([]PropertyEnum, propertyRes.countEnumBlobs)
	for i, enumBlob := range enumBlobs {
		// Name is null termninated, so we remove the trailing 0s.
		name, _, _ := bytes.Cut(enumBlob.name[:], []byte{0})
		enums[i] = PropertyEnum{
			Value: enumBlob.value,
			Name:  string(name),
		}
	}

	// Name is null termninated, so we remove the trailing 0s.
	name, _, _ := bytes.Cut(propertyRes.name[:], []byte{0})
	return &Property{
		ID:        propertyRes.propId,
		Values:    values,
		EnumBlobs: enums,
		Flags:     propertyRes.flags,
		Name:      string(name),
	}, nil
}

func GetBlob(file *os.File, id uint32) (*Blob, error) {
	propertyBlob := &sysGetBlob{blobId: id, length: 0}
	err := ioctl.Do(uintptr(file.Fd()), uintptr(IOCTLModeGetPropBlob),
		uintptr(unsafe.Pointer(propertyBlob)))
	if err != nil {
		return nil, err
	}

	// Create array to store the blob data.
	var (
		data []uint8
	)

	if propertyBlob.length > 0 {
		data = make([]uint8, propertyBlob.length)
		propertyBlob.data = uint64(uintptr(unsafe.Pointer(&data[0])))
	}

	// Repeat the ioctl command to fill the array.
	err = ioctl.Do(uintptr(file.Fd()), uintptr(IOCTLModeGetPropBlob),
		uintptr(unsafe.Pointer(propertyBlob)))
	if err != nil {
		return nil, err
	}

	return &Blob{
		ID:   id,
		Data: data,
	}, nil
}

func CreateBlob(file *os.File, data []uint8) (uint32, error) {
	createBlob := &sysCreateBlob{
		data:   uint64(uintptr(unsafe.Pointer(&data[0]))),
		length: uint32(len(data)),
	}
	err := ioctl.Do(uintptr(file.Fd()), uintptr(IOCTLModeCreateBlob),
		uintptr(unsafe.Pointer(createBlob)))
	if err != nil {
		return 0, err
	}
	return createBlob.blobId, nil
}

func CreateInfoBlob(file *os.File, info Info) (uint32, error) {
	createBlob := &sysCreateBlob{
		data:   uint64(uintptr(unsafe.Pointer(&info))),
		length: uint32(unsafe.Sizeof(info)),
	}
	err := ioctl.Do(uintptr(file.Fd()), uintptr(IOCTLModeCreateBlob),
		uintptr(unsafe.Pointer(createBlob)))
	if err != nil {
		return 0, err
	}
	return createBlob.blobId, nil
}

func DestroyBlob(file *os.File, id uint32) error {
	destroyBlob := &sysDestroyBlob{
		blobId: id,
	}
	return ioctl.Do(uintptr(file.Fd()), uintptr(IOCTLModeDestroyBlob),
		uintptr(unsafe.Pointer(destroyBlob)))
}

func SetClientCap(file *os.File, capability, value uint64) error {
	setClientCap := &sysSetClientCap{
		capability: capability,
		value:      value,
	}
	err := ioctl.Do(uintptr(file.Fd()), uintptr(IOCTLSetClientCap),
		uintptr(unsafe.Pointer(setClientCap)))
	return err
}

func GetProperties(file *os.File, objectID uint32, objectType uint32) (*Properties, error) {
	objGetProperties := &sysObjGetProperties{}
	objGetProperties.objID = objectID
	objGetProperties.objType = objectType
	err := ioctl.Do(uintptr(file.Fd()), uintptr(IOCTLModeObjGetProperties),
		uintptr(unsafe.Pointer(objGetProperties)))
	if err != nil {
		return nil, err
	}

	var (
		props      []uint32
		propValues []uint64
	)

	if objGetProperties.countProps > 0 {
		props = make([]uint32, objGetProperties.countProps)
		objGetProperties.propsPtr = uint64(uintptr(unsafe.Pointer(&props[0])))

		propValues = make([]uint64, objGetProperties.countProps)
		objGetProperties.propValuesPtr = uint64(uintptr(unsafe.Pointer(&propValues[0])))
	}

	err = ioctl.Do(uintptr(file.Fd()), uintptr(IOCTLModeObjGetProperties),
		uintptr(unsafe.Pointer(objGetProperties)))
	if err != nil {
		return nil, err
	}

	ret := &Properties{
		ObjectID:   objectID,
		ObjectType: objectType,
	}

	ret.Props = make([]uint32, len(props))
	copy(ret.Props, props)
	ret.PropValues = make([]uint64, len(propValues))
	copy(ret.PropValues, propValues)

	return ret, nil
}

func Atomic(file *os.File, flags uint32, atomicProperties []AtomicProperty) error {
	// There is nothing to do if no properties are specified.
	if len(atomicProperties) == 0 {
		return nil
	}

	// Sort the properties in the input according to the object id.
	properties := make([]AtomicProperty, len(atomicProperties))
	copy(properties, atomicProperties)
	slices.SortFunc[[]AtomicProperty](properties, func(a, b AtomicProperty) int {
		if a.ObjectID < b.ObjectID {
			return -1
		} else if a.ObjectID == b.ObjectID {
			return 0
		} else {
			return 1
		}
	})

	// Create individual arrays required by the syscall structure.
	objs := make([]uint32, len(atomicProperties))
	countProps := make([]uint32, len(atomicProperties))
	props := make([]uint32, len(atomicProperties))
	propValues := make([]uint64, len(atomicProperties))
	// Set the first object ID.
	objsIndex := 0
	objs[0] = properties[0].ObjectID
	curObj := objs[0]
	countPropsObj := uint32(0)
	for i, property := range properties {
		if property.ObjectID != curObj {
			// Set the number of properties for the current object.
			countProps[objsIndex] = countPropsObj
			// Advance to the next object.
			countPropsObj = 0
			objsIndex++
			objs[objsIndex] = property.ObjectID
		}
		// Set the property id and value.
		props[i] = property.PropertyID
		propValues[i] = property.Value
		// Increase the number of properties for the object.
		countPropsObj++
	}
	// Set the property count for the last object.
	countProps[objsIndex] = countPropsObj

	// Assemble atomic request.
	sysAtomic := &sysAtomic{
		flags:         flags,
		countObjs:     uint32(objsIndex) + 1,
		objsPtr:       uint64(uintptr(unsafe.Pointer(&objs[0]))),
		countPropsPtr: uint64(uintptr(unsafe.Pointer(&countProps[0]))),
		propValuesPtr: uint64(uintptr(unsafe.Pointer(&propValues[0]))),
		propsPtr:      uint64(uintptr(unsafe.Pointer(&props[0]))),
	}
	return ioctl.Do(uintptr(file.Fd()), uintptr(IOCTLModeAtomic),
		uintptr(unsafe.Pointer(sysAtomic)))
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

func AddFB2SinglePlane(file *os.File, width, height uint16,
	pixelFormat uint32, flags, pitch, offset, boHandle uint32, modifier uint64) (uint32, error) {
	f := &sysFBCmd2{}
	f.width = uint32(width)
	f.height = uint32(height)
	f.pixelFormat = pixelFormat
	f.flags = flags
	f.handles[0] = boHandle
	f.pitches[0] = pitch
	f.offsets[0] = offset
	f.modifier[0] = modifier
	err := ioctl.Do(uintptr(file.Fd()), uintptr(IOCTLModeAddFB2),
		uintptr(unsafe.Pointer(f)))
	if err != nil {
		return 0, err
	}
	return f.fbID, nil
}

func AddFB2(file *os.File, width, height uint16,
	pixelFormat uint32, flags uint32, pitches, offsets, boHandles []uint32, modifier []uint64) (uint32, error) {
	f := &sysFBCmd2{}
	f.width = uint32(width)
	f.height = uint32(height)
	f.pixelFormat = pixelFormat
	f.flags = flags
	copy(f.handles[:], boHandles)
	copy(f.pitches[:], pitches)
	copy(f.offsets[:], offsets)
	copy(f.modifier[:], modifier)
	err := ioctl.Do(uintptr(file.Fd()), uintptr(IOCTLModeAddFB2),
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
		crtc.setConnectorsPtr = uint64(uintptr(unsafe.Pointer(connectors)))
	}
	crtc.countConnectors = uint32(count)
	if mode != nil {
		crtc.mode = *mode
		crtc.modeValid = 1
	}
	return ioctl.Do(uintptr(file.Fd()), uintptr(IOCTLModeSetCrtc),
		uintptr(unsafe.Pointer(crtc)))
}
