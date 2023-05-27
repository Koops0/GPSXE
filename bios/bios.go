package bios

import (
	"os"
)

const BIOS_SIZE uint64 = 512 * 1024

type BIOS struct {
	data []uint8
}

func (b BIOS) New(path string) (*BIOS, error) {
	file, err := os.Open(path)

	if err != nil {
		return nil, err
	}
	defer file.Close()

	b.data = make([]uint8, 0)

	_, err = file.Read(b.data)

	if err != nil {
		return nil, err
	}

	if len(b.data) == int(BIOS_SIZE) {
		return &BIOS{data: b.data}, nil
	} else {
		return nil, err
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
