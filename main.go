package main

import (
	"fmt"

	"github.com/Koops0/GPSXE/bios"
	"github.com/Koops0/GPSXE/biosmap"
)

func main() {
	bios, err := bios.New("SCPH1001.bin") //will switch to SCPH7501.bin later
	if err != nil {
		fmt.Println("Error reading file")
		return
	}

	inter := biosmap.Interconnect{}.New(bios)

	cpu := &CPU{}
	cpu.New(inter)
	fmt.Println(cpu.reg[0])
}
