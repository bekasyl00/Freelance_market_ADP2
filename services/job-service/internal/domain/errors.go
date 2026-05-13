package domain

import "errors"

var (
	ErrInvalidInput       = errors.New("invalid input")
	ErrJobNotFound        = errors.New("job not found")
	ErrJobNotOpen         = errors.New("job is not open for applications")
	ErrDuplicateProposal  = errors.New("freelancer already applied to this job")
	ErrForbiddenClient    = errors.New("client does not own this job")
	ErrJobAlreadyTerminal = errors.New("job is already completed or cancelled")
)
