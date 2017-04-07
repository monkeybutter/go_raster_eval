package object

import (
	"fmt"
	"image"
)

type ObjectType string

const (
	NULL_OBJ  = "NULL"
	ERROR_OBJ = "ERROR"

	RASTER_OBJ  = "RASTER"
	NUMBER_OBJ  = "NUMBER"
	BOOLEAN_OBJ = "BOOLEAN"

	RETURN_VALUE_OBJ = "RETURN_VALUE"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Raster struct {
	Value image.Image
}

func (i *Raster) Type() ObjectType { return RASTER_OBJ }
func (i *Raster) Inspect() string {
	switch t := i.Value.(type) {
	case *image.Gray:
		return fmt.Sprintf("%v", t)
	case *image.Gray16:
		return fmt.Sprintf("%v", t)
	default:
		return "ERROR: Raster type not recognised"
	}

}

type Number struct {
	Value float64
}

func (i *Number) Type() ObjectType { return NUMBER_OBJ }
func (i *Number) Inspect() string  { return fmt.Sprintf("%f", i.Value) }

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }

type Null struct{}

func (n *Null) Type() ObjectType { return NULL_OBJ }
func (n *Null) Inspect() string  { return "null" }

type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }
func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }

type Error struct {
	Message string
}

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string  { return "ERROR: " + e.Message }
