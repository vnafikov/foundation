package parameters

import "github.com/vnafikov/common/validation"

type Validator interface {
	IsValid() bool
}

type NormalizerValidator interface {
	Validator
	Normalize()
}

type ValidatorWithErrors interface {
	Validate() validation.Validator
}

type NormalizerValidatorWithErrors interface {
	ValidatorWithErrors
	Normalize()
}
