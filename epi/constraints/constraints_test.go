package constraints_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/composer/epi"
	"github.com/mandelsoft/composer/epi/constraints"
)

// The package cannot be dot-imported here: Or, And and Not collide with the
// matchers of the same name provided by gomega.

////////////////////////////////////////////////////////////////////////////////
// test frames and states

type someState interface {
	isSomeState()
}

type otherState interface {
	isOtherState()
}

type someStateImpl struct{}

var _ someState = (*someStateImpl)(nil)

func (*someStateImpl) isSomeState() {}

// frameA and frameB are plain element frames without any state.
type frameA struct {
	epi.DefaultFrame[epi.None]
}

var _ epi.Frame = (*frameA)(nil)

type frameB struct {
	epi.DefaultFrame[epi.None]
}

var _ epi.Frame = (*frameB)(nil)

// stateElementFrame is an element frame offering a state by embedding
// the state implementation.
type stateElementFrame struct {
	epi.DefaultFrame[epi.None]
	someStateImpl
}

var _ someState = (*stateElementFrame)(nil)

// stateProvider is no frame, but a state added to the frame stack,
// which delegates to the effective state.
type stateProvider struct {
	state any
}

var _ epi.StateProvider = (*stateProvider)(nil)

func (p *stateProvider) GetState() any {
	return p.state
}

////////////////////////////////////////////////////////////////////////////////
// helpers

// assureElementFrames assures that a frame list fulfills the contract of
// a constraint: it must contain element frames, only.
func assureElementFrames(list []epi.Frame) {
	GinkgoHelper()
	for i, f := range list {
		Expect(epi.IsElementFrame(f)).To(BeTrue(), "frame %d (%T) is no element frame", i, f)
	}
}

// frames provides a frame list for a constraint check.
func frames(f ...epi.Frame) []epi.Frame {
	GinkgoHelper()
	assureElementFrames(f)
	return f
}

func succeed(inner, outer []epi.Frame) error {
	return nil
}

func fail(err error) constraints.Constraint {
	return func(inner, outer []epi.Frame) error {
		return err
	}
}

// recorder is a constraint keeping track of its invocations.
type recorder struct {
	calls int
	inner []epi.Frame
	outer []epi.Frame
	err   error
}

func (r *recorder) check(inner, outer []epi.Frame) error {
	r.calls++
	r.inner = inner
	r.outer = outer
	return r.err
}

// errNoState is reported by check, if the requested state is not found.
var errNoState = fmt.Errorf("state not found")

// check evaluates a constraint the same way as epi.EvaluateWithState does it
// for the actual frame stack of an environment with the default state
// extractor: a missing state is an error, the constraint is not evaluated.
// The state epi.None is always provided by the initial frame of an
// environment, it is used to check constraints for elements without
// required state.
func check[S any](env epi.EnvState, c constraints.Constraint) error {
	GinkgoHelper()
	_, inner, outer, ok := epi.GetStateAndEmbedding[S](env)
	if !ok {
		return errNoState
	}
	assureElementFrames(inner)
	assureElementFrames(outer)
	return c(inner, outer)
}

////////////////////////////////////////////////////////////////////////////////

