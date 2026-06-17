package esc_core

import "fmt"

type typedColumn[T any] struct {
	values []T
}

type column interface {
	Len() int
	AppendZero()
	AppendAny(value any) error
	SwapRemove(row int)
	CopyValueTo(row int, targetColumn column) error
	ValueAny(row int) any
	PtrAny(row int) any
	SetAny(row int, value any) error
}

func newTypedColumn[T any]() *typedColumn[T] {
	return &typedColumn[T]{}
}

func (c *typedColumn[T]) Len() int {
	return len(c.values)
}

func (c *typedColumn[T]) AppendZero() {
	var zero T
	c.values = append(c.values, zero)
}

func (c *typedColumn[T]) AppendAny(value any) error {
	var v, ok = value.(T)

	if !ok {
		return fmt.Errorf("%w: column append got %T, want %T",
			ErrInvalidComponentType,
			value,
			*new(T),
		)
	}

	c.values = append(c.values, v)
	return nil
}

func (c *typedColumn[T]) SwapRemove(row int) {
	var last = len(c.values) - 1

	var zero T

	c.values[row] = c.values[last]
	c.values[last] = zero
	c.values = c.values[:last]
}

func (c *typedColumn[T]) CopyValueTo(row int, targetColumn column) error {
	return targetColumn.AppendAny(c.values[row])
}

func (c *typedColumn[T]) ValueAny(row int) any {
	return c.values[row]
}

func (c *typedColumn[T]) PtrAny(row int) any {
	return &c.values[row]
}

func (c *typedColumn[T]) SetAny(row int, value any) error {
	var v, ok = value.(T)

	if !ok {
		return fmt.Errorf("%w: column append got %T, want %T",
			ErrInvalidComponentType,
			value,
			*new(T),
		)
	}

	c.values[row] = v

	return nil
}
