package main

import (
	"fmt"
	"log"
	"os"
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
		if strings.Index(strings.ToLower(proc.Cmd), strings.ToLower(a.filterWord)) == -1 {
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
		parent, hasParent := a.pidMap[p.PPid]
		if hasParent {
			if parent.Cmd != p.Cmd {
				v = append(v, g.Line(a.ProcessWidget(p)), g.Separator())
			}
		} else {
			v = append(v, g.Line(a.ProcessWidget(p)), g.Separator())
		}
	}
	return v
}

func (a *App) ProcessWidget(p Process) g.Widget {
	if len(p.Children) == 0 {
		return g.Line(
			g.TreeNode(p.Cmd, g.TreeNodeFlagsLeaf, nil),
			g.SmallButton(fmt.Sprintf("Select PID %d", p.Pid), func() {
				a.selectedProcess = p
			}),
		)

	}

	children := []g.Widget{}
	for _, c := range p.Children {
		children = append(children, a.ProcessWidget(*c))
	}
	children = append(children,
		g.SmallButton(fmt.Sprintf("Select PID %d", p.Pid), func() {
			a.selectedProcess = p
		}),
	)

	return g.TreeNode(p.Cmd, g.TreeNodeFlagsNone, children)

}

func (a *App) Loop() {
	g.SingleWindow("Quick Kill!", g.Layout{
		g.SplitLayout("MainSplit", g.DirectionHorizontal, true, 500, g.Layout(a.ProcessRows()),
			g.Layout{
				g.SplitLayout("SearchSplit", g.DirectionVertical, true, 55,
					g.Layout{
						g.Label("Search"),
						g.InputText("", 200, &a.filterWord),
					},
					g.Layout{
						g.LabelWrapped(fmt.Sprintf("Selected Process %s with PID %d", a.selectedProcess.Cmd, a.selectedProcess.Pid)),
						g.Label("Press F10 to kill."),
					},
				),
			},
		)})
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

	wnd := g.NewMasterWindow("Quick Kill", 800, 500, 0, nil)
	wnd.Main(a.Loop)
}
