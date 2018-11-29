package main

import "fmt"

type DeviceFloat32 struct {
	*Circuit
	Memory []float32
}

func (d *DeviceFloat32) Reset() {
	memory := d.Memory
	for _, value := range d.Wires {
		if value.Nominal {
			memory[value.Index] = 1.0
		} else {
			memory[value.Index] = 0
		}
	}
}

func (d *DeviceFloat32) SetBus(prefix string, values ...float32) {
	memory := d.Memory
	for i, value := range values {
		name := fmt.Sprintf("%s%d", prefix, i)
		s := d.Wires[name]
		memory[s.Index] = value
	}
}

func (d *DeviceFloat32) Set(name string, value float32) {
	d.Memory[d.Wires[d.Resolve(name)].Index] = value
}

func (d *DeviceFloat32) Get(name string) float32 {
	return d.Memory[d.Wires[d.Resolve(name)].Index]
}

func (d *DeviceFloat32) SetUint64(prefix string, count int, value uint64) {
	memory := d.Memory
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("%s%d", prefix, i)
		s := d.Wires[d.Resolve(name)]
		if value&1 == 0 {
			memory[s.Index] = 0
		} else {
			memory[s.Index] = 1.0
		}
		value >>= 1
	}
}

func (d *DeviceFloat32) Print(prefix string, count int) {
	memory := d.Memory
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("%s%d", prefix, i)
		fmt.Printf("%s=%f\n", name, memory[d.Wires[d.Resolve(name)].Index])
	}
}

func (d *DeviceFloat32) Uint64(prefix string, count int) uint64 {
	var value uint64
	memory := d.Memory
	for i := 0; i < count; i++ {
		name, bit := fmt.Sprintf("%s%d", prefix, i), uint64(0)
		if memory[d.Wires[d.Resolve(name)].Index] > 0.5 {
			bit = 1
		}
		value = value | (bit << uint(i))
	}
	return value
}

func (d *DeviceFloat32) Execute(reverse bool) {
	memory := d.Memory

	if reverse {
		for i := len(d.Gates) - 1; i >= 0; i-- {
			gate := d.Gates[i]
			switch gate.Type {
			case GateTypeNot:
				a := memory[gate.Taps[0]]
				a = 1 - a
				memory[gate.Taps[0]] = a
			case GateTypeCNot:
				a := memory[gate.Taps[0]]
				b := memory[gate.Taps[1]]
				b = (1-a)*b + (1-b)*a
				memory[gate.Taps[1]] = b
			case GateTypeCCNot:
				a := memory[gate.Taps[0]]
				b := memory[gate.Taps[1]]
				c := memory[gate.Taps[2]]
				c = (1-a*b)*c + (1-c)*a*b
				memory[gate.Taps[2]] = c
			}
		}
		return
	}

	for _, gate := range d.Gates {
		switch gate.Type {
		case GateTypeNot:
			a := memory[gate.Taps[0]]
			a = 1 - a
			memory[gate.Taps[0]] = a
		case GateTypeCNot:
			a := memory[gate.Taps[0]]
			b := memory[gate.Taps[1]]
			b = (1-a)*b + (1-b)*a
			memory[gate.Taps[1]] = b
		case GateTypeCCNot:
			a := memory[gate.Taps[0]]
			b := memory[gate.Taps[1]]
			c := memory[gate.Taps[2]]
			c = (1-a*b)*c + (1-c)*a*b
			memory[gate.Taps[2]] = c
		}
	}
}
