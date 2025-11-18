package rules

import "fmt"

// ODH-OLM-010: Conversion Webhook CRD with PreserveUnknownFields=true

type ConversionPreserveUnknownFieldsRule struct{}

func (r *ConversionPreserveUnknownFieldsRule) ID() string {
	return "ODH-OLM-010"
}

func (r *ConversionPreserveUnknownFieldsRule) Name() string {
	return "conversion-webhook-preserve-unknown-fields"
}

func (r *ConversionPreserveUnknownFieldsRule) Category() Category {
	return CategoryOLMRequirement
}

func (r *ConversionPreserveUnknownFieldsRule) Severity() Severity {
	return SeverityError
}

func (r *ConversionPreserveUnknownFieldsRule) Description() string {
	return "CRDs targeted by conversion webhooks must have spec.preserveUnknownFields set to false or nil. This is required for proper conversion webhook functionality."
}

func (r *ConversionPreserveUnknownFieldsRule) Fixable() bool {
	return true // Can be auto-fixed by setting to false
}

func (r *ConversionPreserveUnknownFieldsRule) Validate(bundle *Bundle) []Violation {
	var violations []Violation

	if bundle.CSV == nil {
		return violations
	}

	// Collect CRDs mentioned in conversion webhooks
	conversionCRDs := make(map[string]bool)
	for _, webhook := range bundle.CSV.Spec.WebhookDefinitions {
		if webhook.Type == "ConversionWebhook" {
			for _, crdName := range webhook.ConversionCRDs {
				conversionCRDs[crdName] = true
			}
		}
	}

	if len(conversionCRDs) == 0 {
		return violations
	}

	// Check each CRD
	for _, crd := range bundle.CRDs {
		crdFullName := fmt.Sprintf("%s.%s", crd.Spec.Names.Plural, crd.Spec.Group)
		
		if !conversionCRDs[crdFullName] {
			continue
		}

		// Check PreserveUnknownFields
		if crd.Spec.PreserveUnknownFields != nil && *crd.Spec.PreserveUnknownFields {
			violations = append(violations, Violation{
				RuleID:   r.ID(),
				RuleName: r.Name(),
				Category: r.Category(),
				Severity: r.Severity(),
				Message: fmt.Sprintf("CRD '%s' is targeted by conversion webhook but has preserveUnknownFields=true",
					crdFullName),
				File: crd.FilePath,
				Description: "CRDs used with conversion webhooks must have spec.preserveUnknownFields set to false or nil. Set it to false.",
				Fixable: r.Fixable(),
			})
		}
	}

	return violations
}

