package main

import (
	"log"
	"sort"
	"strings"

	g "github.com/AllenDang/giu"
	"github.com/davecgh/go-spew/spew"
	"github.com/mitchellh/go-ps"
)

// func FindProcess(ps []Process, pid int) Process {
// 	for _, p := range ps {
// 		if p.Pid == pid {
// 			return p
// 		}
// 	}
// 	return Process{}
// }

type Process struct {
	Pid      int
	PPid     int
	Cmd      string
	Children []*Process
}

type App struct {
	processes       []Process
	pidMap          map[int]*Process
	SelectedProcess Process
	FilterWord      string
}

func (a *App) Processes() error {
	processes, err := ps.Processes()
	if err != nil {
		log.Println("cannot get processes: " + err.Error())
		return err
	}

	pids := map[int]*Process{}
	for _, proc := range processes {
		// skip pid 0
		if proc.Pid() == 0 {
			continue
		}

		pids[proc.Pid()] = &Process{
			Pid:  proc.Pid(),
			PPid: proc.PPid(),
			Cmd:  proc.Executable(),
		}
	}

	for _, p := range processes {
		if p.Pid() == p.PPid() {
			continue
		}

		if proc, ok := pids[p.PPid()]; ok {
			proc.Children = append(proc.Children, pids[p.Pid()])
		}
	}

	a.processes = []Process{}
	for _, proc := range pids {
		if strings.Index(proc.Cmd, a.FilterWord) == -1 {
			continue
		}

		a.processes = append(a.processes, *proc)
	}

	sort.Slice(a.processes, func(i, j int) bool {
		return a.processes[i].Cmd < a.processes[j].Cmd
	})
	a.pidMap = pids

	return nil
}

func (a *App) ProcessRows() []g.Widget {
	v := []g.Widget{}
	for _, p := range a.processes {
		if parent, ok := a.pidMap[p.PPid]; ok && parent.Cmd != p.Cmd {
			v = append(v, g.Line(ProcessWidget(p)))
		}
	}
	return v
}

func ProcessWidget(p Process) g.Widget {
	if len(p.Children) == 0 {
		return g.TreeNode(p.Cmd, g.TreeNodeFlagsLeaf, nil)
	}

	children := []g.Widget{}
	for _, c := range p.Children {
		children = append(children, ProcessWidget(*c))
	}
	// children = append(children, g.Button(fmt.Sprintf("Kill-%d", p.Pid), func() {
	// 	proc, err := os.FindProcess(p.Pid)
	// 	if err != nil {
	// 		log.Println(err)
	// 		return
	// 	}

	// 	if err := proc.Kill(); err != nil {
	// 		log.Println("Failed killing process")
	// 	}
	// }))

	return g.TreeNode(p.Cmd, g.TreeNodeFlagsNone, children)
}

func (a *App) Loop() {
	g.SingleWindow("Quick Kill!", a.ProcessRows())
}

func main() {
	a := App{}

	a.Processes()
	// go func() {
	// 	for range time.Tick(time.Second) {
	// 		a.Processes()
	// 	}
	// }()

	spew.Dump(a.pidMap[13680])

	wnd := g.NewMasterWindow("Quick Kill", 500, 500, g.MasterWindowFlagsNotResizable, nil)
	wnd.Main(a.Loop)
}
