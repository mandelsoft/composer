package contraints

import (
	"fmt"

	// avoid package cycle
	"github.com/mandelsoft/composer/epi/internal"
)

// Constraint checks the frame stack down to the found state
// whether it is applicable for an element.
type Constraint func([]internal.Frame) error

// Or composes a logical OR of several constraints.
func Or(cs ...Constraint) Constraint {
	return func(frames []internal.Frame) error {
		for _, f := range cs {
			if err := f(frames); err == nil {
				return nil
			}
		}
		return fmt.Errorf("not possible in actual context")
	}
}

// Or composes a logical AND of several constraints.
func And(cs ...Constraint) Constraint {
	return func(frames []internal.Frame) error {
		for _, f := range cs {
			if err := f(frames); err != nil {
				return err
			}
		}
		return nil
	}
}

// Or composes a logical NOT of a constraint.
func Not(cs Constraint) Constraint {
	return func(frames []internal.Frame) error {
		if cs(frames) == nil {
			return fmt.Errorf("not possible in actual context")
		}
		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////

// FrameTypeConstraint checks for a dedicated frame type.
func FrameTypeConstraint[F internal.Frame](frames []internal.Frame) error {
	for _, frame := range frames {
		if _, ok := any(frame).(F); ok {
			return nil
		}
	}
	return fmt.Errorf("not possible in actual context")
}

// StateTypeConstraint checks for a dedicated state type.
func StateTypeConstraint[S any](frames []internal.Frame) error {
	for _, frame := range frames {
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
	return func(frames []internal.Frame) error {
		if len(frames) == 0 {
			return fmt.Errorf("not possible in actual context")
		}
		return cs(frames[len(frames)-1:])
	}
}

// TopLevel checks whether it is a Top-Level element
func TopLevel(frames []internal.Frame) error {
	if len(frames) != 0 {
		return fmt.Errorf("not possible as top-level element")
	}
	return nil
}

// ApplyToFiltered filters the frames by a constraint filter, the
// result is passed to the effectice constraint cs.
func ApplyToFiltered(filter Constraint, cs Constraint) Constraint {
	return func(frames []internal.Frame) error {
		if len(frames) != 0 {
			return fmt.Errorf("not possible as top-level element")
		}
		var filtered []internal.Frame

		for _, frame := range frames {
			if err := filter([]internal.Frame{frame}); err == nil {
				filtered = append(filtered, frame)
			}
		}
		return cs(filtered)
	}
}
