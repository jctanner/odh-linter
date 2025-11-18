package rules

import (
	"fmt"
	"strings"
)

// ODH-OLM-002: Webhook Rule Intercepting Operator Resources

type WebhookOperatorResourcesRule struct{}

func (r *WebhookOperatorResourcesRule) ID() string {
	return "ODH-OLM-002"
}

func (r *WebhookOperatorResourcesRule) Name() string {
	return "webhook-intercepts-operator-resources"
}

func (r *WebhookOperatorResourcesRule) Category() Category {
	return CategoryOLMRequirement
}

func (r *WebhookOperatorResourcesRule) Severity() Severity {
	return SeverityError
}

func (r *WebhookOperatorResourcesRule) Description() string {
	return "OLM will place the CSV in failed phase if webhook rules intercept: (1) all groups (apiGroups: ['*']), (2) the operators.coreos.com group, or (3) ValidatingWebhookConfigurations or MutatingWebhookConfigurations resources. This prevents operators from breaking OLM's ability to manage other operators."
}

func (r *WebhookOperatorResourcesRule) Fixable() bool {
	return false
}

func (r *WebhookOperatorResourcesRule) Validate(bundle *Bundle) []Violation {
	var violations []Violation

	if bundle.CSV == nil {
		return violations
	}

	for _, webhook := range bundle.CSV.Spec.WebhookDefinitions {
		// Skip conversion webhooks (they have different rules)
		if webhook.Type == "ConversionWebhook" {
			continue
		}

		for _, rule := range webhook.Rules {
			// Check for intercepting all groups
			if containsWildcard(rule.APIGroups) {
				violations = append(violations, Violation{
					RuleID:   r.ID(),
					RuleName: r.Name(),
					Category: r.Category(),
					Severity: r.Severity(),
					Message: fmt.Sprintf("Webhook '%s' intercepts all API groups (apiGroups: ['*']). OLM will fail the CSV.",
						webhook.GenerateName),
					File:        bundle.CSV.FilePath,
					Description: "Webhooks cannot intercept all API groups. This would prevent OLM from managing other operators.",
					Fixable:     r.Fixable(),
				})
			}

			// Check for intercepting operators.coreos.com group
			if containsOperatorGroup(rule.APIGroups) {
				violations = append(violations, Violation{
					RuleID:   r.ID(),
					RuleName: r.Name(),
					Category: r.Category(),
					Severity: r.Severity(),
					Message: fmt.Sprintf("Webhook '%s' intercepts the 'operators.coreos.com' API group. OLM will fail the CSV.",
						webhook.GenerateName),
					File:        bundle.CSV.FilePath,
					Description: "Webhooks cannot intercept the operators.coreos.com group. This would break OLM's ability to manage operators.",
					Fixable:     r.Fixable(),
				})
			}

			// Check for intercepting webhook configuration resources
			if containsWebhookConfigResources(rule.Resources) {
				violations = append(violations, Violation{
					RuleID:   r.ID(),
					RuleName: r.Name(),
					Category: r.Category(),
					Severity: r.Severity(),
					Message: fmt.Sprintf("Webhook '%s' intercepts ValidatingWebhookConfigurations or MutatingWebhookConfigurations resources. OLM will fail the CSV.",
						webhook.GenerateName),
					File:        bundle.CSV.FilePath,
					Description: "Webhooks cannot intercept webhook configuration resources. This would break OLM's ability to configure webhooks.",
					Fixable:     r.Fixable(),
				})
			}
		}
	}

	return violations
}

func containsWildcard(groups []string) bool {
	for _, g := range groups {
		if g == "*" {
			return true
		}
	}
	return false
}

func containsOperatorGroup(groups []string) bool {
	for _, g := range groups {
		if g == "operators.coreos.com" {
			return true
		}
	}
	return false
}

func containsWebhookConfigResources(resources []string) bool {
	for _, r := range resources {
		rLower := strings.ToLower(r)
		if rLower == "validatingwebhookconfigurations" || rLower == "mutatingwebhookconfigurations" {
			return true
		}
	}
	return false
}

