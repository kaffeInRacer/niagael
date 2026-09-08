package IUseCase

import (
	"context"

	"kaffein/auth-service/internal/repository"
)

type RBACUseCase interface {
	GetAllPolicies(ctx context.Context) ([]repository.Policy, error)
	AddPolicy(ctx context.Context, service, role, resource, action string) error
	DeletePolicy(ctx context.Context, service, role, resource, action string) error
	DeleteAllPoliciesForRole(ctx context.Context, service, role string) error
	GetResources(ctx context.Context) (map[string][]string, error)
}
