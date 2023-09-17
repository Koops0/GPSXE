package biosmap

import (
	"fmt"

	"../bios"
	"../ram"
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

var RAM = Range{
	address: 0xa0000000,
	bit:     2 * 1024 * 1024,
}

var REGION_MASK = [8]uint32{0xffffffff, 0xffffffff, 0xffffffff, 0xffffffff, //KUSEG : 2048MB
	// KSEG0 : 512MB
	0x7fffffff,
	// KSEG1 : 512MB
	0x1fffffff,
	// KSEG2 : 1024MB
	0xffffffff, 0xffffffff}

var SPU = Range{
	address: 0x1f801c00,
	bit:     640,
}

var EX2 = Range{
	address: 0x1f802000,
	bit:     66,
}

var EX1 = Range{
	address: 0x1f000000,
	bit:     66,
}

var IRQ_CONTROL = Range{
	address: 0x1f801070,
	bit:     8,
}

var TIMERS = Range{
	address: 0x1f801104,
	bit:     8,
}

var DMA = Range{
	address: 0x1f8010f0,
	bit:     0x80,
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
	ram  ram.RAM
}

func (i Interconnect) New(bios *bios.BIOS) Interconnect {
	i.bios = *bios
	return i
}

func (i *Interconnect) Load32(addr uint32) uint32 { //load 32-bit at addr
	abaddr := Mask_region(addr)

	if addr%4 != 0 {
		panic("error!!!!")
	} else if offset := BIOS.Contains(addr); offset != nil {
		return i.bios.Load32(*offset)
	} else if offset := IRQ_CONTROL.Contains(addr); offset != nil {
		fmt.Println("IRQ Control")
		return 0
	} else if offset := DMA.Contains(abaddr); offset != nil {
		panic(fmt.Sprintf("Unhandled Load32 at address: 0x%08x", addr))
	} else {
		return 0
	}
}

func (i *Interconnect) Load8(addr uint32) uint8 {
	abaddr := Mask_region(addr)

	if offset := RAM.Contains(abaddr); offset != nil {
		return i.ram.Load8(*offset)
	}

	if offset := BIOS.Contains(abaddr); offset != nil {
		return i.bios.Load8(*offset)
	}

	if offset := EX1.Contains(abaddr); offset == nil {
		return 0xff
	}

	panic(fmt.Sprintf("Unhandled Load8 at address: 0x%08x", addr))
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
	} else if offset := IRQ_CONTROL.Contains(addr); offset != nil {
		fmt.Println("IRQ Control")
		return
	}

	panic(fmt.Sprintf("Unhandled store32 into address: 0x%08x", addr))
}

func (i *Interconnect) Store16(addr uint32, val uint16) {
	if addr%4 != 0 {
		panic(fmt.Sprintf("Unhandled Store16 address: 0x%08x", addr))
	}

	abaddr := Mask_region(addr)

	if offset := SPU.Contains(abaddr); offset != nil {
		panic(fmt.Sprintf("Unhandled Write to Register: 0x%08x", val))
	} else if offset := TIMERS.Contains(abaddr); offset != nil {
		panic(fmt.Sprintf("Unhandled Write to Timer Reg: 0x%08x", offset))
	} else if offset := RAM.Contains(abaddr); offset != nil {
		i.ram.Store16(*offset, val)
		return
	}

	panic(fmt.Sprintf("Unhandled Store16 into address: 0x%08x", addr))
}

func (i *Interconnect) Store8(addr uint32, val uint8) {
	abaddr := Mask_region(addr)

	if offset := RAM.Contains(abaddr); offset != nil {
		i.ram.Store8(*offset, val)
	}

	if offset := EX2.Contains(abaddr); offset != nil {
		panic(fmt.Sprintf("Unhandled Write to EX2 Register: 0x%08x", val))
	}

	panic(fmt.Sprintf("Unhandled Store8 into address: 0x%08x", addr))
}

func Mask_region(addr uint32) uint32 {
	index := addr >> 29
	return addr & REGION_MASK[index]
}
