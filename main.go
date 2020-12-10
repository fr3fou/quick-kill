package main

import (
	"fmt"
	"log"
	"sort"

	g "github.com/AllenDang/giu"
	"github.com/JamesHovious/w32"
	"github.com/mitchellh/go-ps"
)

func processesNames(ps []ps.Process) []string {
	s := []string{}
	for _, p := range ps {
		s = append(s, p.Executable())
	}
	return s
}

var idx = 0

func loop(ps []ps.Process) {
	sort.SliceStable(ps, func(i, j int) bool {
		return ps[i].Executable() < ps[j].Executable()
	})
	p := processesNames(ps)
	g.SingleWindow("hello world", g.Layout{
		g.Line(
			g.ListBoxV("pids", 150, 300-20, true, p, nil, func(selectedIndex int) {
				idx = selectedIndex
			}, nil, nil),
			g.Row(
				g.LabelWrapped(fmt.Sprintf("%d Procesesses found", len(p))),
				g.LabelWrapped(fmt.Sprintf("Selected PID: %d - %s", idx, p[idx])),
				g.Button("kill", func() {
					pid := ps[idx].Pid()
					handle, err := w32.OpenProcess(w32.SYNCHRONIZE|w32.PROCESS_TERMINATE, true, uint32(pid))
					if err != nil {
						log.Println(err)
						return
					}
					if !w32.TerminateProcess(handle, 0) {
						log.Println("some error")
						return
					}
				}),
			),
		),
	})
}

func main() {
	processes, err := ps.Processes()
	if err != nil {
		panic(err)
	}

	wnd := g.NewMasterWindow("Quick Kill", 300, 300, g.MasterWindowFlagsNotResizable, nil)
	wnd.Main(func() {
		loop(processes)
	})
}
