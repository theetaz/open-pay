package domain

import "errors"

var (
	ErrInvalidMerchant      = errors.New("invalid merchant data")
	ErrInvalidKYCTransition = errors.New("invalid KYC status transition")
	ErrInvalidBranch        = errors.New("invalid branch data")
	ErrInvalidUser          = errors.New("invalid user data")
	ErrWeakPassword         = errors.New("password must be at least 8 characters with 1 uppercase and 1 number")
	ErrInvalidRole          = errors.New("invalid role")
	ErrKeyAlreadyRevoked    = errors.New("API key is already revoked")
	ErrMerchantNotFound     = errors.New("merchant not found")
	ErrBranchNotFound       = errors.New("branch not found")
	ErrUserNotFound         = errors.New("user not found")
	ErrAPIKeyNotFound       = errors.New("API key not found")
	ErrDuplicateEmail       = errors.New("email already in use")
)
