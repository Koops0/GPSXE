package biosmap

import (
	"fmt"
	"log"

	"github.com/Koops0/GPSXE/bios"
	"github.com/Koops0/GPSXE/dma"
	"github.com/Koops0/GPSXE/gpu"
	"github.com/Koops0/GPSXE/ram"
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

var TIMERS = Range{ //Change?
	address: 0x1f801104,
	bit:     8,
}

var DMA = Range{
	address: 0x1f8010f0,
	bit:     0x80,
}

var GPU = Range{ //Change?
	address: 0x1f801814,
	bit:     32,
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
	dma  dma.DMA
	gpu  gpu.GPU
}

func (i Interconnect) New(bios *bios.BIOS) Interconnect {
	i.bios = *bios
	i.dma.New()
	return i
}

func (i *Interconnect) Dma_reg(offset uint32) uint32 { //DMA reg read
	major := (offset & 0x70) >> 4
	minor := offset & 0xf

	switch major {
	case 0, 1, 2, 3, 4, 5, 6:
		channel_index := i.dma.From_index(major)
		switch minor {
		case 8:
			return i.dma.Channels[channel_index].Control()
		default:
			panic("Unhandled DMA read")
		}
	case 7:
		switch minor {
		case 0:
			return i.dma.Control()
		case 4:
			return i.dma.Interrupt()
		default:
			panic("Unhandled DMA read")
		}

	default:
		panic("Unhandled DMA read")
	}
}

func (i *Interconnect) Set_dma_reg(offset uint32, val uint32) { //DMA reg write
	major := (offset & 0x70) >> 4
	minor := offset & 0xf

	var active_port dma.Port

	switch major {
	case 0, 1, 2, 3, 4, 5, 6:
		port := i.dma.From_index(major)
		channel_index := i.dma.Channel(port)

		switch minor {
		case 0:
			channel_index.Set_base(val)
		case 4:
			channel_index.Set_block_control(val)
		case 8:
			channel_index.Set_control(val)
		default:
			panic("Unhandled DMA Write")
		}

		if channel_index.Active() {
			active_port = port
		} else {
			active_port = -1
		}
	case 7:
		switch minor {
		case 0:
			i.dma.Control()
		case 4:
			i.dma.Interrupt()
		default:
			panic("Unhandled DMA Write")
		}

	default:
		panic("Unhandled DMA Write")
	}

	if active_port != -1 {
		i.DoDMA(active_port)
	}
}

func (i *Interconnect) DoDMA(port dma.Port) { //DMA
	switch i.dma.Channel(port).Sync() {
	case dma.LinkedList:
		i.DoDMALinkedList(port)
	default:
		i.DoDMABlock(port)
	}
}

func (i *Interconnect) DoDMALinkedList(port dma.Port) {
	channel := i.dma.Channel(port)
	addr := channel.Base() & 0x1ffffc

	if channel.Direction() == dma.ToRam {
		panic("Unhandled DMA direction")
	}

	if port != dma.Gpu {
		panic("Unhandled DMA port")
	}

	for {
		//Linked lists need headers
		header := i.ram.Load32(addr)
		remsz := header >> 24

		for remsz > 0 {
			addr = (addr + 4) & 0x1ffffc
			src_word := i.ram.Load32(addr)
			i.gpu.Gp0(src_word)
			remsz--
		}

		if header&0x800000 != 0 {
			break
		}

		addr = header & 0x1ffffc
	}

	channel.Done()
}

func (i *Interconnect) DoDMABlock(port dma.Port) {
	var increment int // Assuming dma.Step is an int for simplicity
	var remsz int

	channel := i.dma.Channel(port)

	// Adjust increment based on the channel's step
	switch channel.Step() {
	case dma.Increment:
		increment = 4
	case dma.Decrement:
		increment = -4
	default:
		// Handle any other cases if needed
	}

	addr := channel.Base()

	// Use TransferSize function directly
	transferSize := channel.TransferSize()
	if transferSize != 0 { // Assuming TransferSize returns 0 if size is unknown
		remsz = int(transferSize)
	} else {
		panic("Couldn't figure out DMA block transfer size")
	}

	for remsz > 0 {
		cur_addr := addr & 0x1ffffc
		var src_word uint32

		switch channel.Direction() {
		case dma.ToRam:
			switch port {
			case dma.Otc:
				switch remsz {
				case 1:
					src_word = 0xffffff
				default:
					src_word = Wrapping_sub(addr, 4, 32) & 0x1fffff
				}
				// Hypothetical use of src_word, e.g., writing to memory
				i.Store32(cur_addr, src_word)
			}
		case dma.FromRam:
			src_word = i.ram.Load32(cur_addr)
			switch port {
			case dma.Gpu:
				i.gpu.Gp0(src_word)
			default:
				panic("Unhandled DMA port")
			}
		}

		addr = Wrapping_add(addr, uint32(increment), 32)
		remsz--
	}
	channel.Done()
}

func (i *Interconnect) Load32(addr uint32) uint32 { //load 32-bit at addr
	abaddr := Mask_region(addr)

	if addr%4 != 0 {
		panic("Load32 address alignment error")
	} else if offset := BIOS.Contains(addr); offset != nil {
		return i.bios.Load32(*offset)
	} else if offset := IRQ_CONTROL.Contains(addr); offset != nil {
		fmt.Println("IRQ Control")
		return 0
	} else if offset := DMA.Contains(abaddr); offset != nil {
		return i.Dma_reg(*offset)
	} else if offset := GPU.Contains(abaddr); offset != nil {
		fmt.Println("GPU READ")
		switch *offset {
		case 4:
			return 0x1c000000
		default:
			return 0
		}
	}

	// Instead of panicking, log the unhandled address and return a default value
	log.Printf("Warning: Unhandled Load32 at address: 0x%08x", addr)
	return 0 // Return a default value, or consider throwing an error if that's more appropriate
}

func (i *Interconnect) Load16(addr uint32) uint16 { //load 32-bit at addr
	abaddr := Mask_region(addr)

	if offset := SPU.Contains(abaddr); offset != nil {
		fmt.Printf("Unhandled Write to Register: 0x%08x", abaddr)
		return 0
	}
	if offset := RAM.Contains(abaddr); offset != nil {
		return i.ram.Load16(*offset)
	}
	if offset := IRQ_CONTROL.Contains(addr); offset != nil {
		fmt.Println("IRQ Control")
		return 0
	}
	log.Printf("Unhandled Load16 at Address: 0x%08x", addr)
	return 0
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

	log.Printf("Unhandled Load8 at address: 0x%08x", addr)
	return 0
}

func (i *Interconnect) Store32(addr uint32, val uint32) { //Store value in address
	abaddr := Mask_region(addr)

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
	} else if offset := DMA.Contains(abaddr); offset != nil {
		i.Set_dma_reg(*offset, val)
		return
	} else if offset := GPU.Contains(abaddr); offset != nil {
		switch *offset {
		case 0:
			fmt.Println("GPU Write")
			i.gpu.Gp0(val)
		default:
			fmt.Println("Unhandled GPU Write")
		}
		return
	}

	log.Printf("Unhandled store32 into address: 0x%08x 0x%08x", addr, val)
}

