package biosmap

import (
	"C:\Users\thesw\Documents\CS\basic-emu\bios"
)

type Interconnect struct {
	bios bios.BIOS
}

func (i Interconnect) new (bios bios.BIOS) Interconnect{
	i.bios = bios
	return i
}

func (i Interconnect) load32(addr uint32) uint32{
	if offset := biosmap.BIOS.contains(addr); offset != nil{

	}else{
		return 0
	}
}