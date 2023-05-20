package main

import (
	"sync/atomic"
)

type CPU struct {
	pc          uint32
	instruction uint32
}

func (c *CPU) init() {
	c.pc = 0xbfc00000
}

func wrapping_add(a uint32, b uint32, mod uint32) uint32 {
	result := a + b

	if result < 0 {
		result += mod
	} else if result >= mod {
		result %= mod
	}

	return result

}

func (c *CPU) run_next() {
	c.instruction = atomic.LoadUint32(&c.pc)
	c.pc = wrapping_add(c.pc, 4, 32)
	//c.decode_and_execute(c.instruction)
}
