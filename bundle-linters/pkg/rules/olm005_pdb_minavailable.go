package rules

import (
	"fmt"
	"strings"
)

// ODH-OLM-005: PodDisruptionBudget with minAvailable=100%

type PDBMinAvailableRule struct{}

func (r *PDBMinAvailableRule) ID() string {
	return "ODH-OLM-005"
}

func (r *PDBMinAvailableRule) Name() string {
	return "pdb-minavailable-hundred-percent"
}

func (r *PDBMinAvailableRule) Category() Category {
	return CategoryUpgrade
}

func (r *PDBMinAvailableRule) Severity() Severity {
	return SeverityError
}

func (r *PDBMinAvailableRule) Description() string {
	return "PodDisruptionBudget minAvailable field cannot be set to 100%. This can make a node impossible to drain and block important lifecycle actions like operator upgrades or even cluster upgrades."
}

func (r *PDBMinAvailableRule) Fixable() bool {
	return false
}

func (r *PDBMinAvailableRule) Validate(bundle *Bundle) []Violation {
	var violations []Violation

	for _, resource := range bundle.OtherResources {
		if resource.Kind != "PodDisruptionBudget" {
			continue
		}

		// Check minAvailable field in spec
		if minAvail, ok := resource.Spec["minAvailable"]; ok {
			if isHundredPercent(minAvail) {
				violations = append(violations, Violation{
					RuleID:      r.ID(),
					RuleName:    r.Name(),
					Category:    r.Category(),
					Severity:    r.Severity(),
					Message:     fmt.Sprintf("PodDisruptionBudget '%s' has minAvailable set to 100%%", resource.Metadata.Name),
					File:        resource.FilePath,
					Description: "Setting minAvailable to 100% prevents node drains and can block cluster lifecycle operations. Use a lower percentage.",
					Fixable:     r.Fixable(),
				})
			}
		}
	}

	return violations
}

// isHundredPercent checks if a value is "100%"
func isHundredPercent(val interface{}) bool {
	switch v := val.(type) {
	case string:
		trimmed := strings.TrimSpace(v)
		return trimmed == "100%" || trimmed == "100"
	case int:
		return v == 100
	case int64:
		return v == 100
	case float64:
		return v == 100.0
	}
	return false
}

