package rules

import "fmt"

// ODH-OLM-003: Conversion Webhook Without AllNamespaces Install Mode

type ConversionWebhookAllNamespacesRule struct{}

func (r *ConversionWebhookAllNamespacesRule) ID() string {
	return "ODH-OLM-003"
}

func (r *ConversionWebhookAllNamespacesRule) Name() string {
	return "conversion-webhook-requires-allnamespaces"
}

func (r *ConversionWebhookAllNamespacesRule) Category() Category {
	return CategoryOLMRequirement
}

func (r *ConversionWebhookAllNamespacesRule) Severity() Severity {
	return SeverityError
}

func (r *ConversionWebhookAllNamespacesRule) Description() string {
	return "CSVs featuring a conversion webhook may only support the AllNamespaces install mode. OLM requires this because conversion webhooks need to be accessible cluster-wide."
}

func (r *ConversionWebhookAllNamespacesRule) Fixable() bool {
	return true // Can be auto-fixed by setting AllNamespaces to true
}

func (r *ConversionWebhookAllNamespacesRule) Validate(bundle *Bundle) []Violation {
	var violations []Violation

	if bundle.CSV == nil {
		return violations
	}

	// Check if there are any conversion webhooks
	hasConversionWebhook := false
	for _, webhook := range bundle.CSV.Spec.WebhookDefinitions {
		if webhook.Type == "ConversionWebhook" {
			hasConversionWebhook = true
			break
		}
	}

	if !hasConversionWebhook {
		return violations
	}

	// Check install modes - AllNamespaces must be supported
	allNamespacesSupported := false
	for _, mode := range bundle.CSV.Spec.InstallModes {
		if mode.Type == "AllNamespaces" && mode.Supported {
			allNamespacesSupported = true
			break
		}
	}

	if !allNamespacesSupported {
		violations = append(violations, Violation{
			RuleID:      r.ID(),
			RuleName:    r.Name(),
			Category:    r.Category(),
			Severity:    r.Severity(),
			Message:     "CSV defines conversion webhook but AllNamespaces install mode is not supported",
			File:        bundle.CSV.FilePath,
			Description: fmt.Sprintf("OLM requires AllNamespaces install mode for operators with conversion webhooks. Set installModes[type=AllNamespaces].supported = true"),
			Fixable:     r.Fixable(),
		})
	}

	return violations
}

