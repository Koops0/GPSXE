package main

import (
    "fmt"
)

type Debugger struct {
	breakpoints         []uint32
    readWatchpoints     []uint32
    writeWatchpoints    []uint32
}

func (d *Debugger) AddBreakpoint(addr uint32) {
	for _, breakpoint := range d.breakpoints {
        if breakpoint == addr {
            return
        }
    }
    d.breakpoints = append(d.breakpoints, addr)
}

func (d *Debugger) AddReadBreakpoint(addr uint32) {
	for _, watchpoint := range d.readWatchpoints {
        if watchpoint == addr {
            return
        }
    }
    d.readWatchpoints = append(d.readWatchpoints, addr)
}

func (d *Debugger) AddWriteBreakpoint(addr uint32) {
	for _, watchpoint := range d.writeWatchpoints {
        if watchpoint == addr {
            return
        }
    }
    d.writeWatchpoints = append(d.writeWatchpoints, addr)
}

func (d *Debugger) DelBreakpoint(addr uint32) {
    // Retain only the breakpoints that are not equal to addr
    newBreakpoints := d.breakpoints[:0]
    for _, breakpoint := range d.breakpoints {
        if breakpoint != addr {
            newBreakpoints = append(newBreakpoints, breakpoint)
        }
    }
    d.breakpoints = newBreakpoints
}

func (d *Debugger) DelReadWatchpoint(addr uint32) {
    // Retain only the watchpoints that are not equal to addr
    newWatchpoints := d.readWatchpoints[:0]
    for _, watchpoint := range d.readWatchpoints {
        if watchpoint != addr {
            newWatchpoints = append(newWatchpoints, watchpoint)
        }
    }
    d.readWatchpoints = newWatchpoints
}

func (d *Debugger) DelWriteWatchpoint(addr uint32) {
    // Retain only the watchpoints that are not equal to addr
    newWatchpoints := d.writeWatchpoints[:0]
    for _, watchpoint := range d.writeWatchpoints {
        if watchpoint != addr {
            newWatchpoints = append(newWatchpoints, watchpoint)
        }
    }
    d.writeWatchpoints = newWatchpoints
}

func (d *Debugger) MemoryRead(cpu *CPU, addr uint32) {
    // Handle unaligned watchpoints if necessary
    for _, watchpoint := range d.readWatchpoints {
        if watchpoint == addr {
            fmt.Printf("Read watchpoint triggered at 0x%08X\n", addr)
            d.Debug(*cpu)
        }
    }
}

func (d *Debugger) MemoryWrite(cpu *CPU, addr uint32) {
    // Handle unaligned watchpoints if necessary
    for _, watchpoint := range d.writeWatchpoints {
        if watchpoint == addr {
            fmt.Printf("Write watchpoint triggered at 0x%08X\n", addr)
            d.Debug(*cpu)
        }
    }
}

func (d *Debugger) PcChange(c CPU) {
    for _, breakpoint := range d.breakpoints {
        if c.pc == breakpoint {
            d.Debug(c)
        }
    }
}

func (d *Debugger) Debug(c CPU) {
    fmt.Println("CPU State:")
    fmt.Printf("PC: 0x%08X\n", c.pc)
    fmt.Println("Registers:")
    for i, reg := range c.reg {
        fmt.Printf("R%d: 0x%08X\n", i, reg)
    }
}