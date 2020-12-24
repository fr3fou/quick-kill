package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"time"

	g "github.com/AllenDang/giu"
	"github.com/go-vgo/robotgo"
	"github.com/mitchellh/go-ps"
)

// Events
const (
	KeyUp = 5
)

// F keys
const (
	F10 = 121
)

type Process struct {
	Pid      int
	PPid     int
	Cmd      string
	Children []*Process
}

type App struct {
	processes       []Process
	pidMap          map[int]*Process
	filterWord      string
	selectedProcess Process
	renderedPids    []int
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
		if !matchesQuery(proc, a.filterWord) {
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

func matchesQuery(proc *Process, query string) bool {
	for _, child := range proc.Children {
		if matchesQuery(child, query) {
			return true
		}
	}

	return strings.Index(strings.ToLower(proc.Cmd), strings.ToLower(query)) != -1
}

func (a *App) ProcessRows() []g.Widget {
	v := []g.Widget{}
	a.renderedPids = a.renderedPids[:0]
	for _, p := range a.processes {
		parent, hasParent := a.pidMap[p.PPid]
		if contains(a.renderedPids, p.Pid) || (hasParent && parent.Cmd == p.Cmd) {
			continue
		}

		v = append(v, g.Line(a.ProcessWidget(p)), g.Separator())
	}
	return v
}

func contains(i []int, n int) bool {
	for _, v := range i {
		if v == n {
			return true
		}
	}
	return false
}

func (a *App) ProcessWidget(p Process) g.Widget {
	a.renderedPids = append(a.renderedPids, p.Pid)

	if len(p.Children) == 0 {
		return g.Line(
			g.TreeNodeV(p.Cmd, g.TreeNodeFlagsLeaf, func() {
				if g.IsItemClicked(g.MouseButtonLeft) {
					a.selectedProcess = p
				}
			}, nil),
		)
	}

	children := []g.Widget{}
	for _, c := range p.Children {
		children = append(children, a.ProcessWidget(*c))
		a.renderedPids = append(a.renderedPids, c.Pid)
	}

	return g.TreeNodeV(p.Cmd, g.TreeNodeFlagsOpenOnArrow, func() {
		if g.IsItemClicked(g.MouseButtonLeft) {
			a.selectedProcess = p
		}
	}, children)

}

func (a *App) Loop() {
	g.SingleWindow("Quick Kill!", g.Layout{
		g.Child("Search", true, -1, 55, g.WindowFlagsNone, g.Layout{
			g.Label("Search"),
			g.InputText("", -1, &a.filterWord),
		}),
		g.Child("LeftPart", true, -1, 600-150, g.WindowFlagsNone, g.Layout(a.ProcessRows())),
		g.Child("Kill", true, -1, -1, g.WindowFlagsNone, g.Layout{
			g.LabelWrapped(fmt.Sprintf("Selected Process %s with PID %d", a.selectedProcess.Cmd, a.selectedProcess.Pid)),
			g.Label("Press F10 to kill."),
			g.Button("Made by fr3fou", func() {
				openURL("https://twitter.com/fr3fou")
			}),
		}),
	})
}

func main() {
	a := App{selectedProcess: Process{Pid: -1, Cmd: "None"}}

	a.Processes()
	go func() {
		ticker := time.NewTicker(time.Second * 1)
		for {
			a.Processes()
			g.Update()
			<-ticker.C
		}
	}()

	// Start async event listener
	hook := robotgo.EventStart()
	defer robotgo.EventEnd()

	go func() {
		for v := range hook {
			if v.Kind != KeyUp {
				continue
			}

			if v.Rawcode == F10 && a.selectedProcess.Pid != -1 {
				proc, err := os.FindProcess(a.selectedProcess.Pid)
				if err != nil {
					log.Println(err)
					continue
				}

				if err := proc.Kill(); err != nil {
					log.Println("Failed killing process")
				}
			}
		}
	}()

	wnd := g.NewMasterWindow("Quick Kill", 500, 600, g.MasterWindowFlagsNotResizable, nil)
	wnd.Main(a.Loop)
}

func openURL(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	if err != nil {
		log.Println(err)
	}
}
