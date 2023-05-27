package main

import (
	"fmt"
	"sync/atomic"

	"./biosmap"
)

type CPU struct {
	pc    uint32 //Program Counter
	reg   [32]uint32
	inter biosmap.Interconnect //Interface
}

func (c CPU) New(inter biosmap.Interconnect) {
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

// Load Upper Immediate
func (c *CPU) Oplui(instruction Instruction) {
	i := instruction.Imm()
	t := instruction.T()
	panic("what now?")
}

func (c *CPU) Decode_and_execute(instruction Instruction) {
	switch instruction.Function() {
	case 0b001111:
		c.Oplui(instruction)
	default:
		panic(fmt.Sprintf("Unhandled instruction: %x", instruction))
	}
}

func (c *CPU) Run_next(inst Instruction) {
	inst.op = atomic.LoadUint32(&c.pc)
	c.pc = Wrapping_add(c.pc, 4, 32)
	c.Decode_and_execute(inst)
}

func (c *CPU) Load32(addr uint32) uint32 {
	return c.inter.Load32(addr)
}
