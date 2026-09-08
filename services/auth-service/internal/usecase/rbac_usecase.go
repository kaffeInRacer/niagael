package usecase

import (
	"context"

	"kaffein/auth-service/internal/repository"
)

type rbacUseCase struct {
	repo *repository.RBACRepository
}

func NewRBACUseCase(repo *repository.RBACRepository) *rbacUseCase {
	return &rbacUseCase{repo: repo}
}

func (u *rbacUseCase) GetAllPolicies(ctx context.Context) ([]repository.Policy, error) {
	return u.repo.GetAllPolicies(ctx)
}

func (u *rbacUseCase) AddPolicy(ctx context.Context, service, role, resource, action string) error {
	return u.repo.AddPolicy(ctx, service, role, resource, action)
}

func (u *rbacUseCase) DeletePolicy(ctx context.Context, service, role, resource, action string) error {
	return u.repo.DeletePolicy(ctx, service, role, resource, action)
}

func (u *rbacUseCase) DeleteAllPoliciesForRole(ctx context.Context, service, role string) error {
	return u.repo.DeleteAllPoliciesForRole(ctx, service, role)
}

func (u *rbacUseCase) GetResources(ctx context.Context) (map[string][]string, error) {
	return u.repo.GetResources(ctx)
}
