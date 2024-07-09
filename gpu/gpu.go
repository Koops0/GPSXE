package gpu

import (
    "fmt"
)

// GPU Variables
type GPU struct {
    PageBaseX               uint8 // 4 bit and 64 byte
    PageBaseY               uint8 // 1 bit and 256 byte
    SemiTransparency        uint8 // blending
    TextureDepth            TextureDepth
    Dithering               bool // 24 to 15 bit
    DrawToDisplay           bool // draw to display
    ForceSetMaskBit         bool // force bit to 1
    PreserveMaskedPixels    bool // preserve masked pixels
    Field                   Field // current field
    TextureDisable          bool // disable texture
    HRes                    HorizontalRes
    VRes                    VerticalRes
    VMode                   VideoMode
    DisplayDepth            DisplayDepth
    Interlaced              bool // Toggle for interlaced and progressive modes
    DisplayDisabled         bool // Disable display
    Interrupt               bool // Interrupt
    DmaDir                  DMADirection // DMA direction
    RectangleTextureXFlip   bool // Flip texture x
    RectangleTextureYFlip   bool // Flip texture y
    TextureWindowXMask      uint8
    TextureWindowYMask      uint8
    TextureWindowXOffset    uint8
    TextureWindowYOffset    uint8
    DrawingAreaLeft         uint16
    DrawingAreaTop          uint16
    DrawingAreaRight        uint16
    DrawingAreaBottom       uint16
    DrawingXOffset          int16
    DrawingYOffset          int16
    DisplayVRAMXStart       uint16
    DisplayVRAMYStart       uint16
    DisplayHorizStart       uint16
    DisplayHorizEnd         uint16
    DisplayLineStart        uint16
    DisplayLineEnd          uint16
	Gp0Command 				CommandBuffer
	Gp0WordsRemaining 	    uint32
	Gp0CommandMethod 		func(*GPU)
    Gp0Mode                 Gp0Mode
}

type TextureDepth uint8

const (
    T4Bit  TextureDepth = iota
    T8Bit
    T15Bit
)

type Field uint8

const (
    Top Field = iota + 1
    Bottom
)

type HorizontalRes struct {
    Val uint8
}

// NewHorizontalRes creates a new HorizontalRes instance from the 2-bit field 'hr1' and the 1-bit field 'hr2'.
func NewHorizontalRes(hr1, hr2 uint8) HorizontalRes {
    hr := (hr2 & 1) | ((hr1 & 3) << 1)
    return HorizontalRes{Val: hr}
}

// IntoStatus retrieves the value of bits [18:16] of the status register.
func (hr HorizontalRes) IntoStatus() uint32 {
    return uint32(hr.Val) << 16
}

type VerticalRes uint8

const (
    Y240Lines VerticalRes = iota
    Y480Lines // Interlaced Only
)

type VideoMode uint8

const (
    NTSC VideoMode = iota // 480i 60Hz
    PAL                    // 576i 50Hz
)

type DisplayDepth uint8

const (
    D15Bit DisplayDepth = iota
    D24Bit
)

type DMADirection uint8

const (
    Off DMADirection = iota
    FIFO
    CPUGP0
    VRAMCPU
)

type Gp0Mode uint8

const (
    Command Gp0Mode = iota
    ImageLoad
)

// NewGPUInstance initializes a new GPU instance with default values.
func (g *GPU) NewGPUInstance() GPU {
    g.PageBaseX = 0
    g.PageBaseY = 0
    g.SemiTransparency = 0
    g.TextureDepth = T4Bit
    g.Dithering = false
    g.DrawToDisplay = false
    g.ForceSetMaskBit = false
    g.PreserveMaskedPixels = false
    g.Field = Top
    g.TextureDisable = false
    g.HRes = NewHorizontalRes(0, 0)
    g.VRes = Y240Lines
    g.VMode = NTSC
    g.DisplayDepth = D15Bit
    g.Interlaced = false
    g.DisplayDisabled = true
    g.Interrupt = false
    g.DmaDir = Off
    g.RectangleTextureXFlip = false
    g.RectangleTextureYFlip = false
	g.TextureWindowXMask = 0
	g.TextureWindowYMask = 0
	g.TextureWindowXOffset = 0
	g.TextureWindowYOffset = 0
	g.DrawingAreaLeft = 0
	g.DrawingAreaTop = 0
	g.DrawingAreaRight = 0
	g.DrawingAreaBottom = 0
	g.DrawingXOffset = 0
	g.DrawingYOffset = 0
	g.DisplayVRAMXStart = 0
	g.DisplayVRAMYStart = 0
	g.DisplayHorizStart = 0
	g.DisplayHorizEnd = 0
	g.DisplayLineStart = 0
	g.DisplayLineEnd = 0
	g.Gp0Command = CommandBuffer{}.New()
	g.Gp0WordsRemaining = 0
	g.Gp0CommandMethod = nil
    return *g
}

