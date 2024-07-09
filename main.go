package main

import (
	"fmt"

	"github.com/Koops0/GPSXE/bios"
	"github.com/Koops0/GPSXE/biosmap"
)

func main() {
	bios, err := bios.New("SCPH1001.bin")
	if err != nil {
		fmt.Println("oops")
	}

	inter := biosmap.Interconnect{}.New(bios)

	cpu := &CPU{}
	cpu.New(inter)
	fmt.Println(cpu.reg[0])
}
