package esc_core

import "Velora/esc"

type StageID uint16

type Context struct {
	World        *World
	Commands     *esc.CommandBuffer //old
	Resources    *esc.Resources     //old
	Tick         uint64
	DeltaSeconds float32
	Stage        StageID
}

type System interface {
	Name() string
	Stage() StageID
	Update(ctx *Context) error
	Access() AccessSet
	DebugQueries() []QueryDebugInfo
}
