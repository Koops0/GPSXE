package main

type Instruction struct {
	op uint32
}

func (i Instruction) Function() uint32 {
	return i.op >> 26
}

func (i Instruction) S() uint32 {
	return (i.op >> 21) & 0x1f
}

func (i Instruction) T() uint32 {
	return (i.op >> 16) & 0x1f
}

func (i Instruction) Imm() uint32 {
	return i.op & 0xffff
}

func (i Instruction) Imm_se() uint32 {
	v := int16(i.op & 0xffff)
	return uint32(v)
}

func (i Instruction) D() uint32 {
	return (i.op >> 11) & 0x1f
}

func (i Instruction) Subfunction() uint32 {
	return i.op & 0x3f
}

func (i Instruction) Shift() uint32 {
	return (i.op >> 6) & 0x1f
}

func (i Instruction) Imm_jump() uint32 {
	return i.op & 0x3ffffff
}