func (i *Interconnect) Store16(addr uint32, val uint16) {
    if addr%2 != 0 {
        log.Printf("Error: Unaligned Store16 address: 0x%08x", addr)
        return // Consider handling this error more gracefully
    }

    abaddr := Mask_region(addr)

    if offset := SPU.Contains(abaddr); offset != nil {
        // Implement or log unhandled write to SPU register
        log.Printf("Warning: Unhandled Write to SPU Register: 0x%08x", val)
    } else if offset := TIMERS.Contains(abaddr); offset != nil {
        // Implement or log unhandled write to Timer register
        log.Printf("Warning: Unhandled Write to Timer Reg: 0x%x: 0x%08x", *offset, val)
    } else if offset := RAM.Contains(abaddr); offset != nil {
        i.ram.Store16(*offset, val)
    } else if offset := IRQ_CONTROL.Contains(addr); offset != nil {
        fmt.Println("IRQ Control Write")
    } else {
        // Log unhandled addresses instead of panicking
        log.Printf("Warning: Unhandled Store16 into address: 0x%08x", addr)
    }
}

func (i *Interconnect) Store8(addr uint32, val uint8) {
	abaddr := Mask_region(addr)

	if offset := RAM.Contains(abaddr); offset != nil {
		i.ram.Store8(*offset, val)
	}

	if offset := EX2.Contains(abaddr); offset != nil {
		panic(fmt.Sprintf("Unhandled Write to EX2 Register: 0x%08x", val))
	}

	log.Printf("Unhandled Store8 into address: 0x%08x", addr)
}

func Mask_region(addr uint32) uint32 {
	index := addr >> 29
	
	return addr & REGION_MASK[index]
}

func Wrapping_add(a uint32, b uint32, mod uint32) uint32 {
	result := int32(a) + int32(b)

	if result < 0 {
		result += int32(mod)
	} else if result >= int32(mod) {
		result %= int32(mod)
	}

	return uint32(result)
}

func Wrapping_sub(a uint32, b uint32, mod uint32) uint32 {
	result := int32(a) - int32(b)

	if result < 0 {
		result += int32(mod)
	} else if result >= int32(mod) {
		result %= int32(mod)
	}

	return uint32(result)
}