func boolToUint32(b bool) uint32 {
    if b {
        return 1
    }
    return 0
}

func (g *GPU) Status() uint32 {
    r := uint32(0)
    r |= uint32(g.PageBaseX) << 0
    r |= uint32(g.PageBaseY) << 4
    r |= uint32(g.SemiTransparency) << 5
    r |= uint32(g.TextureDepth) << 7
    r |= boolToUint32(g.Dithering) << 9
    r |= boolToUint32(g.DrawToDisplay) << 10
    r |= boolToUint32(g.ForceSetMaskBit) << 11
    r |= boolToUint32(g.PreserveMaskedPixels) << 12
    r |= uint32(g.Field) << 13
    // Bit 14: not supported
    r |= boolToUint32(g.TextureDisable) << 15
    r |= g.HRes.IntoStatus()
    //r |= uint32(g.VRes) << 19
    r |= uint32(g.VMode) << 20
    r |= uint32(g.DisplayDepth) << 21
    r |= boolToUint32(g.Interlaced) << 22
    r |= boolToUint32(g.DisplayDisabled) << 23
    r |= boolToUint32(g.Interrupt) << 24

    // For now, pretend that GPU is always ready:
    // Receive
    r |= 1 << 26
    // Send VRAM
    r |= 1 << 27
    // Receive Block
    r |= 1 << 28
    r |= uint32(g.DmaDir) << 29
    // 31 should change depending on the drawn line
    r |= 0 << 31

    // DMA Request
    var dmaReq uint32
    switch g.DmaDir {
    case FIFO:
        dmaReq = 1
    case CPUGP0:
        dmaReq = (r >> 28) & 1
    case VRAMCPU:
        dmaReq = (r >> 27) & 1
    default: // Off
        dmaReq = 0
    }

    r |= dmaReq << 25
    return r
}

type Gp0Method func(*GPU, uint32)

func (g *GPU) Gp0(val uint32) {
    if g.Gp0WordsRemaining == 0 {
        opcode := (val >> 24) & 0xFF

        var method Gp0Method
        var len uint32

        switch opcode {
        case 0x00:
            len, method = 1, Gp0NopWrapper
        case 0x28:
            len, method = 5, Gp0QuadMonoOpaqueWrapper
        case 0x2C:
            len, method = 4, Gp0QTBlendOpaqueWrapper
        case 0x30:
            len, method = 6, Gp0TriShadedOpaqueWrapper
        case 0x38:
            len, method = 8, Gp0QuadShadedOpaqueWrapper
        case 0xA0:
            len, method = 3, Gp0ImgLoadWrapper
        case 0xE1:
            len, method = 1, Gp0DrawModeWrapper
        case 0xE2:
            len, method = 1, Gp0TexWindowWrapper
        case 0xE3:
            len, method = 1, Gp0DrawAreaTLWrapper
        case 0xE4:
            len, method = 1, Gp0DrawAreaBRWrapper
        case 0xE5:
            len, method = 1, Gp0DrawOffsetWrapper
        case 0xE6:
            len, method = 1, Gp0MaskBitSettingWrapper
        default:
            panic(fmt.Sprintf("Unhandled GP0 command: 0x%X", opcode))
        }

        g.Gp0WordsRemaining = len
		g.Gp0CommandMethod = func(g *GPU) {
    		method(g, val)
		}

        g.Gp0Command.Clear()
    }

    g.Gp0WordsRemaining--

    switch g.Gp0Mode {
    case Command:
        g.Gp0Command.Push(val)
        if g.Gp0WordsRemaining == 0 {
            g.Gp0CommandMethod(g)
        }
    case ImageLoad:
        ///Copy Pixel Data
        if g.Gp0WordsRemaining == 0 {
            g.Gp0Mode = Command
        }
    }
}

func (g *GPU) Gp0Nop() { //0x00
}

func (g *GPU) Gp0QuadMonoOpaque() { //0x28
	fmt.Println("Draw Quad")
}

func (g *GPU) Gp0QTBlendOpaque() { //0x2C
    fmt.Println("Draw Quad")
}

