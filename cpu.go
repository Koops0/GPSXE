package main

import (
	"fmt"
	"github.com/Koops0/GPSXE/biosmap"
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
	pc         uint32      //Program Counter
	next_pc    uint32      //Next val
	next       Instruction //next inst
	reg        [32]uint32
	out_reg    [32]uint32           //2nd set
	inter      biosmap.Interconnect //Interface
	sr         uint32               //Stat register
	load       Load
	hi         uint32
	lo         uint32
	current_pc uint32 //Inst address
	cause      uint32 //Cop0 13
	epc        uint32 //Cop0 14
	branch     bool   //if branch occured
	delay_slot bool   //if inst executes
}

type Exception uint32

const (
	SysCall            = 0x8
	Overflow           = 0xc
	LoadAddressError   = 0x4
	StoreAddressError  = 0x5
	Break              = 0x9
	CoprocessorError   = 0xb
	IllegalInstruction = 0xa
)

func (c *CPU) New(inter biosmap.Interconnect) CPU{
	c.reg[0] = 0
	c.pc = 0xbfc00000 //reset val
	c.next_pc = Wrapping_add(c.pc, 4, 32)
	c.inter = inter
	c.next.op = 0x0
	c.sr = 0
	c.out_reg = c.reg
	c.load.Load(0, 0)
	c.hi = 0xdeadbeef
	c.lo = 0xdeadbeef
	c.branch = false
	c.delay_slot = false
	return *c
}

func (c CPU) Reg(index uint32) uint32 {
	return c.reg[index]
}

func (c *CPU) Setreg(index uint32, val uint32) {
    if index == 0 {
        // Log or debug when an attempt is made to set register 0, if unexpected
        fmt.Println("Attempt to set reg[0], operation ignored.")
        return
    }
    c.reg[index] = val
}

func (c *CPU) Set_reg(index RegIn, val uint32) {
    if index == 0 {
        // Similar check for out_reg
        fmt.Println("Attempt to set out_reg[0], operation ignored.")
        return
    }
    c.out_reg[index] = val
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

func Checkedadd(a, b int32) (int32, error) {
	result := a + b
	if (b > 0 && result < a) || (b < 0 && result > a) {
		return 0, fmt.Errorf("integer overflow")
	}
	return result, nil
}

func Checkedsub(a, b int32) (int32, error) {
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

	if addr%4 == 0 {
		c.Store32(addr, v)
	} else {
		c.Exception(StoreAddressError)
	}
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

func (c *CPU) Opmfc0(inst Instruction) { //Coprocessor 0
	cpu_r := inst.T()
	cop_r := inst.D()

	switch cop_r {
	case 12:
		c.sr = cop_r
	case 13:
		c.cause = cop_r
	case 14:
		c.epc = cop_r
	default:
		panic("Unhandled read from cop_r")
	}

	v := cop_r

	c.load.Load(RegIn(cpu_r), v)
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
			panic("Unhandled write to CAUSE")
		}
	default:
		panic(fmt.Sprintf("Unhandled COP reg: %x", cop_r))
	}
}

func (c *CPU) Oprfe(inst Instruction) { //Return from Exception
	if inst.op&0x3f != 0b010000 {
		panic(fmt.Sprintf("Invalid COP inst: %x", inst))
	}

	mode := c.sr & 0x3f //Shift interrupt
	c.sr &^= 0x3f
	c.sr |= mode >> 2

}

func (c *CPU) Opcop0(inst Instruction) {
	switch inst.S() {
	case 0b00000:
		c.Opmfc0(inst)
	case 0b00100:
		c.Opmtc0(inst)
	case 0b10000:
		c.Oprfe(inst)
	default:
		panic(fmt.Sprintf("Unhandled COP inst: %x", inst))
	}
}

func (c *CPU) Opcop1(Instruction) {
	c.Exception(CoprocessorError)
}

func (c *CPU) Opcop2(inst Instruction) {
	panic(fmt.Sprintf("Unhandled inst: %x", inst))
}

