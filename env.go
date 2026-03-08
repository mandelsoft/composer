//go:generate go tool mdref --list doc/src .
package composer

import (
	"github.com/mandelsoft/composer/epi"
)

type Environment = epi.Environment

type StateExtractor[S any] = epi.StateExtractor[S]

// GetState extract state of a given type interface from an environment.
// Optionally a StateProvider can be given used to provide a more complex
// composed state.
func GetState[S any](p Environment, ext ...StateExtractor[S]) (S, bool) {
	return epi.GetState[S](p, ext...)
}