func (g *GPU) Gp0TriShadedOpaque() { //0x30
    fmt.Println("Draw Triangle")
}

func (g *GPU) Gp0QuadShadedOpaque() { //0x38
    fmt.Println("Draw Quad")
}

func (g *GPU) Gp0ImgLoad(val uint32) { //0xA0
    res := g.Gp0Command.Index(2)
    w := (res & 0xFFFF)
    h := res >> 16
    imgsize := w * h
    imgsize = (imgsize + 1) & ^uint32(1)
    g.Gp0WordsRemaining = imgsize/2
    g.Gp0Mode = ImageLoad
}

func (g *GPU) Gp0ImgStore(val uint32) { //0xC0
    res := g.Gp0Command.Index(2)
    w := (res & 0xFFFF)
    h := res >> 16
    fmt.Println(w, h)
}

func (g *GPU) Gp0DrawMode(val uint32) { //0xE1
    g.PageBaseX = uint8(val & 0xF)
    g.PageBaseY = uint8((val >> 4) & 1)
    g.SemiTransparency = uint8((val >> 5) & 3)

    switch (val >> 7) & 3 {
    case 0:
        g.TextureDepth = T4Bit
    case 1:
        g.TextureDepth = T8Bit
    case 2:
        g.TextureDepth = T15Bit
    default:
        panic(fmt.Sprintf("Invalid texture depth: %d", (val >> 7) & 3))
    }

    g.Dithering = (val >> 9) & 1 != 0
    g.DrawToDisplay = (val >> 10) & 1 != 0
    g.TextureDisable = (val >> 11) & 1 != 0
    g.RectangleTextureXFlip = (val >> 12) & 1 != 0
    g.RectangleTextureYFlip = (val >> 13) & 1 != 0
}

func (g *GPU) Gp0TexWindow(val uint32){ //0xE2
	g.TextureWindowXMask = uint8(val & 0x1F)
	g.TextureWindowYMask = uint8((val >> 5) & 0x1F)
	g.TextureWindowXOffset = uint8((val >> 10) & 0x1F)
	g.TextureWindowYOffset = uint8((val >> 15) & 0x1F)
}

func (g *GPU) Gp0DrawAreaTL(val uint32){ //0xE3
	g.DrawingAreaLeft = uint16(val & 0x3FF)
	g.DrawingAreaTop = uint16((val >> 10) & 0x3FF)
}

func (g *GPU) Gp0DrawAreaBR(val uint32){ //0xE4
	g.DrawingAreaRight = uint16(val & 0x3FF)
	g.DrawingAreaBottom = uint16((val >> 10) & 0x3FF)
}

func (g *GPU) Gp0DrawOffset(val uint32){ //0xE5
	x := int16(val & 0x7FF)
	y := int16((val >> 11) & 0x7FF)

	// Sign extend to 16 bits
	g.DrawingXOffset = int16(x << 5) >> 5
	g.DrawingYOffset = int16(y << 5) >> 5
}

func (g *GPU) Gp0MaskBitSetting(val uint32){ //0xE6
	g.ForceSetMaskBit = (val & 1) != 0
	g.PreserveMaskedPixels = (val & 2) != 0
}

func (g *GPU) Gp0ClearCache() {
    
}

func (g *GPU) Gp1 (val uint32) {
	opcode := val & 0xFF
	switch opcode {
	case 0x00:
		// Reset GPU
		g.Gp1Reset()
	default:
		panic(fmt.Sprintf("Invalid GP1 opcode: 0x%X", opcode))
	}
}

func (g *GPU) Gp1Reset() {
	g.Interrupt = false

	g.PageBaseX = 0
	g.PageBaseY = 0
	g.SemiTransparency = 0
	g.TextureDepth = T4Bit
	g.TextureWindowXMask = 0
	g.TextureWindowYMask = 0
	g.TextureWindowXOffset = 0
	g.TextureWindowYOffset = 0
	g.Dithering = false
	g.DrawToDisplay = false
	g.TextureDisable = false
	g.RectangleTextureXFlip = false
	g.RectangleTextureYFlip = false
	g.DrawingAreaLeft = 0
	g.DrawingAreaTop = 0
	g.DrawingAreaRight = 0
	g.DrawingAreaBottom = 0
	g.DrawingXOffset = 0
	g.DrawingYOffset = 0
	g.ForceSetMaskBit = false
	g.PreserveMaskedPixels = false

	g.DmaDir = Off

	g.DisplayDisabled = true
	g.DisplayVRAMXStart = 0
	g.DisplayVRAMYStart = 0
	g.HRes = NewHorizontalRes(0, 0)
	g.VRes = Y240Lines

	g.VMode = NTSC
	g.Interlaced = true
	g.DisplayHorizStart = 0x200
	g.DisplayHorizEnd = 0xc00
	g.DisplayLineStart = 0x10
	g.DisplayLineEnd = 0x100
	g.DisplayDepth = D15Bit
	
	//clear fifo and gpu
}