var _ = Describe("Constraints", func() {
	var (
		errA = fmt.Errorf("error A")
		errB = fmt.Errorf("error B")

		a = &frameA{}
		b = &frameB{}
	)

	Context("Or", func() {
		It("fails without constraint", func() {
			Expect(constraints.Or()(nil, nil)).To(HaveOccurred())
		})

		It("succeeds if any constraint succeeds", func() {
			Expect(constraints.Or(fail(errA), succeed)(nil, nil)).To(Succeed())
			Expect(constraints.Or(succeed, fail(errA))(nil, nil)).To(Succeed())
		})

		It("fails with a generic error if all constraints fail", func() {
			Expect(constraints.Or(fail(errA), fail(errB))(nil, nil)).To(MatchError("not possible in actual context"))
		})

		It("stops at the first succeeding constraint", func() {
			r := &recorder{}
			Expect(constraints.Or(succeed, r.check)(nil, nil)).To(Succeed())
			Expect(r.calls).To(Equal(0))
		})

		It("forwards the frames", func() {
			r := &recorder{}
			Expect(constraints.Or(r.check)(frames(a), frames(b))).To(Succeed())
			Expect(r.inner).To(Equal(frames(a)))
			Expect(r.outer).To(Equal(frames(b)))
		})
	})

	Context("And", func() {
		It("succeeds without constraint", func() {
			Expect(constraints.And()(nil, nil)).To(Succeed())
		})

		It("succeeds if all constraints succeed", func() {
			Expect(constraints.And(succeed, succeed)(nil, nil)).To(Succeed())
		})

		It("fails with the error of the first failing constraint", func() {
			Expect(constraints.And(succeed, fail(errA), fail(errB))(nil, nil)).To(MatchError(errA))
		})

		It("stops at the first failing constraint", func() {
			r := &recorder{}
			Expect(constraints.And(fail(errA), r.check)(nil, nil)).To(MatchError(errA))
			Expect(r.calls).To(Equal(0))
		})

		It("forwards the frames", func() {
			r := &recorder{}
			Expect(constraints.And(r.check)(frames(a), frames(b))).To(Succeed())
			Expect(r.inner).To(Equal(frames(a)))
			Expect(r.outer).To(Equal(frames(b)))
		})
	})

	Context("Not", func() {
		It("inverts the result", func() {
			Expect(constraints.Not(succeed)(nil, nil)).To(MatchError("not possible in actual context"))
			Expect(constraints.Not(fail(errA))(nil, nil)).To(Succeed())
		})

		It("forwards the frames", func() {
			r := &recorder{}
			Expect(constraints.Not(r.check)(frames(a), frames(b))).NotTo(Succeed())
			Expect(r.inner).To(Equal(frames(a)))
			Expect(r.outer).To(Equal(frames(b)))
		})
	})

	Context("FrameTypeConstraint", func() {
		c := constraints.FrameTypeConstraint[*frameA]

		It("finds the frame type at any position of the inner frames", func() {
			Expect(c(frames(a), nil)).To(Succeed())
			Expect(c(frames(a, b), nil)).To(Succeed())
			Expect(c(frames(b, a), nil)).To(Succeed())
		})

		It("fails for other frame types", func() {
			Expect(c(frames(b), nil)).To(MatchError("not possible in actual context"))
		})

		It("fails for no frames", func() {
			Expect(c(nil, nil)).To(HaveOccurred())
		})

		It("ignores the outer frames", func() {
			Expect(c(nil, frames(a))).To(HaveOccurred())
			Expect(c(frames(b), frames(a))).To(HaveOccurred())
		})

		It("matches interface types", func() {
			Expect(constraints.FrameTypeConstraint[epi.Frame](frames(a), nil)).To(Succeed())
		})
	})

	Context("StateTypeConstraint", func() {
		c := constraints.StateTypeConstraint[someState]

		It("finds a state implemented by the frame", func() {
			Expect(c(frames(&stateElementFrame{}), nil)).To(Succeed())
		})

		It("finds the state at any position of the inner frames", func() {
			Expect(c(frames(a, &stateElementFrame{}, b), nil)).To(Succeed())
		})

		It("fails for frames without a matching state", func() {
			Expect(c(frames(a), nil)).To(MatchError("not possible in actual context"))
			Expect(constraints.StateTypeConstraint[otherState](frames(&stateElementFrame{}), nil)).To(HaveOccurred())
		})

		It("fails for no frames", func() {
			Expect(c(nil, nil)).To(HaveOccurred())
		})

		It("ignores the outer frames", func() {
			Expect(c(nil, frames(&stateElementFrame{}))).To(HaveOccurred())
		})
	})

	Context("StateFrame", func() {
		It("succeeds if there are no inner frames", func() {
			Expect(constraints.StateFrame(nil, nil)).To(Succeed())
			Expect(constraints.StateFrame(frames(), nil)).To(Succeed())
			Expect(constraints.StateFrame(nil, frames(a))).To(Succeed())
		})

		It("fails if there are inner frames", func() {
			Expect(constraints.StateFrame(frames(a), nil)).To(HaveOccurred())
			Expect(constraints.StateFrame(frames(a), frames(b))).To(HaveOccurred())
		})
	})

	Context("TopLevel", func() {
		It("succeeds if there are no frames at all", func() {
			Expect(constraints.TopLevel(nil, nil)).To(Succeed())
			Expect(constraints.TopLevel(frames(), frames())).To(Succeed())
		})

		It("fails if there are frames", func() {
			Expect(constraints.TopLevel(frames(a), nil)).To(MatchError("not possible as top-level element"))
			Expect(constraints.TopLevel(nil, frames(a))).To(MatchError("not possible as top-level element"))
			Expect(constraints.TopLevel(frames(a), frames(b))).To(HaveOccurred())
		})
	})

	Context("ApplyToFiltered", func() {
		It("passes the matching frames and preserves their order and origin", func() {
			r := &recorder{}
			a1, a2, a3 := &frameA{}, &frameA{}, &frameA{}

			c := constraints.ApplyToFiltered(constraints.FrameTypeConstraint[*frameA], r.check)
			Expect(c(frames(a1, b, a2), frames(b, a3))).To(Succeed())
			Expect(r.inner).To(Equal(frames(a1, a2)))
			Expect(r.outer).To(Equal(frames(a3)))
		})

		It("passes empty lists if nothing matches", func() {
			r := &recorder{}

			c := constraints.ApplyToFiltered(constraints.FrameTypeConstraint[*frameA], r.check)
			Expect(c(frames(b), frames(b))).To(Succeed())
			Expect(r.calls).To(Equal(1))
			Expect(r.inner).To(BeEmpty())
			Expect(r.outer).To(BeEmpty())
		})

		It("evaluates the filter per frame", func() {
			f := &recorder{}

			c := constraints.ApplyToFiltered(f.check, succeed)
			Expect(c(frames(a), frames(b))).To(Succeed())
			Expect(f.calls).To(Equal(2))
			Expect(f.inner).To(Equal(frames(b)))
			Expect(f.outer).To(BeNil())
		})

		It("reports the error of the constraint", func() {
			c := constraints.ApplyToFiltered(succeed, fail(errA))
			Expect(c(frames(a), nil)).To(MatchError(errA))
		})
	})

	Context("DirectEmbedding", func() {
		c := constraints.DirectEmbedding(constraints.FrameTypeConstraint[*frameA])

		It("fails for no frames without evaluating the constraint", func() {
			r := &recorder{}
			Expect(constraints.DirectEmbedding(r.check)(nil, nil)).To(MatchError("not possible in actual context"))
			Expect(r.calls).To(Equal(0))
		})

		It("evaluates the latest inner frame only", func() {
			r := &recorder{}
			Expect(constraints.DirectEmbedding(r.check)(frames(a, b), frames(&frameB{}))).To(Succeed())
			Expect(r.inner).To(Equal(frames(a)))
			Expect(r.outer).To(BeEmpty())
		})

		It("evaluates the latest outer frame without inner frames", func() {
			r := &recorder{}
			Expect(constraints.DirectEmbedding(r.check)(nil, frames(a, b))).To(Succeed())
			Expect(r.inner).To(Equal(frames(a)))
			Expect(r.outer).To(BeEmpty())
		})

		It("checks the latest inner frame", func() {
			Expect(c(frames(a, b), nil)).To(Succeed())
			Expect(c(frames(b, a), nil)).To(HaveOccurred())
		})

		It("checks the latest outer frame without inner frames", func() {
			Expect(c(nil, frames(a, b))).To(Succeed())
			Expect(c(nil, frames(b, a))).To(HaveOccurred())
		})

		It("prefers inner frames to outer frames", func() {
			Expect(c(frames(b), frames(a))).To(HaveOccurred())
		})

		It("reports the error of the constraint", func() {
			Expect(constraints.DirectEmbedding(fail(errA))(frames(a), nil)).To(MatchError(errA))
		})
	})

	Context("frame stack of an environment", func() {
		isA := constraints.FrameTypeConstraint[*frameA]
		isB := constraints.FrameTypeConstraint[*frameB]

		var env epi.EnvState

		BeforeEach(func() {
			env = epi.NewEnvState()
		})

		It("passes element frames only", func() {
			r := &recorder{}

			env.With(0, a, func() {
				env.AddState(&someStateImpl{})
				env.With(0, b, func() {
					env.AddState(&someStateImpl{})

					Expect(check[someState](env, r.check)).To(Succeed())
					Expect(r.inner).To(BeEmpty())
					Expect(r.outer).To(Equal(frames(b, a)))
				})
			})
		})

		It("requires the state on an empty stack", func() {
			Expect(check[someState](env, constraints.TopLevel)).To(MatchError(errNoState))
			Expect(check[someState](env, constraints.StateFrame)).To(MatchError(errNoState))
			Expect(check[someState](env, succeed)).To(MatchError(errNoState))
		})

		It("evaluates to top-level on an empty stack without required state", func() {
			Expect(check[epi.None](env, constraints.TopLevel)).To(Succeed())
			Expect(check[epi.None](env, constraints.StateFrame)).To(Succeed())
			Expect(check[epi.None](env, constraints.DirectEmbedding(succeed))).To(HaveOccurred())
		})

		It("sees a top-level state frame", func() {
			env.AddState(&someStateImpl{})

			Expect(check[someState](env, constraints.TopLevel)).To(Succeed())
			Expect(check[someState](env, constraints.StateFrame)).To(Succeed())
		})

		It("requires the state below element frames without state", func() {
			env.With(0, a, func() {
				env.With(0, b, func() {
					Expect(check[someState](env, succeed)).To(MatchError(errNoState))
				})
			})
		})

		It("handles element frames without state as inner frames without required state", func() {
			env.With(0, a, func() {
				env.With(0, b, func() {
					Expect(check[epi.None](env, constraints.TopLevel)).To(HaveOccurred())
					Expect(check[epi.None](env, constraints.StateFrame)).To(HaveOccurred())
					Expect(check[epi.None](env, isA)).To(Succeed())
					Expect(check[epi.None](env, isB)).To(Succeed())
					Expect(check[epi.None](env, constraints.DirectEmbedding(isB))).To(Succeed())
					Expect(check[epi.None](env, constraints.DirectEmbedding(isA))).To(HaveOccurred())
				})
			})
		})

		It("splits the stack at the state", func() {
			env.With(0, a, func() {
				env.AddState(&someStateImpl{})
				env.With(0, b, func() {
					Expect(check[someState](env, constraints.StateFrame)).To(HaveOccurred())
					Expect(check[someState](env, isB)).To(Succeed())
					Expect(check[someState](env, isA)).To(HaveOccurred())
				})
			})
		})

		It("finds a state behind state providers", func() {
			env.With(0, a, func() {
				env.AddState(&stateProvider{state: &stateProvider{state: &someStateImpl{}}})

				Expect(check[someState](env, constraints.StateFrame)).To(Succeed())
				Expect(check[someState](env, constraints.DirectEmbedding(isA))).To(Succeed())
			})
		})

		It("does not find a state in a state provider without a matching state", func() {
			env.With(0, a, func() {
				env.AddState(&stateProvider{})
				env.AddState(&stateProvider{state: "other"})

				Expect(check[someState](env, succeed)).To(MatchError(errNoState))
			})
		})

		It("includes an element frame providing the state into the inner frames", func() {
			env.With(0, a, func() {
				env.With(0, &stateElementFrame{}, func() {
					Expect(check[someState](env, constraints.StateFrame)).To(HaveOccurred())
					Expect(check[someState](env, constraints.StateTypeConstraint[someState])).To(Succeed())
					Expect(check[someState](env, isA)).To(HaveOccurred())
				})
			})
		})

		It("detects the outer element on top of a state frame", func() {
			env.With(0, a, func() {
				env.AddState(&someStateImpl{})

				Expect(check[someState](env, constraints.TopLevel)).To(HaveOccurred())
				Expect(check[someState](env, constraints.StateFrame)).To(Succeed())
				Expect(check[someState](env, constraints.DirectEmbedding(isA))).To(Succeed())
				Expect(check[someState](env, constraints.DirectEmbedding(isB))).To(HaveOccurred())
			})
		})

		It("skips state frames between the element and its embedding element", func() {
			env.With(0, a, func() {
				env.AddState(&someStateImpl{})
				env.AddState(&someStateImpl{})
				env.With(0, b, func() {
					env.AddState(&someStateImpl{})

					Expect(check[someState](env, constraints.DirectEmbedding(isB))).To(Succeed())
					Expect(check[someState](env, constraints.DirectEmbedding(isA))).To(HaveOccurred())
				})
			})
		})

		It("supports the embedding rule for nestable elements", func() {
			// an element is only possible directly on a state frame or
			// embedded in an element providing the state.
			c := constraints.Or(constraints.StateFrame, constraints.DirectEmbedding(constraints.StateTypeConstraint[someState]))

			env.AddState(&someStateImpl{})
			Expect(check[someState](env, c)).To(Succeed())

			env.With(0, &stateElementFrame{}, func() {
				Expect(check[someState](env, c)).To(Succeed())
				env.With(0, a, func() {
					Expect(check[someState](env, c)).To(HaveOccurred())
				})
			})
		})

		It("supports the rule for non-nestable elements", func() {
			// an element must not be nested in an element of the same type,
			// as long as there is no new state.
			c := constraints.Or(constraints.StateFrame, constraints.Not(isA))

			Expect(check[someState](env, c)).To(MatchError(errNoState))
			env.AddState(&someStateImpl{})
			Expect(check[someState](env, c)).To(Succeed())
			env.With(0, a, func() {
				err := check[someState](env, c)
				Expect(err).To(HaveOccurred())
				Expect(err).NotTo(MatchError(errNoState))
				env.AddState(&someStateImpl{})
				Expect(check[someState](env, c)).To(Succeed())
			})
		})
	})
})
