package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Policy struct {
	ID       int64  `json:"id"`
	Service  string `json:"service"`
	Role     string `json:"role"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

type RBACRepository struct {
	productDB   *pgxpool.Pool
	pricingDB   *pgxpool.Pool
	orderDB     *pgxpool.Pool
}

func NewRBACRepository(productDB, pricingDB, orderDB *pgxpool.Pool) *RBACRepository {
	return &RBACRepository{
		productDB: productDB,
		pricingDB: pricingDB,
		orderDB:   orderDB,
	}
}

func (r *RBACRepository) GetAllPolicies(ctx context.Context) ([]Policy, error) {
	var allPolicies []Policy

	productPolicies, err := r.getPoliciesFromDB(ctx, r.productDB, "product")
	if err != nil {
		return nil, fmt.Errorf("get product policies: %w", err)
	}
	allPolicies = append(allPolicies, productPolicies...)

	pricingPolicies, err := r.getPoliciesFromDB(ctx, r.pricingDB, "dynamic-pricing")
	if err != nil {
		return nil, fmt.Errorf("get pricing policies: %w", err)
	}
	allPolicies = append(allPolicies, pricingPolicies...)

	orderPolicies, err := r.getPoliciesFromDB(ctx, r.orderDB, "order")
	if err != nil {
		return nil, fmt.Errorf("get order policies: %w", err)
	}
	allPolicies = append(allPolicies, orderPolicies...)

	return allPolicies, nil
}

func (r *RBACRepository) getPoliciesFromDB(ctx context.Context, db *pgxpool.Pool, serviceName string) ([]Policy, error) {
	if db == nil {
		return nil, nil
	}

	rows, err := db.Query(ctx, `SELECT id, v0, v1, v2 FROM casbin_rule WHERE ptype = 'p' ORDER BY v0, v1, v2`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []Policy
	for rows.Next() {
		var p Policy
		if err := rows.Scan(&p.ID, &p.Role, &p.Resource, &p.Action); err != nil {
			return nil, err
		}
		p.Service = serviceName
		policies = append(policies, p)
	}
	return policies, rows.Err()
}

func (r *RBACRepository) AddPolicy(ctx context.Context, service, role, resource, action string) error {
	db := r.getDBForService(service)
	if db == nil {
		return fmt.Errorf("unknown service: %s", service)
	}

	var exists bool
	err := db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM casbin_rule WHERE ptype = 'p' AND v0 = $1 AND v1 = $2 AND v2 = $3)`, role, resource, action).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check existing policy: %w", err)
	}
	if exists {
		return fmt.Errorf("policy already exists")
	}

	_, err = db.Exec(ctx, `INSERT INTO casbin_rule (ptype, v0, v1, v2, v3, v4, v5) VALUES ('p', $1, $2, $3, '', '', '')`, role, resource, action)
	if err != nil {
		return fmt.Errorf("insert policy: %w", err)
	}

	return nil
}

func (r *RBACRepository) DeletePolicy(ctx context.Context, service, role, resource, action string) error {
	db := r.getDBForService(service)
	if db == nil {
		return fmt.Errorf("unknown service: %s", service)
	}

	result, err := db.Exec(ctx, `DELETE FROM casbin_rule WHERE ptype = 'p' AND v0 = $1 AND v1 = $2 AND v2 = $3`, role, resource, action)
	if err != nil {
		return fmt.Errorf("delete policy: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("policy not found")
	}

	return nil
}

func (r *RBACRepository) DeleteAllPoliciesForRole(ctx context.Context, service, role string) error {
	db := r.getDBForService(service)
	if db == nil {
		return fmt.Errorf("unknown service: %s", service)
	}

	_, err := db.Exec(ctx, `DELETE FROM casbin_rule WHERE ptype = 'p' AND v0 = $1`, role)
	if err != nil {
		return fmt.Errorf("delete policies for role: %w", err)
	}

	return nil
}

func (r *RBACRepository) getDBForService(service string) *pgxpool.Pool {
	switch service {
	case "product":
		return r.productDB
	case "dynamic-pricing":
		return r.pricingDB
	case "order":
		return r.orderDB
	default:
		return nil
	}
}

func (r *RBACRepository) GetResources(ctx context.Context) (map[string][]string, error) {
	services := map[string]*pgxpool.Pool{
		"product":        r.productDB,
		"dynamic-pricing": r.pricingDB,
		"order":          r.orderDB,
	}

	result := make(map[string][]string)
	for serviceName, db := range services {
		if db == nil {
			continue
		}
		rows, err := db.Query(ctx, `SELECT DISTINCT v1 FROM casbin_rule WHERE ptype = 'p' ORDER BY v1`)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var resources []string
		for rows.Next() {
			var resource string
			if err := rows.Scan(&resource); err != nil {
				return nil, err
			}
			resources = append(resources, resource)
		}
		result[serviceName] = resources
	}
	return result, nil
}
