package main

import "fmt"

type BoolDevice struct {
	*Circuit
	Memory []bool
}

func (d *BoolDevice) Reset() {
	memory := d.Memory
	for _, value := range d.Wires {
		memory[value.Index] = value.Nominal
	}
}

func (d *BoolDevice) SetBus(prefix string, values ...bool) {
	memory := d.Memory
	for i, value := range values {
		name := fmt.Sprintf("%s%d", prefix, i)
		s := d.Wires[name]
		memory[s.Index] = value
	}
}

func (d *BoolDevice) Set(name string, value bool) {
	d.Memory[d.Wires[d.Resolve(name)].Index] = value
}

func (d *BoolDevice) Get(name string) bool {
	return d.Memory[d.Wires[d.Resolve(name)].Index]
}

func (d *BoolDevice) SetUint64(prefix string, count int, value uint64) {
	memory := d.Memory
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("%s%d", prefix, i)
		s := d.Wires[d.Resolve(name)]
		if value&1 == 0 {
			memory[s.Index] = false
		} else {
			memory[s.Index] = true
		}
		value >>= 1
	}
}

func (d *BoolDevice) Print(prefix string, count int) {
	memory := d.Memory
	for i := 0; i < count; i++ {
		name, bit := fmt.Sprintf("%s%d", prefix, i), 0
		if alias, ok := d.Aliases[name]; ok {
			s := d.Wires[alias]
			if memory[s.Index] {
				bit = 1
			}
		} else {
			s := d.Wires[name]
			if memory[s.Index] {
				bit = 1
			}
		}
		fmt.Printf("%s=%d\n", name, bit)
	}
}

func (d *BoolDevice) Uint64(prefix string, count int) uint64 {
	var value uint64
	memory := d.Memory
	for i := 0; i < count; i++ {
		name, bit := fmt.Sprintf("%s%d", prefix, i), uint64(0)
		if alias, ok := d.Aliases[name]; ok {
			s := d.Wires[alias]
			if memory[s.Index] {
				bit = 1
			}
		} else {
			s := d.Wires[name]
			if memory[s.Index] {
				bit = 1
			}
		}
		value = value | (bit << uint(i))
	}
	return value
}

func (d *BoolDevice) Execute(reverse bool) {
	memory := d.Memory

	if reverse {
		for i := len(d.Gates) - 1; i >= 0; i-- {
			gate := d.Gates[i]
			switch gate.Type {
			case GateTypeNot:
				a := memory[gate.Taps[0]]
				a = !a
				memory[gate.Taps[0]] = a
			case GateTypeCNot:
				a := memory[gate.Taps[0]]
				b := memory[gate.Taps[1]]
				b = a != b
				memory[gate.Taps[1]] = b
			case GateTypeCCNot:
				a := memory[gate.Taps[0]]
				b := memory[gate.Taps[1]]
				c := memory[gate.Taps[2]]
				c = (a && b) != c
				memory[gate.Taps[2]] = c
			}
		}
		return
	}

	for _, gate := range d.Gates {
		switch gate.Type {
		case GateTypeNot:
			a := memory[gate.Taps[0]]
			a = !a
			memory[gate.Taps[0]] = a
		case GateTypeCNot:
			a := memory[gate.Taps[0]]
			b := memory[gate.Taps[1]]
			b = a != b
			memory[gate.Taps[1]] = b
		case GateTypeCCNot:
			a := memory[gate.Taps[0]]
			b := memory[gate.Taps[1]]
			c := memory[gate.Taps[2]]
			c = (a && b) != c
			memory[gate.Taps[2]] = c
		}
	}
}
