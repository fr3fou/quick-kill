package main

import (
	"fmt"
	"sort"

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

var idx = 0

type App struct {
	Processes     []ps.Process
	SelectedIndex int
}

func (a *App) Loop() {
	p := processesNames(a.Processes)
	if g.IsKeyReleased(g.KeyL) {
		fmt.Println("clicked!")
		// pid := a.Processes[idx].Pid()
		// handle, err := w32.OpenProcess(w32.SYNCHRONIZE|w32.PROCESS_TERMINATE, true, uint32(pid))
		// if err != nil {
		// 	log.Println(err)
		// }
		// if !w32.TerminateProcess(handle, 0) {
		// 	log.Println("Failed terminating.")
		// }
	}

	g.SingleWindow("hello world", g.Layout{
		g.ListBoxV("pids", 150, 400-50, true, p, nil, func(selectedIndex int) {
			a.SelectedIndex = selectedIndex
		}, nil, nil),
		g.LabelWrapped(fmt.Sprintf("%d Procesesses found", len(p))),
		g.LabelWrapped(fmt.Sprintf("Selected PID: %d - %s", idx, p[idx])),
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

	wnd := g.NewMasterWindow("Quick Kill", 300, 400, g.MasterWindowFlagsNotResizable, nil)
	wnd.Main(a.Loop)
}
