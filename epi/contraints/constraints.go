package contraints

import (
	"fmt"

	// avoid package cycle
	"github.com/mandelsoft/composer/epi/internal"
)

// Constraint checks the frame stack down to the found state
// whether it is applicable for an element.
// inner are the elements on-top of the state (inclusive),
// and outer the elements before the state, both lists
// are ordered latest to oldest.
type Constraint func(inner, outer []internal.Frame) error

// Or composes a logical OR of several constraints.
func Or(cs ...Constraint) Constraint {
	return func(inner, outer []internal.Frame) error {
		for _, f := range cs {
			if err := f(inner, outer); err == nil {
				return nil
			}
		}
		return fmt.Errorf("not possible in actual context")
	}
}

// Or composes a logical AND of several constraints.
func And(cs ...Constraint) Constraint {
	return func(inner, outer []internal.Frame) error {
		for _, f := range cs {
			if err := f(inner, outer); err != nil {
				return err
			}
		}
		return nil
	}
}

// Or composes a logical NOT of a constraint.
func Not(cs Constraint) Constraint {
	return func(inner, outer []internal.Frame) error {
		if cs(inner, outer) == nil {
			return fmt.Errorf("not possible in actual context")
		}
		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////

// FrameTypeConstraint checks for a dedicated frame type down to found state.
func FrameTypeConstraint[F internal.Frame](inner, outer []internal.Frame) error {
	for _, frame := range inner {
		if _, ok := any(frame).(F); ok {
			return nil
		}
	}
	return fmt.Errorf("not possible in actual context")
}

// StateTypeConstraint checks for a dedicated frame state type down to found state.
func StateTypeConstraint[S any](inner, outer []internal.Frame) error {
	for _, frame := range inner {
		_, ok := internal.GetFrameState[S](frame)
		if ok {
			return nil
		}
	}
	return fmt.Errorf("not possible in actual context")
}

// DirectEmbedding checks whether the embedding element
// satisfies a constraint.
func DirectEmbedding(cs Constraint) Constraint {
	return func(inner, outer []internal.Frame) error {
		if len(inner) == 0 && len(outer) == 0 {
			return fmt.Errorf("not possible in actual context")
		}
		if len(inner) != 0 {
			return cs(inner[0:1], nil)
		}
		return cs(outer[0:1], nil)
	}
}

// StateFrame checks whether it is on top of a state frame.
func StateFrame(inner, outer []internal.Frame) error {
	if len(inner) != 0 {
		return fmt.Errorf("not possible as top-level element")
	}
	return nil
}

// TopLevel checks whether it is a Top-Level element
func TopLevel(inner, outer []internal.Frame) error {
	if len(inner) != 0 || len(outer) != 0 {
		return fmt.Errorf("not possible as top-level element")
	}
	return nil
}

// ApplyToFiltered filters the frames by a constraint filter, the
// result is passed to the effectice constraint cs.
func ApplyToFiltered(filter Constraint, cs Constraint) Constraint {
	return func(inner, outer []internal.Frame) error {
		var finner []internal.Frame
		var fouter []internal.Frame

		for _, frame := range inner {
			if err := filter([]internal.Frame{frame}, nil); err == nil {
				finner = append(finner, frame)
			}
		}
		for _, frame := range outer {
			if err := filter([]internal.Frame{frame}, nil); err == nil {
				fouter = append(fouter, frame)
			}
		}
		return cs(finner, fouter)
	}
}