func (c *CPU) Opcop3(Instruction) {
	c.Exception(CoprocessorError)
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
		c.Exception(Overflow)
	} else {
		c.Setreg(t, uint32(res))
	}
}

func (c *CPU) Oplw(inst Instruction) { //Load word
	if c.sr&0x10000 != 0 {
		fmt.Println("ignoring store while cache is isolated")
		return
	}

	i := inst.Imm_se()
	s := inst.S()
	t := inst.T()
	t_reg := RegIn(t)

	addr := Wrapping_add(c.reg[s], i, 32)

	if addr%4 == 0 {
		v := c.Load32(addr)
		c.load.Load(t_reg, v)
	} else {
		c.Exception(LoadAddressError)
	}
}

func (c *CPU) Opsltu(inst Instruction) { //Set on less than unsigned
	d := inst.D()
	t := inst.T()
	s := inst.S()

	v := c.reg[s] < c.reg[t]
	u := 0

	if v{
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
	if c.sr&0x10000 != 0 {
		fmt.Println("ignoring store while cache is isolated")
		return
	}

	i := inst.Imm_se()
	s := inst.S()
	t := inst.T()

	addr := Wrapping_add(c.reg[s], i, 32)

	v := c.reg[t]

	if addr%2 == 0 {
		c.Store16(addr, uint16(v))
	} else {
		c.Exception(StoreAddressError)
	}
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
	if c.sr&0x10000 != 0 {
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

	if c.reg[s] == c.reg[t] {
		c.Branch(i)
	}
}

func (c *CPU) Opand(inst Instruction) { //Bit and
	d := inst.D()
	s := inst.S()
	t := inst.T()

	v := c.reg[s] + c.reg[t]
	c.Setreg(d, v)
}

func (c *CPU) Opadd(inst Instruction) { //Bit add
	d := inst.D()
	s := inst.S()
	t := inst.T()

	ss := int32(c.reg[s])
	tt := int32(c.reg[t])
	res, err := Checkedadd(ss, tt)

	if err != nil {
		c.Exception(Overflow)
	} else {
		c.Setreg(d, uint32(res))
	}
}

func (c *CPU) Opbgtz(inst Instruction) { //Branch if > 0
	i := inst.Imm_se()
	s := inst.S()

	v := int32(c.reg[s])
	if v > 0 {
		c.Branch(i)
	}
}

func (c *CPU) Opblez(inst Instruction) { //Branch if </= 0
	i := inst.Imm_se()
	s := inst.S()

	v := int32(c.reg[s])
	if v <= 0 {
		c.Branch(i)
	}
}

func (c *CPU) Oplbu(inst Instruction) { //Load unsigned byte
	i := inst.Imm_se()
	t := inst.T()
	s := inst.S()

	addr := Wrapping_add(c.reg[s], i, 32)
	v := c.Load8(addr)
	c.load.Load(RegIn(t), uint32(v))
}

func (c *CPU) Opjalr(inst Instruction) { //Jump+Link
	d := inst.D()
	s := inst.S()

	ra := c.pc

	c.Setreg(d, ra)
	c.pc = c.reg[s]
}

func (c *CPU) Opbxx(inst Instruction) { //Tons of inst
	i := inst.Imm_se()
	s := inst.S()

	instruction := inst.op
	is_bgez := (instruction >> 16) & 1
	is_link := (instruction>>20)&1 != 0

	v := int32(c.reg[s])

	test := uint32(0)
	if v < 0 {
		test = 1
	}
	test = test ^ is_bgez

	if test != 0 {
		if is_link {
			ra := c.pc
			c.Set_reg(RegIn(31), ra)
		} else {
			c.Branch(i)
		}
	}
}

func (c *CPU) Opslti(inst Instruction) { //set if less than imm
	i := int32(inst.Imm_se())
	s := inst.S()
	t := inst.T()

	v := int32(c.reg[s]) < i

	x := 0
	if v {
		x = 1
	}

	c.Setreg(t, uint32(x))
}

func (c *CPU) Opsubu(inst Instruction) { //sub unsigned
	s := inst.S()
	t := inst.T()
	d := inst.D()

	v := Wrapping_sub(c.reg[s], c.reg[t], 32)

	c.Setreg(d, v)
}

func (c *CPU) Opsra(inst Instruction) { //shift right arithmetic
	i := inst.Shift()
	t := inst.T()
	d := inst.D()

	v := int32(c.reg[t]) >> i

	c.Setreg(d, uint32(v))
}

func (c *CPU) Opdiv(inst Instruction) { //divide
	s := inst.S()
	t := inst.T()
	n := int32(c.reg[s])
	d := int32(c.reg[t])

	if d == 0 { //divide by 0
		c.hi = uint32(n)

		if n >= 0 {
			c.lo = 0xffffffff
		} else {
			c.lo = 1
		}
	} else if uint32(n) == 0x80000000 && d == -1 { //unsigned
		c.hi = 0
		c.lo = 0x80000000
	} else {
		c.hi = uint32(n % d)
		c.lo = uint32(n / d)
	}
}

func (c *CPU) Opmflo(inst Instruction) { //move from lo
	d := inst.D()
	lo := c.lo
	c.Setreg(d, lo)
}

func (c *CPU) Opsrl(inst Instruction) { //shift right logical
	i := inst.Shift()
	t := inst.T()
	d := inst.D()

	v := c.reg[t] >> i

	c.Setreg(d, v)
}

func (c *CPU) Opsltiu(inst Instruction) { //set less than imm unsigned
	i := inst.Imm_se()
	s := inst.S()
	t := inst.T()

	v := c.reg[s] << i

	c.Setreg(t, v)
}

func (c *CPU) Opdivu(inst Instruction) { //divide unsigned
	s := inst.S()
	t := inst.T()
	n := c.reg[s]
	d := c.reg[t]

	if d == 0 { //divide by 0
		c.hi = n
		c.lo = 0xffffffff
	} else {
		c.hi = n % d
		c.lo = n / d
	}
}

func (c *CPU) Opmfhi(inst Instruction) { //move from hi
	d := inst.D()
	hi := c.hi
	c.Setreg(d, hi)
}

func (c *CPU) Opslt(inst Instruction) { //set on less signed
	d := inst.D()
	t := inst.T()
	s := inst.S()

	v := int32(c.reg[s]) < int32(c.reg[t])
	u := 0

	if v{
		u = 1
	}
	c.Setreg(d, uint32(u))
}

func (c *CPU) Opmtlo(inst Instruction) { //Move to LO
	s := inst.S()
	c.lo = c.Reg(s)
}

func (c *CPU) Opmthi(inst Instruction) { //Move to LO
	s := inst.S()
	c.hi = c.Reg(s)
}

func (c *CPU) Oplhu(inst Instruction) { //Load halfword unsigned
	i := inst.Imm_se()
	t := inst.T()
	s := inst.S()

	addr := Wrapping_add(c.reg[s], i, 32)

	if addr%2 == 0 {
		v := c.Load16(addr)
		c.load.Load(RegIn(t), uint32(v))
	} else {
		c.Exception(LoadAddressError)
	}
}

func (c *CPU) Opsllv(inst Instruction) { //shift ll var
	d := inst.D()
	t := inst.T()
	s := inst.S()

	v := c.reg[t] << (c.reg[s] & 0x1f)

	c.Setreg(d, v)
}

func (c *CPU) Oplh(inst Instruction) { //Load halfword
	i := inst.Imm_se()
	t := inst.T()
	s := inst.S()

	addr := Wrapping_add(c.reg[s], i, 32)
	v := int16(c.Load16(addr))
	c.load.Load(RegIn(t), uint32(v))
}

func (c *CPU) Opnor(inst Instruction) { //Not Or
	d := inst.D()
	t := inst.T()
	s := inst.S()

	v := ^(c.reg[s] | c.reg[t])

	c.Setreg(d, v)
}

func (c *CPU) Opsrav(inst Instruction) { //shift ra var
	d := inst.D()
	t := inst.T()
	s := inst.S()

	v := (int32(c.reg[t])) >> (c.reg[s] & 0x1f)

	c.Setreg(d, uint32(v))
}

func (c *CPU) Opsrlv(inst Instruction) { //shift rl var
	d := inst.D()
	t := inst.T()
	s := inst.S()

	v := c.reg[t] >> (c.reg[s] & 0x1f)

	c.Setreg(d, v)
}

func (c *CPU) Opmultu(inst Instruction) { //multiply unsigned
	t := inst.T()
	s := inst.S()

	a := uint64(c.reg[s])
	b := uint64(c.reg[t])

	v := a * b

	c.hi = uint32(v >> 32)
	c.lo = uint32(v)
}

func (c *CPU) Opxor(inst Instruction) { //Exclusive Or
	d := inst.D()
	t := inst.T()
	s := inst.S()

	v := c.reg[s] ^ c.reg[t]

	c.Setreg(d, v)
}

func (c *CPU) Opmult(inst Instruction) { //multiply
	t := inst.T()
	s := inst.S()

	a := int64(c.reg[s])
	b := int64(c.reg[t])

	v := uint64(a * b)

	c.hi = uint32(v >> 32)
	c.lo = uint32(v)
}

func (c *CPU) Opsub(inst Instruction) { //Subtract
	d := inst.D()
	t := int32(c.reg[inst.T()])
	s := int32(c.reg[inst.S()])

	res, err := Checkedsub(s, t)

	if err != nil {
		c.Exception(Overflow)
	} else {
		c.Setreg(d, uint32(res))
	}
}

func (c *CPU) Opxori(inst Instruction) { //Exclusive Or + Imm
	i := inst.Imm()
	t := inst.T()
	s := inst.S()

	v := c.reg[s] ^ i

	c.Setreg(t, v)
}

func (c *CPU) Oplwl(inst Instruction) { //Load word left
	i := inst.Imm_se()
	t := inst.T()
	s := inst.S()

	addr := Wrapping_add(c.reg[s], i, 32)

	cur_v := c.out_reg[t] //Bypass LD

	aligned_addr := addr &^ 3
	aligned_word := c.Load32(aligned_addr)

	v := addr & 3

	switch v {
	case 0:
		v = (cur_v & 0x00ffffff) | (aligned_word << 24)
	case 1:
		v = (cur_v & 0x0000ffff) | (aligned_word << 16)
	case 2:
		v = (cur_v & 0x000000ff) | (aligned_word << 8)
	case 3:
		v = 0 | (aligned_word << 0)
	default:
		panic("Unreachable")
	}

	c.load.Load(RegIn(t), v)
}

func (c *CPU) Oplwr(inst Instruction) { //Load word right
	i := inst.Imm_se()
	t := inst.T()
	s := inst.S()

	addr := Wrapping_add(c.reg[s], i, 32)

	cur_v := c.out_reg[t] //Bypass LD

	aligned_addr := addr &^ 3
	aligned_word := c.Load32(aligned_addr)

	v := addr & 3

	switch v {
	case 0:
		v = (cur_v & 0x00ffffff) | (aligned_word >> 24)
	case 1:
		v = (cur_v & 0x0000ffff) | (aligned_word >> 16)
	case 2:
		v = (cur_v & 0x000000ff) | (aligned_word >> 8)
	case 3:
		v = 0 | (aligned_word >> 0)
	default:
		panic("Unreachable")
	}

	c.load.Load(RegIn(t), v)
}

func (c *CPU) Opswl(inst Instruction) { //Shift word left
	i := inst.Imm_se()
	t := inst.T()
	s := inst.S()

	addr := Wrapping_add(c.reg[s], i, 32)
	v := c.reg[t]

	aligned_addr := addr &^ 3
	cur_mem := c.Load32(aligned_addr)

	var mem uint32 = addr & 3

	switch v {
	case 0:
		mem = (cur_mem & 0x00ffffff) | (v >> 24)
	case 1:
		mem = (cur_mem & 0x0000ffff) | (v >> 16)
	case 2:
		mem = (cur_mem & 0x000000ff) | (v >> 8)
	case 3:
		mem = 0 | (v >> 0)
	default:
		panic("Unreachable")
	}

	c.Store32(addr, mem)
}

func (c *CPU) Opswr(inst Instruction) { //Shift word right
	i := inst.Imm_se()
	t := inst.T()
	s := inst.S()

	addr := Wrapping_add(c.reg[s], i, 32)
	v := c.reg[t]

	aligned_addr := addr &^ 3
	cur_mem := c.Load32(aligned_addr)

	var mem uint32 = addr & 3

	switch v {
	case 0:
		mem = (cur_mem & 0x00ffffff) | (v >> 24)
	case 1:
		mem = (cur_mem & 0x0000ffff) | (v >> 16)
	case 2:
		mem = (cur_mem & 0x000000ff) | (v >> 8)
	case 3:
		mem = 0 | (v >> 0)
	default:
		panic("Unreachable")
	}

	c.Store32(addr, mem)
}

func (c *CPU) Oplwc0(Instruction) {
	c.Exception(CoprocessorError)
}

func (c *CPU) Oplwc1(Instruction) {
	c.Exception(CoprocessorError)
}

func (c *CPU) Oplwc2(inst Instruction) {
	panic(fmt.Sprintf("Unhandled inst: %x", inst))
}

func (c *CPU) Oplwc3(Instruction) {
	c.Exception(CoprocessorError)
}

func (c *CPU) Opswc0(Instruction) {
	c.Exception(CoprocessorError)
}

func (c *CPU) Opswc1(Instruction) {
	c.Exception(CoprocessorError)
}

func (c *CPU) Opswc2(inst Instruction) {
	panic(fmt.Sprintf("Unhandled inst: %x", inst))
}

func (c *CPU) Opswc3(Instruction) {
	c.Exception(CoprocessorError)
}

func (c *CPU) Opbreak(Instruction) {
	c.Exception(Break)
}

func (c *CPU) Opsyscall(Instruction) { //Syscall
	c.Exception(SysCall)
}

func (c *CPU) Opillegal(inst Instruction) { //Syscall
	fmt.Printf("Illegal instruction: %x", inst)
	c.Exception(IllegalInstruction)
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

func (c *CPU) Load16(addr uint32) uint16 {
	return c.inter.Load16(addr)
}

func (c *CPU) Load8(addr uint32) uint8 {
	return c.inter.Load8(addr)
}

func (c *CPU) Decode_and_execute(inst Instruction) {
	switch inst.Function() {
	case 0b000000:
		switch inst.Subfunction() {
		case 0b000000:
			c.Opsll(inst)
		case 0b000010:
			c.Opsrl(inst)
		case 0b000011:
			c.Opsra(inst)
		case 0b000100:
			c.Opsllv(inst)
		case 0b000110:
			c.Opsrlv(inst)
		case 0b000111:
			c.Opsrav(inst)
		case 0b001000:
			c.Opjr(inst)
		case 0b001001:
			c.Opjalr(inst)
		case 0b001100:
			c.Opsyscall(inst)
		case 0b001101:
			c.Opbreak(inst)
		case 0b010000:
			c.Opmfhi(inst)
		case 0b010001:
			c.Opmthi(inst)
		case 0b010010:
			c.Opmflo(inst)
		case 0b010011:
			c.Opmtlo(inst)
		case 0b011000:
			c.Opmult(inst)
		case 0b011001:
			c.Opmultu(inst)
		case 0b011010:
			c.Opdiv(inst)
		case 0b011011:
			c.Opdivu(inst)
		case 0b100000:
			c.Opadd(inst)
		case 0b100001:
			c.Opaddu(inst)
		case 0b100010:
			c.Opsub(inst)
		case 0b100011:
			c.Opsubu(inst)
		case 0b100100:
			c.Opand(inst)
		case 0b100101:
			c.Opor(inst)
		case 0b100110:
			c.Opxor(inst)
		case 0b100111:
			c.Opnor(inst)
		case 0b101010:
			c.Opslt(inst)
		case 0b101011:
			c.Opsltu(inst)
		default:
			c.Opillegal(inst)
		}
	case 0b000001:
		c.Opbxx(inst)
	case 0b000010:
		c.Opj(inst)
	case 0b000011:
		c.Opjal(inst)
	case 0b000100:
		c.Opbeq(inst)
	case 0b000101:
		c.Opbne(inst)
	case 0b000110:
		c.Opblez(inst)
	case 0b000111:
		c.Opbgtz(inst)
	case 0b001000:
		c.Opaddi(inst)
	case 0b001001:
		c.Opaddiu(inst)
	case 0b001010:
		c.Opslti(inst)
	case 0b001011:
		c.Opsltiu(inst)
	case 0b001100:
		c.Opandi(inst)
	case 0b001101:
		c.Opori(inst)
	case 0b001110:
		c.Opxori(inst)
	case 0b001111:
		c.Oplui(inst)
	case 0b010000:
		c.Opcop0(inst)
	case 0b010001:
		c.Opcop1(inst)
	case 0b010010:
		c.Opcop2(inst)
	case 0b010011:
		c.Opcop3(inst)
	case 0b100000:
		c.Oplb(inst)
	case 0b100001:
		c.Oplh(inst)
	case 0b100010:
		c.Oplwl(inst)
	case 0b100011:
		c.Oplw(inst)
	case 0b100100:
		c.Oplbu(inst)
	case 0b100101:
		c.Oplhu(inst)
	case 0b100110:
		c.Oplwr(inst)
	case 0b101000:
		c.Opsb(inst)
	case 0b101001:
		c.Opsh(inst)
	case 0b101010:
		c.Opslt(inst)
	case 0b101011:
		c.Opsltu(inst)
	case 0b101110:
		c.Opswr(inst)
	case 0b110000:
		c.Oplwc0(inst)
	case 0b110001:
		c.Oplwc1(inst)
	case 0b110010:
		c.Oplwc2(inst)
	case 0b110011:
		c.Oplwc3(inst)
	case 0b111000:
		c.Opswc0(inst)
	case 0b111001:
		c.Opswc1(inst)
	case 0b111010:
		c.Opswc2(inst)
	case 0b111011:
		c.Opswc3(inst)
	default:
		c.Opillegal(inst)
	}
}

func (c *CPU) Run_next() {
	c.current_pc = c.pc

	if c.current_pc % 4 != 0 {
		c.Exception(LoadAddressError)
		return
	}

	inst := Instruction{op: c.Load32(c.pc)}

	c.pc = c.next_pc
	c.next_pc = Wrapping_add(c.next_pc, 4, 32)

	c.Set_reg(c.load.r, c.load.val)
	c.load.Load(0,0)

	c.delay_slot = c.branch
	c.branch = false
	c.Decode_and_execute(inst) // Cast inst to Instruction type

	c.reg = c.out_reg
}

func (c *CPU) Load32(addr uint32) uint32 { //load 32-bit from inter
	return c.inter.Load32(addr)
}

func (c *CPU) Exception(cause Exception) { //Trigger Exception
	var handler uint32

	if (c.sr & (1 << 22)) != 0 { //Handler depending on BEV
		handler = 0xbfc00180
	} else {
		handler = 0x80000080
	}

	mode := c.sr & 0x3f //Shift bits 2 places to the left
	c.sr &= ^uint32(0x3f)
	c.sr |= (mode << 2) & 0x3f

	c.cause = uint32(cause) << 2 //Update cause

	c.epc = c.current_pc

	if c.delay_slot {
		c.epc = Wrapping_sub(c.epc, 4, 32)
		c.cause |= 1 << 31
	}
	c.pc = handler
	c.next_pc = Wrapping_add(c.next_pc, 4, 32)
}
