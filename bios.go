package main

import (
	"io/ioutil"
	"path/filepath"
)

type BIOS struct {
	data []uint8
}

func (b *BIOS) New(path string) ([]byte, error){
	//filepath
}