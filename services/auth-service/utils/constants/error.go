package constants

const (
	ErrInvalidCredentials = "invalid email or password"
	ErrEmailExists        = "email already registered"
	ErrUserNotFound       = "user not found"
	ErrUserDisabled       = "user is disabled"
	ErrInvalidToken       = "invalid token"
	ErrSessionNotFound    = "session not found"
	ErrUnauthorized       = "unauthorized"
	ErrForbidden          = "forbidden"
	ErrInvalidRole        = "invalid role"
	ErrPolicyExists       = "policy already exists"
	ErrPolicyNotFound     = "policy not found"
	ErrInvalidPolicy      = "invalid service, role, resource, or action combination"
	ErrInvalidUserID      = "invalid user id"
	ErrLogoutFailed       = "logout failed"
	ErrInternalServer     = "internal server error"
	ErrFailedListPolicies = "failed to list policies"
	ErrFailedGetResources = "failed to get resources"
	ErrFailedAddPolicy    = "failed to add policy"
	ErrFailedDeletePolicy = "failed to delete policy"
)