func (g *GPU) Gp1DisplayMode(val uint32){ //0x80
	hr1 := uint8(val & 3)
	hr2 := uint8((val >> 6) & 1)

	g.HRes = NewHorizontalRes(hr1, hr2)

	switch (val & 0x4) != 0 {
	case true:
		g.VRes = Y480Lines
	case false:
		g.VRes = Y240Lines
	}

	switch (val & 0x8) != 0 {
	case true:
		g.VMode = PAL
	case false:
		g.VMode = NTSC
	}

	switch (val & 0x10) != 0 {
	case true:
		g.DisplayDepth = D15Bit
	case false:
		g.DisplayDepth = D24Bit
	}

	g.Interlaced = (val & 0x20) != 0

	if (val & 0x80) != 0 {
		panic("Unsupported display mode")
	}
}

func (g *GPU) Gp1ResetCommBuffer(val uint32){ //0x01
    g.Gp0Command.Clear()
    g.Gp0WordsRemaining = 0
    g.Gp0Mode = Command
}

func (g *GPU) Gp1AcknowledgeIRQ(){ //0x02
    g.Interrupt = false
}

func (g *GPU) Gp1DisplayEnable(val uint32){ //0x03
    g.DisplayDisabled = (val & 1) != 0
}

func (g *GPU) Gp1DMADir(val uint32){ //0x04
	switch val & 3 {
	case 0:
		g.DmaDir = Off
	case 1:
		g.DmaDir = FIFO
	case 2:
		g.DmaDir = CPUGP0
	case 3:
		g.DmaDir = VRAMCPU
	default:
		panic("Invalid DMA direction")
	}
}

func (g *GPU) Gp1DisplayVRAMStart(val uint32){ //0x05
	g.DisplayVRAMXStart = uint16(val & 0x3FE)
	g.DisplayVRAMYStart = uint16((val >> 10) & 0x1FF)
}

func (g *GPU) Gp1DisplayHRange(val uint32){ //0x06
	g.DisplayHorizStart = uint16(val & 0xFFF)
	g.DisplayHorizEnd = uint16((val >> 12) & 0xFFF)
}

func (g *GPU) Gp1DisplayVRange(val uint32){ //0x07
	g.DisplayLineStart = uint16(val & 0x3FF)
	g.DisplayLineEnd = uint16((val >> 10) & 0x3FF)
}

func (g *GPU) Read() uint32{
	return 0
}

func Gp0NopWrapper(g *GPU, val uint32) {
    g.Gp0Nop()
}
func Gp0QuadMonoOpaqueWrapper(g *GPU, val uint32) {
    // Assuming Gp0QuadMonoOpaque is a method that requires the val parameter
    g.Gp0QuadMonoOpaque()
}

func Gp0QTBlendOpaqueWrapper(g *GPU, val uint32) {
    g.Gp0QTBlendOpaque()
}

func Gp0TriShadedOpaqueWrapper(g *GPU, val uint32) {
    g.Gp0TriShadedOpaque()
}

func Gp0QuadShadedOpaqueWrapper(g *GPU, val uint32) {
    g.Gp0QuadShadedOpaque()
}

func Gp0ImgLoadWrapper(g *GPU, val uint32) {
    g.Gp0ImgLoad(val)
}

func Gp0DrawModeWrapper(g *GPU, val uint32) {
    g.Gp0DrawMode(val)
}

func Gp0TexWindowWrapper(g *GPU, val uint32) {
    g.Gp0TexWindow(val)
}

func Gp0DrawAreaTLWrapper(g *GPU, val uint32) {
    g.Gp0DrawAreaTL(val)
}

func Gp0DrawAreaBRWrapper(g *GPU, val uint32) {
    g.Gp0DrawAreaBR(val)
}

func Gp0DrawOffsetWrapper(g *GPU, val uint32) {
    g.Gp0DrawOffset(val)
}

func Gp0MaskBitSettingWrapper(g *GPU, val uint32) {
    g.Gp0MaskBitSetting(val)
}