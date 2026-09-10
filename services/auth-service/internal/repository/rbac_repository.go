package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"kaffein/auth-service/utils/constants"

	"kaffein/auth-service/pkg/kafka/event"
)

var ()

type Policy struct {
	ID       int64  `json:"id"`
	Service  string `json:"service"`
	Role     string `json:"role"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

type RBACRepository struct {
	db           *pgxpool.Pool
	casbinTopic  string
	kafkaBrokers []string
}

func NewRBACRepository(db *pgxpool.Pool, casbinTopic string, kafkaBrokers []string) *RBACRepository {
	return &RBACRepository{db: db, casbinTopic: casbinTopic, kafkaBrokers: kafkaBrokers}
}

func (r *RBACRepository) GetAllPolicies(ctx context.Context) ([]Policy, error) {
	rows, err := r.db.Query(ctx, `SELECT id, service, v0, v1, v2 FROM casbin_rule WHERE ptype = 'p' ORDER BY service, v0, v1, v2`)
	if err != nil {
		return nil, fmt.Errorf("get all policies: %w", err)
	}
	defer rows.Close()

	var policies []Policy
	for rows.Next() {
		var p Policy
		if err := rows.Scan(&p.ID, &p.Service, &p.Role, &p.Resource, &p.Action); err != nil {
			return nil, fmt.Errorf("scan policy: %w", err)
		}
		policies = append(policies, p)
	}

	return policies, rows.Err()
}

func (r *RBACRepository) AddPolicy(ctx context.Context, service, role, resource, action string) error {
	result, err := r.db.Exec(ctx, `INSERT INTO casbin_rule (ptype, v0, v1, v2, v3, v4, v5, service)
		VALUES ('p', $1, $2, $3, '', '', '', $4)
		ON CONFLICT (ptype, v0, v1, v2, v3, v4, v5, service) DO NOTHING`, role, resource, action, service)
	if err != nil {
		return fmt.Errorf("insert policy: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New(constants.ErrPolicyExists)
	}

	r.publishPoliciesForService(ctx, service)

	return nil
}

func (r *RBACRepository) DeletePolicy(ctx context.Context, service, role, resource, action string) error {
	result, err := r.db.Exec(ctx, `DELETE FROM casbin_rule WHERE ptype = 'p' AND v0 = $1 AND v1 = $2 AND v2 = $3 AND service = $4`, role, resource, action, service)
	if err != nil {
		return fmt.Errorf("delete policy: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New(constants.ErrPolicyNotFound)
	}

	r.publishPoliciesForService(ctx, service)

	return nil
}

func (r *RBACRepository) DeleteAllPoliciesForRole(ctx context.Context, service, role string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM casbin_rule WHERE ptype = 'p' AND v0 = $1 AND service = $2`, role, service)
	if err != nil {
		return fmt.Errorf("delete policies for role: %w", err)
	}

	r.publishPoliciesForService(ctx, service)

	return nil
}

func (r *RBACRepository) GetResources(ctx context.Context) (map[string][]string, error) {
	rows, err := r.db.Query(ctx, `SELECT DISTINCT service, v1 FROM casbin_rule WHERE ptype = 'p' ORDER BY service, v1`)
	if err != nil {
		return nil, fmt.Errorf("get resources: %w", err)
	}
	defer rows.Close()

	result := make(map[string][]string)
	for rows.Next() {
		var service, resource string
		if err := rows.Scan(&service, &resource); err != nil {
			return nil, fmt.Errorf("scan resource: %w", err)
		}
		result[service] = append(result[service], resource)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// publishPoliciesForService fetches all policies of a service from the
// database and publishes them so each service can rewrite its local
// casbin_rule.conf file.
func (r *RBACRepository) publishPoliciesForService(ctx context.Context, service string) {
	policies, err := r.GetPoliciesByService(ctx, service)
	if err != nil {
		return
	}
	type policyLine struct {
		Role     string `json:"role"`
		Resource string `json:"resource"`
		Action   string `json:"action"`
	}
	lines := make([]policyLine, 0, len(policies))
	for _, p := range policies {
		lines = append(lines, policyLine{Role: p.Role, Resource: p.Resource, Action: p.Action})
	}
	event.Publish(r.kafkaBrokers, r.casbinTopic, "policy-"+service, map[string]any{
		"service":  service,
		"policies": lines,
	})
}

func (r *RBACRepository) GetPoliciesByService(ctx context.Context, service string) ([]Policy, error) {
	rows, err := r.db.Query(ctx, `SELECT id, service, v0, v1, v2 FROM casbin_rule WHERE ptype = 'p' AND service = $1 ORDER BY v0, v1, v2`, service)
	if err != nil {
		return nil, fmt.Errorf("get policies by service: %w", err)
	}
	defer rows.Close()

	var policies []Policy
	for rows.Next() {
		var p Policy
		if err := rows.Scan(&p.ID, &p.Service, &p.Role, &p.Resource, &p.Action); err != nil {
			return nil, fmt.Errorf("scan policy: %w", err)
		}
		policies = append(policies, p)
	}
	return policies, rows.Err()
}
