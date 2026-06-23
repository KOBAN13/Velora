package esc_core

import (
	"errors"
)

type CommandBuffer struct {
	commands []Command
	errors   []error
}

type Command interface {
	Apply(world *World) error
}

func (c *CommandBuffer) Add() *SpawnCommandBuilder {
	return &SpawnCommandBuilder{}
}

func (c *CommandBuffer) Apply(world *World) error {
	if len(c.errors) > 0 {
		return errors.Join(c.errors...)
	}

	var prev = world.
}

func (c *CommandBuffer) Clear() {
	clear(c.commands)
	c.commands = c.commands[:0]
	clear(c.errors)
	c.errors = c.errors[:0]
}
