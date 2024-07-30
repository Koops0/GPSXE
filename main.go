package main

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/go-gl/gl/v4.6-core/gl"

	"github.com/Koops0/GPSXE/bios"
	"github.com/Koops0/GPSXE/biosmap"
	"github.com/Koops0/GPSXE/gpu"
)

func main() {
	bios, err := bios.New("SCPH1001.bin") //will switch to SCPH7501.bin later
	if err != nil {
		fmt.Println("Error reading file")
		return
	}

	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		fmt.Println("Error initializing SDL:", err)
		return
	}
	defer sdl.Quit()

	// Create an SDL renderer for the window
	renderer := gpu.Renderer{}.New()
	gpu := gpu.GPU{}.New(renderer)
	inter := biosmap.Interconnect{}.New(bios, gpu)
	cpu := &CPU{}
	cpu.New(inter)
	fmt.Println(cpu.reg[0])
    debugger := Debugger{}

	for {
        for i := 0; i < 1000000; i++ {
            cpu.Run_next(debugger)
        }
        for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
            switch e := event.(type) {
            case *sdl.KeyboardEvent:
                if e.Type == sdl.KEYDOWN {
                    switch e.Keysym.Sym {
                    case sdl.K_PAUSE:
                        debugger.Debug(*cpu)
                    case sdl.K_ESCAPE:
                        return
                    }
                }
            case *sdl.QuitEvent:
                return
            }
        }
    }
}

func CheckForErrors() {
    fatal := false

    for {
        buffer := make([]byte, 4096)
        severity := uint32(0)
        source := uint32(0)
        mSize := int32(0)
        mtype := uint32(0)
        id := uint32(0)
        count := gl.GetDebugMessageLog(1, int32(len(buffer)), &source, &mtype, &id, &severity, &mSize, &buffer[0])

        if count == 0 {
            break
        }

        // Assuming mSize is the actual message length, trim the buffer to mSize
        message := string(buffer[:mSize])

        fmt.Printf("OpenGL [source: %d | type: %d | id: 0x%x | severity: %d] %s\n", source, mtype, id, severity, message)

        // Example severity check (adjust according to your severity values)
        if severity == gl.DEBUG_SEVERITY_HIGH {
            fatal = true
            fmt.Println("Fatal OpenGL error encountered.")
            break // or handle fatal error as needed
        }
    }

    if fatal {
        // Handle fatal error, e.g., clean up and exit or throw panic
        panic("Fatal OpenGL error encountered.")
    }
}