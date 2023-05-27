package main

type Instruction struct{
	op	uint32
}

func (i Instruction) Function() uint32{
	return i.op >> 26
}

func (i Instruction) T() uint32{
	return (i.op >> 16) & 0x1f
}

func (i Instruction) Imm() uint32{
	return i.op & 0xffff
}
