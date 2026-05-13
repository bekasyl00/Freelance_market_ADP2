package domain

import "errors"

var (
	ErrInvalidCredentials   = errors.New("invalid email or password")
	ErrOnlyFreelancerSkills = errors.New("only freelancers can update skills")
	ErrInvalidRole          = errors.New("invalid role")
)
