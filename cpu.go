package main

import (
	"fmt"

	"./biosmap"
)

type CPU struct {
	pc    uint32      //Program Counter
	next  Instruction //next instruction
	reg   [32]uint32
	inter biosmap.Interconnect //Interface
}

func (c *CPU) New(inter biosmap.Interconnect) {
	c.reg[0] = 0
	c.pc = 0xbfc00000 //reset val
	c.inter = inter
	c.next.op = 0x0
}

func (c CPU) Reg(index uint32) uint32 {
	return c.reg[index]
}

func (c *CPU) Setreg(index uint32, val uint32) {
	c.reg[index] = val
	c.reg[0] = 0
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
	v := i << 16
	c.Setreg(t, v)

}

// Bitwise OR Imm
func (c *CPU) Opori(instruction Instruction) {
	i := instruction.Imm()
	t := instruction.T()
	s := instruction.S()
	v := c.reg[s] | i
	c.Setreg(t, v)
}

func (c *CPU) Opsw(instruction Instruction) {
	i := instruction.Imm()
	t := instruction.T()
	s := instruction.S()
	addr := c.reg[s] + i
	v := c.reg[t]
	c.Store32(addr, v)
}

func (c *CPU) Opsll(instruction Instruction) { //Shift Left
	i := instruction.Shift()
	t := instruction.T()
	d := instruction.D()
	v := c.reg[t] << i
	c.Setreg(d, v)
}

func (c *CPU) Opaddiu(instruction Instruction) { //Add immediate uint
	i := instruction.Imm_se()
	t := instruction.T()
	s := instruction.Function()
	v := c.reg[s] + i
	c.Setreg(t, v)
}

func (c *CPU) Opj(instruction Instruction) { //Jump
	i := instruction.Imm_jump()
	c.pc = (c.pc & 0xf0000000) | (i << 2)

}

func (c *CPU) Opor(instruction Instruction) { //Or
	d := instruction.D()
	s := instruction.S()
	t := instruction.T()
	v := c.reg[s] | c.reg[t]
	c.Setreg(d, v)
}

func (c *CPU) Store32(addr uint32, val uint32) {
	c.inter.Store32(addr, val)
}

func (c *CPU) Decode_and_execute(instruction Instruction) {
	switch instruction.Function() {
	case 0b001111:
		c.Oplui(instruction)
	case 0b001101:
		c.Opori(instruction)
	case 0b001011:
		c.Opsw(instruction)
	case 0b001001:
		c.Opaddiu(instruction)
	case 0b101111:
		c.Opj(instruction)
	case 0b100000:
		c.Opor(instruction)
	case 0b000000:
		switch instruction.Subfunction() {
		case 0b000000:
			c.Opsll(instruction)
		default:
			panic(fmt.Sprintf("Unhandled instruction: %x", instruction))
		}
	default:
		panic(fmt.Sprintf("Unhandled instruction: %x", instruction))
	}
}

func (c *CPU) Run_next(inst Instruction) {
	inst.op = c.next.op              //Grab inst
	c.next.op = c.Load32(c.pc)       //Grab inst
	c.pc = Wrapping_add(c.pc, 4, 32) //Increment PC
	c.Decode_and_execute(inst)
}

func (c *CPU) Load32(addr uint32) uint32 { //load 32-bit from inter
	return c.inter.Load32(addr)
}
