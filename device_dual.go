// Copyright 2018 The Janus Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "fmt"

type Mapping interface {
	Not(a Dual) Dual
	CNot(a, b Dual) Dual
	CCNot(a, b, c Dual) Dual
}

type HyperbolicParaboloidMapping struct {
}

func (h *HyperbolicParaboloidMapping) Not(a Dual) Dual {
	return Sub(One, a)
}

func (h *HyperbolicParaboloidMapping) CNot(a, b Dual) Dual {
	return Add(Mul(Sub(One, a), b), Mul(Sub(One, b), a))
}

func (h *HyperbolicParaboloidMapping) CCNot(a, b, c Dual) Dual {
	return Add(Mul(Sub(One, Mul(a, b)), c), Mul(Mul(Sub(One, c), a), b))
}

type NeuralMapping struct {
	CNotNetwork, CCNotNetwork NetState
}

func NewNeuralMapping() *NeuralMapping {
	cNotNetwork := NewNetwork(2, 2, 1)
	data := []TrainingData{
		{
			[]float32{0, 0}, []float32{0},
		},
		{
			[]float32{1, 0}, []float32{1},
		},
		{
			[]float32{0, 1}, []float32{1},
		},
		{
			[]float32{1, 1}, []float32{0},
		},
	}
	cNotNetwork.Train(data, .001, .4, .6)

	ccNotNetwork := NewNetwork(3, 3, 1)
	data = []TrainingData{
		{
			[]float32{0, 0, 0}, []float32{0},
		},
		{
			[]float32{1, 0, 0}, []float32{0},
		},
		{
			[]float32{0, 1, 0}, []float32{0},
		},
		{
			[]float32{1, 1, 0}, []float32{1},
		},
		{
			[]float32{0, 0, 1}, []float32{1},
		},
		{
			[]float32{1, 0, 1}, []float32{1},
		},
		{
			[]float32{0, 1, 1}, []float32{1},
		},
		{
			[]float32{1, 1, 1}, []float32{0},
		},
	}
	ccNotNetwork.Train(data, .001, .4, .6)

	return &NeuralMapping{
		CNotNetwork:  cNotNetwork.NewNetState(),
		CCNotNetwork: ccNotNetwork.NewNetState(),
	}
}

func (n *NeuralMapping) Not(a Dual) Dual {
	return Sub(One, a)
}

func (n *NeuralMapping) CNot(a, b Dual) Dual {
	n.CNotNetwork.State[0][0] = a
	n.CNotNetwork.State[0][1] = b
	n.CNotNetwork.Inference()
	return n.CNotNetwork.State[2][0]
}

func (n *NeuralMapping) CCNot(a, b, c Dual) Dual {
	n.CCNotNetwork.State[0][0] = a
	n.CCNotNetwork.State[0][1] = b
	n.CCNotNetwork.State[0][2] = c
	n.CCNotNetwork.Inference()
	return n.CCNotNetwork.State[2][0]
}

type DeviceDual struct {
	*Circuit
	Memory  []Dual
	Mapping Mapping
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

func (d *DeviceDual) SetUint64(prefix string, value uint64) {
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

func (d *DeviceDual) Uint64(prefix string) uint64 {
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
		if memory[d.Wires[d.Resolve(name)].Index].Val > 0.5 {
			bit = 1
		}
		value = value | (bit << uint(i))
	}
	return value
}

func (d *DeviceDual) String(prefix string) string {
	width, ok := d.Buses[prefix]
	if !ok {
		panic(fmt.Errorf("bus %s not found", prefix))
	}
	value, memory := make([]rune, width), d.Memory
	for i := 0; i < int(width); i++ {
		name := fmt.Sprintf("%s%d", prefix, i)
		if memory[d.Wires[d.Resolve(name)].Index].Val > 0.5 {
			value[i] = '1'
		} else {
			value[i] = '0'
		}
	}
	return string(value)
}

func (d *DeviceDual) AllocateSlice(prefix string) []Dual {
	count := int(d.Buses[prefix])
	return make([]Dual, count)
}

func (d *DeviceDual) GetSlice(prefix string, values []Dual) {
	count, memory := int(d.Buses[prefix]), d.Memory
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("%s%d", prefix, i)
		values[i] = memory[d.Wires[d.Resolve(name)].Index]
	}
}

func (d *DeviceDual) SetSlice(prefix string, values []Dual) {
	count, memory := int(d.Buses[prefix]), d.Memory
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("%s%d", prefix, i)
		memory[d.Wires[d.Resolve(name)].Index] = values[i]
	}
}

func (d *DeviceDual) Execute(reverse bool) {
	memory, mapping := d.Memory, d.Mapping
	not, cnot, ccnot := mapping.Not, mapping.CNot, mapping.CCNot

	if reverse {
		for i := len(d.Gates) - 1; i >= 0; i-- {
			gate := d.Gates[i]
			switch gate.Type {
			case GateTypeNot:
				a := memory[gate.Taps[0]]
				memory[gate.Taps[0]] = not(a)
			case GateTypeCNot:
				a := memory[gate.Taps[0]]
				b := memory[gate.Taps[1]]
				memory[gate.Taps[1]] = cnot(a, b)
			case GateTypeCCNot:
				a := memory[gate.Taps[0]]
				b := memory[gate.Taps[1]]
				c := memory[gate.Taps[2]]
				memory[gate.Taps[2]] = ccnot(a, b, c)
			}
		}
		return
	}

	for _, gate := range d.Gates {
		switch gate.Type {
		case GateTypeNot:
			a := memory[gate.Taps[0]]
			memory[gate.Taps[0]] = not(a)
		case GateTypeCNot:
			a := memory[gate.Taps[0]]
			b := memory[gate.Taps[1]]
			memory[gate.Taps[1]] = cnot(a, b)
		case GateTypeCCNot:
			a := memory[gate.Taps[0]]
			b := memory[gate.Taps[1]]
			c := memory[gate.Taps[2]]
			memory[gate.Taps[2]] = ccnot(a, b, c)
		}
	}
}
