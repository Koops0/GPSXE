package biosmap

import (
	"../bios"
)

type Range struct {
	address uint32
	bit     uint32
}

var BIOS = Range{
	address: 0xbfc00000,
	bit:     512 * 1024,
}

func (r Range) Contains(addr uint32) *uint32 {
	if addr >= r.address && addr < r.address+r.bit {
		option := addr - r.address
		return &option
	} else {
		return nil
	}
}

type Interconnect struct {
	bios bios.BIOS
}

func (i Interconnect) New(bios *bios.BIOS) Interconnect {
	i.bios = *bios
	return i
}

func (i *Interconnect) Load32(addr uint32) uint32 {
	if offset := BIOS.Contains(addr); offset != nil {
		return i.bios.Load32(*offset)
	} else {
		return 0
	}
}