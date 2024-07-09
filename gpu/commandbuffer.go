package gpu

import (
	"fmt"
)

type CommandBuffer struct {
	Buffer 	[12]uint32
	Len			int
}

func (cb CommandBuffer) New() CommandBuffer {
    cb.Buffer = [12]uint32{}
    cb.Len = 0
    return cb
}

func (cb *CommandBuffer) Clear() {
	cb.Len = 0
}

func (cb *CommandBuffer) Push(val uint32) {
	if cb.Len >= 12 {
		panic("Command buffer overflow")
	}
	cb.Buffer[cb.Len] = val
	cb.Len++
}

func (cb *CommandBuffer) Index(index int) uint32 {
    if index >= cb.Len {
        panic(fmt.Sprintf("Command buffer index out of range: %d (%d)", index, cb.Len))
    }
    return cb.Buffer[index]
}