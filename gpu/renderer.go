package gpu

import (
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/go-gl/gl/v4.6-core/gl"
	"modernc.org/libc"
)

// Position struct
type Position struct {
	X int16
	Y int16
}

// Colour struct
type Colour struct {
	R uint8
	G uint8
	B uint8
}

func PFromGP0(val uint32) Position {
    x := int16(val)        
    y := int16(val >> 16) 
    return Position{X: x, Y: y}
}

func CFromGP0(val uint32) Colour {
    r := uint8(val)
	g := uint8(val >> 8)
	b := uint8(val >> 16)
    return Colour{R: r, G: g, B: b}
}

// Renderer struct
type Renderer struct {
	sdlc   error
	window *sdl.Window
	context sdl.GLContext
}

func (r Renderer) New() Renderer {
    
    var initFlags uint32 = sdl.INIT_EVERYTHING

    if err := sdl.Init(initFlags); err != nil {
		// Handle the error, for example, log it, return it, etc.
		fmt.Printf("SDL could not initialize! SDL_Error: %s\n", sdl.GetError())
	}else{
		fmt.Println("SDL initialized")
	}
    r.sdlc = sdl.Init(initFlags)

	// Set Attributes
	sdl.GLSetAttribute(sdl.GLattr(sdl.GL_CONTEXT_MAJOR_VERSION), 4)
	sdl.GLSetAttribute(sdl.GLattr(sdl.GL_CONTEXT_MINOR_VERSION), 4)

	window, err := sdl.CreateWindow("Go PSX Emulator", int32(sdl.WINDOWPOS_CENTERED), 
				int32(sdl.WINDOWPOS_CENTERED), 1024, 512, uint32(sdl.WINDOW_OPENGL))
    if err != nil {
        fmt.Printf("Window could not be created! SDL_Error: %s\n", sdl.GetError())
    }

	r.window = window
    r.context, err = window.GLCreateContext()
    if err != nil {
        fmt.Printf("OpenGL context could not be created! SDL_Error: %s\n", sdl.GetError())
    }

	//Load GL
	if err := gl.InitWithProcAddrFunc(sdl.GLGetProcAddress); err != nil {
		fmt.Println(err)
	}

	//fully init 
	gl.Init()
	gl.ClearColor(0.0, 0.0, 0.0, 1.0)
	gl.Clear(gl.COLOR_BUFFER_BIT)

	window.GLSwap()

	return r
}