package main

import (
	"sync/atomic"

	"./biosmap"
)

type CPU struct {
	pc          uint32 //Program Counter
	instruction uint32
	inter       biosmap.Interconnect //Interface
}

func (c *CPU) New(inter biosmap.Interconnect) {
	c.pc = 0xbfc00000
	c.inter = inter
}

func Wrapping_add(a uint32, b uint32, mod uint32) uint32 {
	result := a + b

	if result < 0 {
		result += mod
	} else if result >= mod {
		result %= mod
	}

	return result

}

func (c *CPU) Run_next() {
	c.instruction = atomic.LoadUint32(&c.pc)
	c.pc = Wrapping_add(c.pc, 4, 32)
	//c.decode_and_execute(c.instruction)
}

func (c *CPU) Load32(addr uint32) uint32 {
	return c.inter.Load32(addr)
}
