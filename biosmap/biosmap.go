package biosmap

type Range struct {
	address uint32
	bit     uint32
}

var BIOS = Range{
	address: 0xbfc00000,
	bit:	 512*1024,
}

func (r Range) contains(addr uint32) *uint32 {
	if addr >= r.address && addr < r.address+r.bit {
		option := addr-r.address
		return &option
	} else {
		return nil
	}
}
