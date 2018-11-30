package main

import "fmt"

type DeviceDual struct {
	*Circuit
	Memory []Dual
}

func (d *DeviceDual) Reset() {
	memory := d.Memory
	for _, value := range d.Wires {
		if value.Nominal {
			memory[value.Index] = Dual{Val: 1.0}
		} else {
			memory[value.Index] = Dual{Val: 0}
		}
	}
}

func (d *DeviceDual) SetBus(prefix string, values ...Dual) {
	memory := d.Memory
	for i, value := range values {
		name := fmt.Sprintf("%s%d", prefix, i)
		s := d.Wires[name]
		memory[s.Index] = value
	}
}

func (d *DeviceDual) Set(name string, value Dual) {
	d.Memory[d.Wires[d.Resolve(name)].Index] = value
}

func (d *DeviceDual) Get(name string) Dual {
	return d.Memory[d.Wires[d.Resolve(name)].Index]
}

func (d *DeviceDual) SetUint64(prefix string, count int, value uint64) {
	memory := d.Memory
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("%s%d", prefix, i)
		s := d.Wires[d.Resolve(name)]
		if value&1 == 0 {
			memory[s.Index] = Dual{Val: 0}
		} else {
			memory[s.Index] = Dual{Val: 1.0}
		}
		value >>= 1
	}
}

func (d *DeviceDual) Print(prefix string, count int) {
	memory := d.Memory
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("%s%d", prefix, i)
		fmt.Printf("%s=%v\n", name, memory[d.Wires[d.Resolve(name)].Index])
	}
}

func (d *DeviceDual) Uint64(prefix string, count int) uint64 {
	var value uint64
	memory := d.Memory
	for i := 0; i < count; i++ {
		name, bit := fmt.Sprintf("%s%d", prefix, i), uint64(0)
		if memory[d.Wires[d.Resolve(name)].Index].Val > 0.5 {
			bit = 1
		}
		value = value | (bit << uint(i))
	}
	return value
}

func (d *DeviceDual) Execute(reverse bool) {
	memory := d.Memory

	if reverse {
		for i := len(d.Gates) - 1; i >= 0; i-- {
			gate := d.Gates[i]
			switch gate.Type {
			case GateTypeNot:
				a := memory[gate.Taps[0]]
				a = Sub(One, a)
				memory[gate.Taps[0]] = a
			case GateTypeCNot:
				a := memory[gate.Taps[0]]
				b := memory[gate.Taps[1]]
				b = Add(Mul(Sub(One, a), b), Mul(Sub(One, b), a))
				memory[gate.Taps[1]] = b
			case GateTypeCCNot:
				a := memory[gate.Taps[0]]
				b := memory[gate.Taps[1]]
				c := memory[gate.Taps[2]]
				c = Add(Mul(Sub(One, Mul(a, b)), c), Mul(Mul(Sub(One, c), a), b))
				memory[gate.Taps[2]] = c
			}
		}
		return
	}

	for _, gate := range d.Gates {
		switch gate.Type {
		case GateTypeNot:
			a := memory[gate.Taps[0]]
			a = Sub(One, a)
			memory[gate.Taps[0]] = a
		case GateTypeCNot:
			a := memory[gate.Taps[0]]
			b := memory[gate.Taps[1]]
			b = Add(Mul(Sub(One, a), b), Mul(Sub(One, b), a))
			memory[gate.Taps[1]] = b
		case GateTypeCCNot:
			a := memory[gate.Taps[0]]
			b := memory[gate.Taps[1]]
			c := memory[gate.Taps[2]]
			c = Add(Mul(Sub(One, Mul(a, b)), c), Mul(Mul(Sub(One, c), a), b))
			memory[gate.Taps[2]] = c
		}
	}
}
