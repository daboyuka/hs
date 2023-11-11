package scope

import (
	"github.com/daboyuka/hs/program/record"
)

type Func func(args ...record.Record) (record.Record, error)

type FuncTable struct {
	parent *FuncTable
	funcs  map[string]Func
}

func NewFuncTable(parent *FuncTable, funcs map[string]Func) *FuncTable {
	return &FuncTable{parent: parent, funcs: funcs}
}

func (ft *FuncTable) Get(name string) Func {
	if ft == nil {
		return nil
	}
	if f := ft.funcs[name]; f != nil {
		return f
	}
	return ft.parent.Get(name)
}
