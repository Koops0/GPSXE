package bios

import (
	"errors"
    "io"
    "os"
	"log"
)

const BIOS_SIZE uint64 = 512 * 1024

type BIOS struct {
	data []uint8 //Memory
}

func New(path string) (*BIOS, error) { // Load in PSX BIOS
    b := &BIOS{} 

    file, err := os.Open(path)
    if err != nil {
		log.Fatalf("failed to open file!")
        return nil, err
    }
    defer file.Close()

    // Read the file's contents into a byte slice
    data, err := io.ReadAll(file)
    if err != nil {
        return nil, err
    }

    // Check if the read data matches the expected BIOS size
    if len(data) == int(BIOS_SIZE) {
        b.data = data
        return b, nil
    } else {
        return nil, errors.New("incorrect BIOS size")
    }
}

func (b *BIOS) Load32(offset uint32) uint32 {
	b0 := uint32(b.data[offset])
	b1 := uint32(b.data[offset+1])
	b2 := uint32(b.data[offset+2])
	b3 := uint32(b.data[offset+3])

	res := b0 | (b1 << 8) | (b2 << 16) | (b3 << 24)
	return res
}

func (b *BIOS) Load8(offset uint32) uint8 {
	return b.data[offset]
}