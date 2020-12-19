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

type App struct {
	Processes     []ps.Process
	SelectedIndex int
}

func (a *App) Loop() {
	p := processesNames(a.Processes)
	g.SingleWindow("Quick Kill!", g.Layout{
		g.Line(
			g.ListBoxV("pids", 150, 300-15, true, p, nil, func(selectedIndex int) {
				a.SelectedIndex = selectedIndex
			}, nil, nil),
			g.Group(g.Layout{
				g.LabelWrapped(fmt.Sprintf("%d Procesesses found", len(p))),
				g.LabelWrapped(fmt.Sprintf("Selected PID: %d - %s", a.SelectedIndex, p[a.SelectedIndex])),
				g.Button("kill", func() {
					pid := a.Processes[a.SelectedIndex].Pid()
					handle, err := w32.OpenProcess(w32.SYNCHRONIZE|w32.PROCESS_TERMINATE, true, uint32(pid))
					if err != nil {
						log.Println(err)
						return
					}
					if !w32.TerminateProcess(handle, 0) {
						log.Println("Failed terminating.")
						return
					}
				}),
			}),
		),
	})
}

func main() {
	processes, err := ps.Processes()
	if err != nil {
		panic(err)
	}

	a := App{Processes: processes}
	sort.SliceStable(a.Processes, func(i, j int) bool {
		return a.Processes[i].Executable() < a.Processes[j].Executable()
	})

	wnd := g.NewMasterWindow("Quick Kill", 300, 300, g.MasterWindowFlagsNotResizable, nil)
	wnd.Main(a.Loop)
}
