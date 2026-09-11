package auth

import (
	_ "embed"
	"fmt"
	"os"
	"strings"
)

const policyFileName = "internal/auth/casbin_rule.conf"

//go:embed casbin_rule.conf
var embeddedPolicies string

func policyFilePath() string {
	return policyFileName
}

func loadPolicyFromDisk() (string, error) {
	data, err := os.ReadFile(policyFilePath())
	if err != nil {
		if !os.IsNotExist(err) {
			return "", fmt.Errorf("read policy file: %w", err)
		}
		if err := os.WriteFile(policyFilePath(), []byte(embeddedPolicies), 0o644); err != nil {
			return "", fmt.Errorf("create default policy file: %w", err)
		}
		return embeddedPolicies, nil
	}
	content := strings.TrimSpace(string(data))
	if content == "" {
		return embeddedPolicies, nil
	}
	return content, nil
}

func writePolicyFile(lines []string) error {
	content := strings.Join(lines, "\n") + "\n"
	tmp := policyFileName + ".tmp"
	if err := os.WriteFile(tmp, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write policy temp file: %w", err)
	}
	return os.Rename(tmp, policyFilePath())
}

type PolicyLine struct {
	Role     string `json:"role"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

type PolicyEvent struct {
	Service  string       `json:"service"`
	Policies []PolicyLine `json:"policies"`
}

func (e PolicyEvent) PolicyFileLines() []string {
	lines := make([]string, 0, len(e.Policies))
	for _, p := range e.Policies {
		lines = append(lines, p.Role+", "+p.Resource+", "+p.Action)
	}
	return lines
}
