package rules

import (
	"fmt"
	"strings"
)

// ODH-OLM-004: PodDisruptionBudget with maxUnavailable=0

type PDBMaxUnavailableRule struct{}

func (r *PDBMaxUnavailableRule) ID() string {
	return "ODH-OLM-004"
}

func (r *PDBMaxUnavailableRule) Name() string {
	return "pdb-maxunavailable-zero"
}

func (r *PDBMaxUnavailableRule) Category() Category {
	return CategoryUpgrade
}

func (r *PDBMaxUnavailableRule) Severity() Severity {
	return SeverityError
}

func (r *PDBMaxUnavailableRule) Description() string {
	return "PodDisruptionBudget maxUnavailable field cannot be set to 0 or 0%. This can make a node impossible to drain and block important lifecycle actions like operator upgrades or even cluster upgrades."
}

func (r *PDBMaxUnavailableRule) Fixable() bool {
	return false
}

func (r *PDBMaxUnavailableRule) Validate(bundle *Bundle) []Violation {
	var violations []Violation

	for _, resource := range bundle.OtherResources {
		if resource.Kind != "PodDisruptionBudget" {
			continue
		}

		// Check maxUnavailable field in spec
		if maxUnav, ok := resource.Spec["maxUnavailable"]; ok {
			if isZeroValue(maxUnav) {
				violations = append(violations, Violation{
					RuleID:      r.ID(),
					RuleName:    r.Name(),
					Category:    r.Category(),
					Severity:    r.Severity(),
					Message:     fmt.Sprintf("PodDisruptionBudget '%s' has maxUnavailable set to 0 or 0%%", resource.Metadata.Name),
					File:        resource.FilePath,
					Description: "Setting maxUnavailable to 0 or 0% prevents node drains and can block cluster lifecycle operations. Use a value >= 1.",
					Fixable:     r.Fixable(),
				})
			}
		}
	}

	return violations
}

// isZeroValue checks if a value is 0, "0", or "0%"
func isZeroValue(val interface{}) bool {
	switch v := val.(type) {
	case int:
		return v == 0
	case int64:
		return v == 0
	case float64:
		return v == 0
	case string:
		trimmed := strings.TrimSpace(v)
		return trimmed == "0" || trimmed == "0%"
	}
	return false
}

