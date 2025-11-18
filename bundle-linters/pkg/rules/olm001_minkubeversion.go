package rules

// ODH-OLM-001: Missing minKubeVersion in CSV

type MinKubeVersionRule struct{}

func (r *MinKubeVersionRule) ID() string {
	return "ODH-OLM-001"
}

func (r *MinKubeVersionRule) Name() string {
	return "missing-minkubeversion"
}

func (r *MinKubeVersionRule) Category() Category {
	return CategoryOLMBestPractice
}

func (r *MinKubeVersionRule) Severity() Severity {
	return SeverityWarning
}

func (r *MinKubeVersionRule) Description() string {
	return "ClusterServiceVersion should specify spec.minKubeVersion to indicate the minimum Kubernetes version supported by the operator. Without this, the operator may be installed on incompatible clusters."
}

func (r *MinKubeVersionRule) Fixable() bool {
	return false // Requires user to determine minimum version
}

func (r *MinKubeVersionRule) Validate(bundle *Bundle) []Violation {
	var violations []Violation

	if bundle.CSV == nil {
		return violations
	}

	if bundle.CSV.Spec.MinKubeVersion == "" {
		violations = append(violations, Violation{
			RuleID:      r.ID(),
			RuleName:    r.Name(),
			Category:    r.Category(),
			Severity:    r.Severity(),
			Message:     "ClusterServiceVersion is missing spec.minKubeVersion field",
			File:        bundle.CSV.FilePath,
			Description: "It is recommended to specify the minimum Kubernetes version your operator supports. This prevents installation on incompatible clusters.",
			Fixable:     r.Fixable(),
		})
	}

	return violations
}

