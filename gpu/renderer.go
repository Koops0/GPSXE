package gpu

import (
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/go-gl/gl/v4.6-core/gl"
	//"modernc.org/libc"
)

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
	return r
}