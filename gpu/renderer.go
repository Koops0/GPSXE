package gpu

import (
	"fmt"
	"os"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/veandco/go-sdl2/sdl"
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
	sdlc           error
	window         *sdl.Window
	context        sdl.GLContext
	vertexShader   uint32
	fragmentShader uint32
	program        uint32
	vao            uint32
	positions      Buffer[Position]
	colours        Buffer[Colour]
	nVertices      uint32
}

func (r Renderer) New() Renderer {

	var initFlags uint32 = sdl.INIT_EVERYTHING

	if err := sdl.Init(initFlags); err != nil {
		// Handle the error, for example, log it, return it, etc.
		fmt.Printf("SDL could not initialize! SDL_Error: %s\n", sdl.GetError())
	} else {
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

	//Slurp contents of vs and fs
	vs, err := os.ReadFile("ps1.vs")
	if err != nil {
		fmt.Println("Failed to load vertex shader")
	}
	fs, err := os.ReadFile("ps1.fs")
	if err != nil {
		fmt.Println("Failed to load fragment shader")
	}

	// Convert the file contents to a string
	vsSrc := string(vs)
	fsSrc := string(fs)

	vertexShader := CompileShader(vsSrc, gl.VERTEX_SHADER)
	fragmentShader := CompileShader(fsSrc, gl.FRAGMENT_SHADER)
	program := LinkProgram([]uint32{vertexShader, fragmentShader})
	gl.UseProgram(program)

	r.vertexShader = vertexShader
	r.fragmentShader = fragmentShader
	r.program = program

	//VAO
	vao := uint32(0)
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	r.vao = vao

	//Positions and Colours
	var positionsBuffer *Buffer[Position] = new(Buffer[Position])
	var coloursBuffer *Buffer[Colour] = new(Buffer[Colour])
	positions := positionsBuffer.New()
	colours := coloursBuffer.New()

	index := FindProgAttrib(program, "vertex_position")
	gl.EnableVertexAttribArray(index)
	gl.VertexAttribPointer(index, 2, gl.SHORT, false, 0, nil)

	index = FindProgAttrib(program, "vertex_color")
	gl.EnableVertexAttribArray(index)
	gl.VertexAttribPointer(index, 3, gl.UNSIGNED_BYTE, true, 0, nil)

	r.positions = positions
	r.colours = colours
	r.nVertices = 0

	return r
}

func CompileShader(source string, shaderType uint32) uint32 {
	shader := gl.CreateShader(shaderType)

	//compile
	cStr := gl.Str(source + "\x00")
	gl.ShaderSource(shader, 1, &cStr, nil)
	gl.CompileShader(shader)

	//check
	status := int32(gl.FALSE)
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		panic("Failed to compile shader")
	}
	return shader
}

func LinkProgram(shaders []uint32) uint32 {
	program := gl.CreateProgram()

	for _, shader := range shaders {
		gl.AttachShader(program, shader)
	}

	gl.LinkProgram(program)

	status := int32(gl.FALSE)
	gl.GetShaderiv(program, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		panic("Linkage Failed")
	}

	return program
}

func FindProgAttrib(program uint32, name string) uint32 {
	cStr := gl.Str(name + "\x00")
	index := gl.GetAttribLocation(program, cStr)
	if index < 0 {
		panic("Failed to find attribute")
	}
	return uint32(index)
}

func (r *Renderer) pushTriangle(positions []Position, colours []Colour) {
	if r.nVertices+3 > VERTEX_BUFFER_LEN {
		fmt.Println("Too many vertices, forcing draw")
		r.Draw()
	}

	for i := 0; i < 3; i++ {
		r.positions.Set(r.nVertices, positions[i])
		r.colours.Set(r.nVertices, colours[i])
		r.nVertices++
	}

}

func (r *Renderer) Draw() {
	//flush to buffer
	gl.MemoryBarrier(gl.CLIENT_MAPPED_BUFFER_BARRIER_BIT)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(r.nVertices))

	//wait
	sync := gl.FenceSync(gl.SYNC_GPU_COMMANDS_COMPLETE, 0)

	//during render
	for{
		render := gl.ClientWaitSync(sync, gl.SYNC_FLUSH_COMMANDS_BIT, 10000000)
		if render == gl.ALREADY_SIGNALED || render == gl.CONDITION_SATISFIED {
			break
		}
	}

	r.nVertices = 0
}

func (r *Renderer) Display(){
	gl.Clear(gl.COLOR_BUFFER_BIT)
	r.Draw()
	r.window.GLSwap()
}

func (r *Renderer) Drop(){
	gl.DeleteVertexArrays(1, &r.vao)
	gl.DeleteShader(r.vertexShader)
	gl.DeleteShader(r.fragmentShader)
	gl.DeleteProgram(r.program)
}