package biosmap

import (
	"fmt"

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

var MEM_CONTROL = Range{
	address: 0x1f801000,
	bit:     36,
}

var RAM_SIZE = Range{
	address: 0x1f801060,
	bit:     4,
}

var CACHECONTROL = Range{
	address: 0xfffe0130,
	bit:     4,
}

func (r Range) Contains(addr uint32) *uint32 { //Return offset if it exists
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

func (i *Interconnect) Load32(addr uint32) uint32 { //load 32-bit at addr
	if addr%4 != 0 {
		panic("error!!!!")
	} else if offset := BIOS.Contains(addr); offset != nil {
		return i.bios.Load32(*offset)
	} else {
		return 0
	}
}

func (i *Interconnect) Store32(addr uint32, val uint32) { //Store value in address
	if addr%4 != 0 {
		panic(fmt.Sprintf("Unhandled Store32 address: 0x%08x", addr))
	} else if offset := BIOS.Contains(addr); offset != nil {
		switch *offset {
		case 0: // Expansion 1 base address
			if val != 0x1f00000 {
				panic(fmt.Sprintf("Bad expansion 1 base address: 0x%08x", val))
			}
		case 4: // Expansion 2 base address
			if val != 0x1f802000 {
				panic(fmt.Sprintf("Bad expansion 2 base address: 0x%08x", val))
			}
		default:
			fmt.Println("Unhandled write to MEM_CONTROL register")
		}
		return
	}

	panic(fmt.Sprintf("Unhandled store32 into address: 0x%08x", addr))
}
