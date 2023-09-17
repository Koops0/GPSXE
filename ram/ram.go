package ram

type RAM struct {
	data []uint8
}

func (r *RAM) New() RAM {
	data := make([]uint8, 2*1024*1024)
	return RAM{data: data}
}

func (r *RAM) Load32(offset uint32) uint32 { //Fetch word at offset
	b0 := uint32(r.data[offset])
	b1 := uint32(r.data[offset+1])
	b2 := uint32(r.data[offset+2])
	b3 := uint32(r.data[offset+3])

	return b0 | (b1 << 8) | (b2 << 16) | (b3 << 24)
}

func (r *RAM) Load16(offset uint32) uint16 { //Fetch word at offset
	b0 := uint16(r.data[offset])
	b1 := uint16(r.data[offset+1])

	return b0 | (b1 << 8)
}

func (r *RAM) Load8(offset uint32) uint8 { //Fetch word at offset
	return r.data[offset]
}

func (r *RAM) Store32(offset uint32, val uint32) { //Store val into offset
	b0 := uint8(val)
	b1 := uint8(val >> 8)
	b2 := uint8(val >> 16)
	b3 := uint8(val >> 24)

	r.data[offset] = b0
	r.data[offset+1] = b1
	r.data[offset+2] = b2
	r.data[offset+3] = b3
}

func (r *RAM) Store16(offset uint32, val uint16) { //Store val into offset
	b0 := uint8(val)
	b1 := uint8(val >> 8)

	r.data[offset] = b0
	r.data[offset+1] = b1
}

func (r *RAM) Store8(offset uint32, val uint8) { //Store val into offset
	r.data[offset] = val
}
