package usecase

import (
	"context"
	"errors"
	"sort"
	"strings"

	"kaffein/auth-service/internal/domain"
	"kaffein/auth-service/internal/repository"
	"kaffein/auth-service/utils/constants"
)

var allowedPolicies = map[string]map[string]map[string]bool{
	"auth": {
		"users":    {"read": true, "update": true},
		"policies": {"read": true, "create": true, "delete": true},
	},
	"product": {
		"categories":     {"read": true, "create": true, "update": true, "delete": true},
		"products":       {"read": true, "create": true, "update": true, "delete": true},
		"variants":       {"read": true, "create": true, "update": true, "delete": true},
		"product-images": {"read": true, "create": true, "delete": true},
	},
	"order": {
		"carts":     {"read": true, "create": true, "update": true, "delete": true},
		"addresses": {"read": true, "create": true, "update": true, "delete": true},
		"orders":    {"read": true, "create": true, "update": true},
		"payments":  {"read": true, "create": true},
	},
	"dynamic-pricing": {
		"promos":      {"read": true, "create": true, "update": true, "delete": true, "apply": true},
		"flash-sales": {"read": true, "create": true, "update": true, "delete": true},
	},
}

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
	service, role, resource, action = normalizePolicy(service, role, resource, action)
	if !validPolicy(service, role, resource, action) {
		return errors.New(constants.ErrInvalidPolicy)
	}
	return u.repo.AddPolicy(ctx, service, role, resource, action)
}

func (u *rbacUseCase) DeletePolicy(ctx context.Context, service, role, resource, action string) error {
	service, role, resource, action = normalizePolicy(service, role, resource, action)
	if !validPolicy(service, role, resource, action) {
		return errors.New(constants.ErrInvalidPolicy)
	}
	return u.repo.DeletePolicy(ctx, service, role, resource, action)
}

func (u *rbacUseCase) DeleteAllPoliciesForRole(ctx context.Context, service, role string) error {
	service, role, _, _ = normalizePolicy(service, role, "", "")
	if _, ok := allowedPolicies[service]; !ok || !domain.ValidRole(role) {
		return errors.New(constants.ErrInvalidPolicy)
	}
	return u.repo.DeleteAllPoliciesForRole(ctx, service, role)
}

func (u *rbacUseCase) GetResources(ctx context.Context) (map[string][]string, error) {
	resources := make(map[string][]string, len(allowedPolicies))
	for service, serviceResources := range allowedPolicies {
		for resource := range serviceResources {
			resources[service] = append(resources[service], resource)
		}
		sort.Strings(resources[service])
	}
	return resources, nil
}

func normalizePolicy(service, role, resource, action string) (string, string, string, string) {
	return strings.ToLower(strings.TrimSpace(service)), strings.ToLower(strings.TrimSpace(role)), strings.ToLower(strings.TrimSpace(resource)), strings.ToLower(strings.TrimSpace(action))
}

func validPolicy(service, role, resource, action string) bool {
	if !domain.ValidRole(role) {
		return false
	}
	resources, ok := allowedPolicies[service]
	if !ok {
		return false
	}
	actions, ok := resources[resource]
	return ok && actions[action]
}
