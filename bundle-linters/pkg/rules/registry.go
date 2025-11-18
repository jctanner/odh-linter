package rules

// GetAllRules returns all available validation rules
func GetAllRules() []Rule {
	return []Rule{
		&MinKubeVersionRule{},
		&WebhookOperatorResourcesRule{},
		&ConversionWebhookAllNamespacesRule{},
		&PDBMaxUnavailableRule{},
		&PDBMinAvailableRule{},
		&PriorityClassGlobalDefaultRule{},
		&ChannelNamingRule{},
		&ConversionPreserveUnknownFieldsRule{},
	}
}

// GetRuleByID returns a rule by its ID
func GetRuleByID(id string) Rule {
	for _, rule := range GetAllRules() {
		if rule.ID() == id {
			return rule
		}
	}
	return nil
}

// ValidateBundle runs all rules against a bundle and returns violations
func ValidateBundle(bundle *Bundle, rules []Rule) []Violation {
	var allViolations []Violation

	for _, rule := range rules {
		violations := rule.Validate(bundle)
		allViolations = append(allViolations, violations...)
	}

	return allViolations
}

