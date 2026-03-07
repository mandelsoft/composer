// Package epi provided the Environment Programming Interface.
// An environment at least consists of a core part the EnvState object,
// core environment functions provided by the Group type
//
// A final Environment consist of the core group and optionally
// any number of additional functional Group implementations.
// Anew functional area may create an own Group type implementing additional
// environment functions. Every Group object part of an environment must share
// the same EnvState.
//
// A final environment is then a struct embedding all desired functional Group
// objects. An example can be seen in package [github.com/mandelsoft/composer/common].
//
// The core Group provided some basic API function like Environment.With, used
// to execute some code with an additional state, or Environment.AddState used
// to add state as part of a command sequence.
//
// The EnvState is responsible for the
// state handling of nested calls and shared among all groups used to compose
// an end user Environment. It should only be visible for functional area
// implementations as member of the area's Group type.
//
// Every functional method may require some outer state. Such state is defined
// by a state interface used to access the state.
// The method EvaluateWithState is used by functional method to request a particular
// state (use None is no state is required). This state is then passed to a FrameProvider.
// The FrameProvider's task is to setup a new Frame representing the nesting level
// for Block s passed to the functional method. THose Block s will then be
// executed with the additional Frame provided by the method. After the Block are
// completed the Frame is closed with Frame.Close and removed from the frame stack held by
// the EnvState. The close method can do dome cleanup or finalization for the functional
// element managed by the functional method.
//
// All those function may return an error, every error directly aborts the further
// processing
package epi
