package main

import (
	"sort"
	"strconv"

	g "github.com/AllenDang/giu"
	"github.com/mitchellh/go-ps"
)

func processesNames(ps []ps.Process) []string {
	s := []string{}
	for _, p := range ps {
		s = append(s, p.Executable())
	}
	return s
}

func loop(ps []ps.Process) {
	p := processesNames(ps)
	sort.Strings(p)

	g.SingleWindow("hello world", g.Layout{
		g.Line(
			g.ListBoxV("pids", 250, 500-20, true, p, nil, nil, nil, nil),
			g.Label(strconv.Itoa(len(ps))),
		),
	})
}

func main() {
	processes, err := ps.Processes()
	if err != nil {
		panic(err)
	}

	wnd := g.NewMasterWindow("Quick Kill", 500, 500, g.MasterWindowFlagsNotResizable, nil)
	wnd.Main(func() {
		loop(processes)
	})
}
