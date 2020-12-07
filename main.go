package main

import (
	"syscall"
	"unsafe"

	g "github.com/AllenDang/giu"
	"github.com/JamesHovious/w32"
)

func getProcesses() []string {
	v := []string{}
	ps := make([]uint32, 255)
	var read uint32
	if !w32.EnumProcesses(ps, uint32(len(ps)), &read) {
		panic("couldn't read")
		return v
	}
	for _, p := range ps {
		v = append(v, getProcessName(p))
	}
	return v
}

func getModuleInfo(me32 *w32.MODULEENTRY32) string {
	procName := syscall.UTF16ToString(me32.SzModule[:])
	return procName
}

func getProcessName(pid uint32) string {
	snap := w32.CreateToolhelp32Snapshot(w32.TH32CS_SNAPMODULE, pid)
	if snap == 0 {
		return ""
	}
	defer w32.CloseHandle(snap)

	var me32 w32.MODULEENTRY32
	me32.Size = uint32(unsafe.Sizeof(me32))
	if !w32.Module32First(snap, &me32) {
		return ""
	}
	return getModuleInfo(&me32)
}

func loop() {
	g.SingleWindow("hello world", g.Layout{
		g.ListBoxV("pids", 250, 500-20, true, getProcesses(), nil, nil, nil, nil),
	})
}

func main() {

	wnd := g.NewMasterWindow("Hello world", 500, 500, g.MasterWindowFlagsNotResizable, nil)
	wnd.Main(loop)
}
