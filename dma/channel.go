package dma

type Channel struct {
	enable      bool
	dir         Direction
	step        Step
	sync        Sync
	trigger     bool
	chop        bool
	chop_dma_sz uint8
	chop_cpu_sz uint8
	dummy       uint8
	base        uint32
	block_size  uint16
	block_count uint16
}

func (c *Channel) New() {
	c.enable = false
	c.dir = ToRam
	c.step = Increment
	c.sync = Manual
	c.trigger = false
	c.chop = false
	c.chop_dma_sz = 0
	c.chop_cpu_sz = 0
	c.dummy = 0
	c.base = 0
	c.block_size = 0
	c.block_count = 0
}

func (c *Channel) Control() uint32 { //Get interrupt val
	r := uint32(0)
	var chop_res uint32
	var enable_res uint32
	var trigger_res uint32

	if c.chop {
		chop_res = 1
	} else {
		chop_res = 0
	}

	if c.enable {
		enable_res = 1
	} else {
		enable_res = 0
	}

	if c.trigger {
		trigger_res = 1
	} else {
		trigger_res = 0
	}

	r |= uint32(c.dir) << 0
	r |= uint32(c.step) << 1
	r |= chop_res << 8
	r |= uint32(c.sync) << 9
	r |= uint32(c.chop_dma_sz) << 16
	r |= uint32(c.chop_cpu_sz) << 20
	r |= enable_res << 24
	r |= trigger_res << 28
	r |= uint32(c.dummy) << 29

	return r
}

func (c *Channel) Set_control(val uint32) { //Set control
	if val&1 != 0 {
		c.dir = FromRam
	} else {
		c.dir = ToRam
	}

	if (val>>1)&1 != 0 {
		c.step = Decrement
	} else {
		c.step = Increment
	}

	c.chop = (val>>8)&1 != 0

	switch (val >> 9) & 3 {
	case 0:
		c.sync = Manual
	case 1:
		c.sync = Request
	case 2:
		c.sync = LinkedList
	default:
		panic("error!!!!")
	}

	c.chop_dma_sz = uint8((val >> 16) & 7)
	c.chop_cpu_sz = uint8((val >> 20) & 7)

	c.enable = (val>>24)&1 != 0
	c.trigger = (val>>28)&1 != 0
	c.dummy = uint8((val >> 29) & 3)
}

func (c *Channel) Base() uint32 { //Get base
	return c.base
}

func (c *Channel) Set_base(val uint32) { //Set base
	c.base = val & 0x1fffff
}

func (c *Channel) Block_control() uint32 { //Retrieve block control
	bs := uint32(c.block_size)
	bc := uint32(c.block_count)
	return (bc << 16) | bs
}

func (c *Channel) Set_block_control(val uint32) { //Set block control
	c.block_size = uint16(val)
	c.block_count = uint16(val >> 16)
}

func (c *Channel) Active() bool { //Check if channel is active
	var trigger bool

	switch c.sync {
	case Manual:
		trigger = c.trigger
	default:
		trigger = true
	}

	return c.enable && trigger
}

func (c *Channel) Direction() Direction { //Get direction
	return c.dir
}	

func (c *Channel) Step() Step { //Get step
	return c.step
}

func (c *Channel) Sync() Sync { //Get sync
	return c.sync
}

func (c *Channel) TransferSize() uint32 { //Get transfer size
	bs := uint32(c.block_size)
	bc := uint32(c.block_count)
	var transferSize *uint32
	switch c.sync {
	case Manual:
	    // For manual mode only the block size is used
	    transferSize = new(uint32)
	    *transferSize = bs
	case Request:
	    // In DMA request mode we must transfer ‘bc‘ blocks
	    size := bc * bs
	    transferSize = &size
	case LinkedList:
	    // In linked list mode the size is not known ahead of time:
	    // we stop when we encounter the "end of list" marker (0xffffffff)
	    transferSize = nil
	}

	return *transferSize
}

func (c *Channel) Done() { //completed
	c.enable = false
	c.trigger = false
}

type Direction int

const (
	ToRam   = 0
	FromRam = 1
)

type Step int

const (
	Increment = 0
	Decrement = 1
)

type Sync int

const (
	Manual     = 0
	Request    = 1
	LinkedList = 2
)
