package main

import (
	"fmt"
	"./bios"
	"./biosmap"
)

func main() {
	bios, err := bios.BIOS.New("roms/SCPH1001.bin")
	if err != nil {
		fmt.Println("oops")
	}

	inter := biosmap.Interconnect{}.New(bios)

	cpu := CPU.New(inter)
}
