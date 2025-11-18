package rules

import "fmt"

// ODH-OLM-006: PriorityClass with globalDefault=true

type PriorityClassGlobalDefaultRule struct{}

func (r *PriorityClassGlobalDefaultRule) ID() string {
	return "ODH-OLM-006"
}

func (r *PriorityClassGlobalDefaultRule) Name() string {
	return "priorityclass-globaldefault-true"
}

func (r *PriorityClassGlobalDefaultRule) Category() Category {
	return CategorySecurity
}

func (r *PriorityClassGlobalDefaultRule) Severity() Severity {
	return SeverityError
}

func (r *PriorityClassGlobalDefaultRule) Description() string {
	return "PriorityClass globalDefault should always be false in operator bundles. Setting globalDefault means all pods in the cluster without an explicit priority class will use this default, which can unintentionally affect other workloads."
}

func (r *PriorityClassGlobalDefaultRule) Fixable() bool {
	return true // Can be auto-fixed by setting to false
}

func (r *PriorityClassGlobalDefaultRule) Validate(bundle *Bundle) []Violation {
	var violations []Violation

	for _, resource := range bundle.OtherResources {
		if resource.Kind != "PriorityClass" {
			continue
		}

		// Check globalDefault field
		if globalDefault, ok := resource.Spec["globalDefault"]; ok {
			if isTrueValue(globalDefault) {
				violations = append(violations, Violation{
					RuleID:      r.ID(),
					RuleName:    r.Name(),
					Category:    r.Category(),
					Severity:    r.Severity(),
					Message:     fmt.Sprintf("PriorityClass '%s' has globalDefault set to true", resource.Metadata.Name),
					File:        resource.FilePath,
					Description: "PriorityClass globalDefault should be false in operator bundles. Setting it to true affects all pods cluster-wide.",
					Fixable:     r.Fixable(),
				})
			}
		}
	}

	return violations
}

// isTrueValue checks if a value is true
func isTrueValue(val interface{}) bool {
	switch v := val.(type) {
	case bool:
		return v
	case string:
		return v == "true" || v == "True" || v == "TRUE"
	}
	return false
}

