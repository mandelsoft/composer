package epi

// Option can be used to configure an Environment.
// If elements specific for a functional group
// are required, the environment must be mapped to the
// Group implementation. This should be done
// by a private interface method implemented by the Group.
// If this is not possible an appropriate error
// ErrGroupNotSupported should be provided.
type Option interface {
	ApplyTo(e Environment) error
}

func ApplyOptionsTo(e Environment, opts ...Option) error {
	for _, o := range opts {
		err := o.ApplyTo(e)
		if err != nil {
			return err
		}
	}
	return nil
}

func WithOptionsApplied[E Environment](e E, opts ...Option) (E, error) {
	var _nil E
	err := ApplyOptionsTo(e, opts...)
	if err != nil {
		return _nil, err
	}
	return e, nil
}
