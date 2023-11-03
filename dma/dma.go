package dma

type DMA struct {
	control    uint32 //Ctrl reg
	irq_en     bool
	chan_irq_en   uint8
	chan_flags uint8
	force_irq  bool
	dummy_irq  uint8
}


func (d *DMA) New() {
	d.control = 0x07654321
}

func (d *DMA) Irq() bool { //Return interrupt
	channel := d.chan_flags & d.chan_irq_en
	return d.force_irq || (d.chan_irq_en != 0 && channel != 0)
}

func (d *DMA) Interrupt() uint32 { //Get interrupt val
	r := uint32(0)
	var force_irq_res uint32
	var irq_en_res uint32
	var irq_res uint32

	if d.force_irq {
		force_irq_res = 1
	} else {
		force_irq_res = 0
	}

	if d.irq_en {
		irq_en_res = 1
	} else {
		irq_en_res = 0
	}

	if d.Irq(){
		irq_res = 1
	} else {
		irq_res = 0
	}

	r |= uint32(d.dummy_irq)
	r |= force_irq_res << 15
	r |= uint32(d.chan_irq_en) << 16
	r |= irq_en_res << 23
	r |= uint32(d.chan_flags) << 24
	r |= irq_res << 31

	return r

}

func (d *DMA) Set_interrupt(val uint32) { //Set interrupt
	//no idea about bits 0-5
	d.dummy_irq = uint8(val & 0x3f)
	d.force_irq = (val >> 15) & 1 != 0
	d.chan_irq_en = uint8((val >> 16) & 0x7f)
	d.irq_en = (val >> 23) & 1 != 0

	//Writing 1 resets
	ack := uint8((val >> 24) & 0x3f)
	
	d.chan_flags &= ^ack
}

func (d *DMA) Control() uint32 {
	return d.control
}

func (d *DMA) Set_control(val uint32) {
	d.control = val
}
