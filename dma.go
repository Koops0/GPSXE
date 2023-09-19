package main

import (

)

type DMA struct {
	control        uint32      //Ctrl reg
}

func(d* DMA) New(){
	d.control =  0x07654321
}

func(d* DMA) Control() uint32{
	return d.control
}