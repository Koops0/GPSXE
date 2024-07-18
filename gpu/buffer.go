package gpu

import (
	"github.com/go-gl/gl/v4.6-core/gl"
	"unsafe"
    "fmt"
)

const VERTEX_BUFFER_LEN uint32 = 64 * 1024

type Buffer[T any] struct {
	Object 	uint32
	Map 	*T
}

func (b *Buffer[T]) New() Buffer[T] {
	object := uint32(0)
	var memory *T

	gl.GenBuffers(1, &object)
	gl.BindBuffer(gl.ARRAY_BUFFER, uint32(object))

	elementSize := uint32(unsafe.Sizeof(*new(T)))
    bufferSize := elementSize * VERTEX_BUFFER_LEN

	access := uint32(gl.MAP_WRITE_BIT | gl.MAP_PERSISTENT_BIT)

	//allocate buffer
	gl.BufferStorage(gl.ARRAY_BUFFER, int(bufferSize), nil, access)

	//map buffer
	ptr := gl.MapBufferRange(gl.ARRAY_BUFFER, 0, int(bufferSize), access)
	if ptr == nil {
    	fmt.Println("Failed to map buffer")
    	return Buffer[T]{}
	}

	// Convert unsafe.Pointer to *T
	memory = (*T)(ptr)

	//reset to 0
	s := unsafe.Slice(memory, int(VERTEX_BUFFER_LEN))

	for x := range s {
		s[x] = *new(T)
	}

	return Buffer[T]{
		Object: object,
		Map: memory,
	}
}

func (b *Buffer[T]) Set(index uint32, value T) {
	if index >= VERTEX_BUFFER_LEN {
		fmt.Println("Index out of range")
		return
	}
	
	// Convert the map (which is a pointer to the start of the buffer) to a slice for easier indexing
    s := unsafe.Slice(b.Map, int(VERTEX_BUFFER_LEN))

    // Set the value at the specified index
    s[index] = value
}

func (b *Buffer[T]) Drop(){
	gl.BindBuffer(gl.ARRAY_BUFFER, b.Object)
	gl.UnmapBuffer(gl.ARRAY_BUFFER)
	gl.DeleteBuffers(1, &b.Object)
}