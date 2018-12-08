// Copyright 2018 The Janus Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "fmt"

type DeviceBool struct {
	*Circuit
	Memory []bool
}

func (d *DeviceBool) Reset() {
	memory := d.Memory
	for _, value := range d.Wires {
		memory[value.Index] = value.Nominal
	}
}

func (d *DeviceBool) SetBus(prefix string, values ...bool) {
	memory := d.Memory
	for i, value := range values {
		name := fmt.Sprintf("%s%d", prefix, i)
		s := d.Wires[name]
		memory[s.Index] = value
	}
}

func (d *DeviceBool) Set(name string, value bool) {
	d.Memory[d.Wires[d.Resolve(name)].Index] = value
}

func (d *DeviceBool) Get(name string) bool {
	return d.Memory[d.Wires[d.Resolve(name)].Index]
}

func (d *DeviceBool) SetUint64(prefix string, value uint64) {
	width, ok := d.Buses[prefix]
	if !ok {
		panic(fmt.Errorf("bus %s not found", prefix))
	}
	if width > 64 {
		panic(fmt.Errorf("bus %s is larger than uint64", prefix))
	}
	memory := d.Memory
	for i := 0; i < int(width); i++ {
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

func (d *DeviceBool) Print(prefix string, count int) {
	memory := d.Memory
	for i := 0; i < count; i++ {
		name, bit := fmt.Sprintf("%s%d", prefix, i), 0
		if memory[d.Wires[d.Resolve(name)].Index] {
			bit = 1
		}
		fmt.Printf("%s=%d\n", name, bit)
	}
}

func (d *DeviceBool) Uint64(prefix string) uint64 {
	width, ok := d.Buses[prefix]
	if !ok {
		panic(fmt.Errorf("bus %s not found", prefix))
	}
	if width > 64 {
		panic(fmt.Errorf("bus %s is larger than uint64", prefix))
	}
	var value uint64
	memory := d.Memory
	for i := 0; i < int(width); i++ {
		name, bit := fmt.Sprintf("%s%d", prefix, i), uint64(0)
		if memory[d.Wires[d.Resolve(name)].Index] {
			bit = 1
		}
		value = value | (bit << uint(i))
	}
	return value
}

func (d *DeviceBool) Execute(reverse bool) {
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
