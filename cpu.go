package main

import (
	"fmt"

	"./biosmap"
)

type RegIn uint32
type Load struct {
	r   RegIn
	val uint32
}

func (l *Load) Load(reg RegIn, value uint32) {
	l.r = reg
	l.val = value
}

type CPU struct {
	pc      uint32      //Program Counter
	next    Instruction //next inst
	reg     [32]uint32
	out_reg [32]uint32           //2nd set
	inter   biosmap.Interconnect //Interface
	sr      uint32               //Stat register
	load    Load
}

func (c *CPU) New(inter biosmap.Interconnect) {
	c.reg[0] = 0
	c.pc = 0xbfc00000 //reset val
	c.inter = inter
	c.next.op = 0x0
	c.sr = 0
	c.out_reg = c.reg
	c.load.Load(0, 0)
}

func (c CPU) Reg(index uint32) uint32 {
	return c.reg[index]
}

func (c *CPU) Setreg(index uint32, val uint32) {
	c.reg[index] = val
	c.reg[0] = 0
}

func (c *CPU) Set_reg(index RegIn, val uint32) {
	c.out_reg[index] = val
	c.out_reg[0] = 0
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

func Wrapping_sub(a uint32, b uint32, mod uint32) uint32 {
	result := a - b

	if result < 0 {
		result += mod
	} else if result >= mod {
		result %= mod
	}

	return result
}

func Checkedadd(a, b int32) (int32, error) {
	result := a + b
	if (b > 0 && result < a) || (b < 0 && result > a) {
		return 0, fmt.Errorf("integer overflow")
	}
	return result, nil
}

func (c *CPU) Branch(offset uint32) {
	off := offset << 2
	pc := c.pc
	pc = Wrapping_add(off, pc, 32)
	pc = Wrapping_sub(pc, 4, 32)
	c.pc = pc

}

// Load Upper Immediate
func (c *CPU) Oplui(inst Instruction) {
	i := inst.Imm()
	t := inst.T()
	v := i << 16
	c.Setreg(t, v)

}

// Bitwise OR Imm
func (c *CPU) Opori(inst Instruction) {
	i := inst.Imm()
	t := inst.T()
	s := inst.S()
	v := c.reg[s] | i
	c.Setreg(t, v)
}

func (c *CPU) Opsw(inst Instruction) { //stores word

	if c.sr != 0 && 0x10000 != 0 {
		fmt.Println("ignoring store while cache is isolated")
		return
	}

	i := inst.Imm()
	t := inst.T()
	s := inst.S()
	addr := c.reg[s] + i
	v := c.reg[t]
	c.Store32(addr, v)
}

func (c *CPU) Opsll(inst Instruction) { //Shift Left
	i := inst.Shift()
	t := inst.T()
	d := inst.D()
	v := c.reg[t] << i
	c.Setreg(d, v)
}

func (c *CPU) Opaddiu(inst Instruction) { //Add immediate uint
	i := inst.Imm_se()
	t := inst.T()
	s := inst.S()
	v := c.reg[s] + i
	c.Setreg(t, v)
}

func (c *CPU) Opj(inst Instruction) { //Jump
	i := inst.Imm_jump()
	c.pc = (c.pc & 0xf0000000) | (i << 2)

}

func (c *CPU) Opor(inst Instruction) { //Or
	d := inst.D()
	s := inst.S()
	t := inst.T()
	v := s | c.reg[t]
	c.Setreg(d, v)
}

func (c *CPU) Opmtc0(inst Instruction) { //Or
	cpu_r := inst.T()
	cop_r := inst.D()
	v := c.reg[cpu_r]

	switch cop_r {
	case 3 | 5 | 6 | 7 | 9 | 11:
		if v != 0 {
			panic(fmt.Sprintf("Unhandled write to COP: %x", cop_r))
		}
	case 12:
		c.sr = v
	case 13:
		if v != 0 {
			panic(fmt.Sprintf("Unhandled write to CAUSE"))
		}
	default:
		panic(fmt.Sprintf("Unhandled COP reg: %x", cop_r))
	}
}

func (c *CPU) Opcop0(inst Instruction) {
	switch inst.S() {
	case 0b01000:
		c.Opmtc0(inst)
	default:
		panic(fmt.Sprintf("Unhandled COP inst: %x", inst))
	}
}

func (c *CPU) Opbne(inst Instruction) { //branch not equal
	i := inst.Imm_se()
	s := inst.S()
	t := inst.T()

	if c.reg[s] != c.reg[t] {
		c.Branch(i)
	}
}

func (c *CPU) Opaddi(inst Instruction) { //Add unsigned imm, check for overflow
	i := int32(inst.Imm_se())
	s := inst.S()
	t := inst.T()

	v := int32(c.reg[s])

	res, err := Checkedadd(v, i)

	if err != nil {
		panic("ADDI OVERFLOW")
	} else {
		s = uint32(res)
	}

	c.Setreg(t, s)
}

func (c *CPU) Oplw(inst Instruction) { //Load word
	if c.sr & 0x10000 != 0 {
		fmt.Println("ignoring store while cache is isolated")
		return
	}

	i := inst.Imm_se()
	s := inst.S()
	t := inst.T()
	t_reg := RegIn(t)

	addr := Wrapping_add(c.reg[s], i, 32)

	v := c.Load32(addr)

	c.load.Load(t_reg, v)
}

func (c *CPU) Opsltu(inst Instruction) { //Set on less than unsigned
	d := inst.D()
	t := inst.T()
	s := inst.S()
	v := c.reg[s] < c.reg[t]
	u := 0

	if v == true {
		u = 1
	}
	c.Setreg(d, uint32(u))
}

func (c *CPU) Opaddu(inst Instruction) { //Add unsigned
	d := inst.D()
	t := inst.T()
	s := inst.S()
	v := Wrapping_add(c.reg[s], c.reg[t], 32)
	c.Setreg(d, v)
}

func (c *CPU) Opsh(inst Instruction) { //Store Halfword
	if c.sr & 0x10000 != 0 {
		fmt.Println("ignoring store while cache is isolated")
		return
	}

	i := inst.Imm_se()
	s := inst.S()
	t := inst.T()

	addr := Wrapping_add(c.reg[s], i, 32)

	v := c.reg[t]

	c.Store16(addr, uint16(v))
}

func (c *CPU) Opjal(inst Instruction) { //Jump and Link
	ra := c.pc
	c.Setreg(uint32(c.load.r), ra)
	c.Opj(inst)
}

func (c *CPU) Opandi(inst Instruction) { //Jump and Link
	i := inst.Imm()
	t := inst.T()
	s := inst.S()
	v := c.reg[s] & i
	c.Setreg(t, v)
}

func (c *CPU) Opsb(inst Instruction) { //Store byte
	if c.sr & 0x10000 != 0 {
		fmt.Println("ignoring store while cache is isolated")
		return
	}

	i := inst.Imm_se()
	s := inst.S()
	t := inst.T()

	addr := Wrapping_add(c.reg[s], i, 32)
	v := c.reg[t]

	c.Store8(addr, uint8(v))
}

func (c *CPU) Opjr(inst Instruction) { //Jump
	s := inst.S()
	c.pc = c.reg[s]
}

func (c *CPU) Oplb(inst Instruction) { //Load byte
	i := inst.Imm_se()
	s := inst.S()
	t := inst.T()

	addr := Wrapping_add(c.reg[s], i, 32)
	v := int8(c.Load8(addr)) //cast as int8

	c.load.Load(RegIn(t), uint32(v))
}

func (c *CPU) Opbeq(inst Instruction) { //Branch if equal
	i := inst.Imm_se()
	s := inst.S()
	t := inst.T()

	if c.reg[s] == c.reg[t]{
		c.Branch(i)
	}
}

func (c *CPU) Opmfc0(inst Instruction) { //Coprocessor 0
	cpu_r := inst.T()
	cop_r := inst.D()

	switch cop_r{
	case 12:
		c.sr = cop_r
	case 13:
		panic(fmt.Sprintf("Unhandled read from CAUSE reg."))
	default:
		panic(fmt.Sprintf("Unhandled read from cop_r"))
	}

	v := cop_r

	c.load.Load(RegIn(cpu_r), v)
}

func (c *CPU) Opand(inst Instruction) { //Bit and
	d := inst.D()
	s := inst.S()
	t := inst.T()
	
	v := c.reg[s]+c.reg[t]
	c.Setreg(d,v)
}

func (c *CPU) Opadd(inst Instruction) { //Bit add
	d := inst.D()
	s := inst.S()
	t := inst.T()
	
	ss := int32(c.reg[s])
	tt := int32(c.reg[t])
	res,err := Checkedadd(ss,tt)

	if err != nil{
		panic(fmt.Sprintf("ADD OVERFLOW"))
	}

	c.Setreg(d, uint32(res))
}

func (c *CPU) Store32(addr uint32, val uint32) {
	c.inter.Store32(addr, val)
}

func (c *CPU) Store16(addr uint32, val uint16) {
	c.inter.Store16(addr, val)
}

func (c *CPU) Store8(addr uint32, val uint8) {
	c.inter.Store8(addr, val)
}

func (c *CPU) Load8(addr uint32) uint8{
	return c.inter.Load8(addr)
}

func (c *CPU) Decode_and_execute(inst Instruction) {
	switch inst.Function() {
	case 0b001111:
		c.Oplui(inst)
	case 0b001101:
		c.Opori(inst)
	case 0b001011:
		c.Opsw(inst)
	case 0b001001:
		c.Opaddiu(inst)
	case 0b101111:
		c.Opj(inst)
	case 0b100000: //Subfunction
		switch inst.Subfunction() {
		case 0b100101:
			c.Opor(inst)
		case 0b011110:
			c.Oplb(inst)
		case 0b000000:
			c.Opmfc0(inst)
		case 0b100100:
			c.Opand(inst)
		default:
			panic(fmt.Sprintf("Unhandled instruction: %x", inst))
		}
	case 0b010000:
		c.Opcop0(inst)
	case 0b101010:
		c.Opbne(inst)
	case 0b100001: //subfunction
		switch inst.Subfunction() {
		case 0b010010:
			c.Opaddi(inst)
		case 0b100001:
			c.Opsltu(inst)
		case 0b001010:
			c.Opadd(inst)
		default:
			panic(fmt.Sprintf("Unhandled instruction: %x", inst))
		}
	case 0b100011: //sub
		c.Oplw(inst)
		switch inst.Subfunction() {
		case 0b010000:
			c.Oplw(inst)
		case 0b011110:
			c.Oplb(inst)
		case 0b110000:
			c.Opbeq(inst)
		default:
			panic(fmt.Sprintf("Unhandled instruction: %x", inst))
		}
	case 0b111010:
		c.Opaddu(inst)
	case 0b101001:
		c.Opsh(inst)
	case 0b111111:
		c.Opjal(inst)
	case 0b110000:
		c.Opandi(inst)
	case 0b101000:
		c.Opsb(inst)
	case 0b111110:
		c.Opjr(inst)
	case 0b000000:
		switch inst.Subfunction() {
		case 0b000000:
			c.Opsll(inst)
		default:
			panic(fmt.Sprintf("Unhandled instruction: %x", inst))
		}
	default:
		panic(fmt.Sprintf("Unhandled instruction: %x", inst))
	}
}

func (c *CPU) Run_next(inst Instruction) {
	inst.op = c.next.op              //Grab inst
	c.next.op = c.Load32(c.pc)       //Grab inst
	c.pc = Wrapping_add(c.pc, 4, 32) //Increment PC
	c.Decode_and_execute(inst)

	c.Set_reg(c.load.r, c.load.val)
	c.load.Load(0, 0)
	c.Decode_and_execute(inst)
	c.reg = c.out_reg
}

func (c *CPU) Load32(addr uint32) uint32 { //load 32-bit from inter
	return c.inter.Load32(addr)
}
