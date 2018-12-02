// Copyright 2018 The Janus Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"sort"

	"github.com/alixaxel/pagerank"
)

type GateType int

const (
	GateTypeNot GateType = iota
	GateTypeCNot
	GateTypeCCNot
)

type Wire struct {
	Name    string
	Nominal bool
	Index   uint32
	Rank    float64
}

type Gate struct {
	Type GateType
	Taps [3]uint32
}

type Circuit struct {
	Buses   map[string]uint32
	Wires   map[string]Wire
	Aliases map[string]string
	Gates   []Gate
}

func NewCircuit() Circuit {
	return Circuit{
		Buses:   make(map[string]uint32),
		Wires:   make(map[string]Wire),
		Aliases: make(map[string]string),
	}
}

func (c *Circuit) AddBus(prefix string, count int, alias bool, nominal ...bool) {
	_, ok := c.Buses[prefix]
	if ok {
		panic(fmt.Errorf("bus %s already exists", prefix))
	}
	c.Buses[prefix] = uint32(count)
	if alias {
		return
	}
	nom := false
	if len(nominal) > 0 {
		nom = nominal[0]
	}
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("%s%d", prefix, i)
		c.Wires[name] = Wire{
			Name:    name,
			Nominal: nom,
			Index:   uint32(len(c.Wires)),
		}
	}
}

func (c *Circuit) AddWire(name string, nominal bool) {
	_, ok := c.Wires[name]
	if ok {
		panic(fmt.Errorf("wire %s already exists", name))
	}
	c.Wires[name] = Wire{
		Name:    name,
		Nominal: nominal,
		Index:   uint32(len(c.Wires)),
	}
}

func (c *Circuit) AddAlias(name, alias string) []string {
	f := func(name, alias string) string {
		if i, ok := c.Buses[alias]; ok {
			c.Buses[alias] = i + 1
			alias = fmt.Sprintf("%s%d", alias, i)
		}
		_, ok := c.Aliases[alias]
		if ok {
			panic(fmt.Errorf("alias %s already exists", alias))
		}
		c.Aliases[alias] = name
		return alias
	}
	if i, ok := c.Buses[name]; ok {
		aliases := make([]string, i)
		for j := 0; j < int(i); j++ {
			aliases[j] = f(fmt.Sprintf("%s%d", name, j), alias)
		}
		return aliases
	}
	return []string{f(name, alias)}
}

func (c *Circuit) Resolve(name string) string {
	alias, ok := c.Aliases[name]
	if ok {
		return alias
	}
	return name
}

func (c *Circuit) AddGateNot(a string) {
	gate := Gate{
		Type: GateTypeNot,
		Taps: [3]uint32{c.Wires[a].Index},
	}
	c.Gates = append(c.Gates, gate)
}

func (c *Circuit) AddGateCNot(a, b string) {
	gate := Gate{
		Type: GateTypeCNot,
		Taps: [3]uint32{c.Wires[a].Index, c.Wires[b].Index},
	}
	c.Gates = append(c.Gates, gate)
}

func (cc *Circuit) AddGateCCNot(a, b, c string) {
	gate := Gate{
		Type: GateTypeCCNot,
		Taps: [3]uint32{cc.Wires[a].Index, cc.Wires[b].Index, cc.Wires[c].Index},
	}
	cc.Gates = append(cc.Gates, gate)
}

func (c *Circuit) ComputeRanks() {
	ranks := make([]float64, len(c.Wires))
	graph := pagerank.NewGraph()
	for _, gate := range c.Gates {
		switch gate.Type {
		case GateTypeNot:
		case GateTypeCNot:
			graph.Link(gate.Taps[0], gate.Taps[1], 1.0)
			//graph.Link(gate.Taps[1], gate.Taps[0], 1.0)
		case GateTypeCCNot:
			graph.Link(gate.Taps[0], gate.Taps[2], 0.5)
			graph.Link(gate.Taps[1], gate.Taps[2], 0.5)
			//graph.Link(gate.Taps[2], gate.Taps[0], 0.5)
			//graph.Link(gate.Taps[2], gate.Taps[1], 0.5)
		}
	}
	graph.Rank(0.9, 0.000001, func(node uint32, rank float64) {
		ranks[node] = rank
	})
	for name, wire := range c.Wires {
		wire.Rank = ranks[wire.Index]
		c.Wires[name] = wire
	}
}

func (c *Circuit) PrintRanked() {
	wires := make([]Wire, len(c.Wires))
	for _, wire := range c.Wires {
		wires[wire.Index] = wire
	}
	sort.Slice(wires, func(i, j int) bool {
		return wires[i].Rank > wires[j].Rank
	})
	for _, wire := range wires {
		fmt.Printf("%s %f\n", wire.Name, wire.Rank)
	}
}

func (c *Circuit) PrintConnections(name string) {
	index := c.Wires[c.Resolve(name)].Index
	for _, gate := range c.Gates {
		switch gate.Type {
		case GateTypeNot:
			a := gate.Taps[0]
			if a == index {
				fmt.Println(gate)
			}
		case GateTypeCNot:
			b := gate.Taps[1]
			if b == index {
				fmt.Println(gate)
			}
		case GateTypeCCNot:
			c := gate.Taps[2]
			if c == index {
				fmt.Println(gate)
			}
		}
	}
}

func (c *Circuit) NewDeviceBool() DeviceBool {
	memory := make([]bool, len(c.Wires))
	for _, value := range c.Wires {
		memory[value.Index] = value.Nominal
	}
	return DeviceBool{
		Circuit: c,
		Memory:  memory,
	}
}

func (c *Circuit) NewDeviceFloat32() DeviceFloat32 {
	memory := make([]float32, len(c.Wires))
	for _, value := range c.Wires {
		if value.Nominal {
			memory[value.Index] = 1.0
		}
	}
	return DeviceFloat32{
		Circuit: c,
		Memory:  memory,
	}
}

func (c *Circuit) NewDeviceDual() DeviceDual {
	memory := make([]Dual, len(c.Wires))
	for _, value := range c.Wires {
		if value.Nominal {
			memory[value.Index] = Dual{Val: 1.0}
		}
	}
	return DeviceDual{
		Circuit: c,
		Memory:  memory,
	}
}
