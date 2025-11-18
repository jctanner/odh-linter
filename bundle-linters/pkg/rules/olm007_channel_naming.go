package rules

import (
	"fmt"
	"strings"
)

// ODH-OLM-007: Channel Name Without Stability Indicator

type ChannelNamingRule struct{}

func (r *ChannelNamingRule) ID() string {
	return "ODH-OLM-007"
}

func (r *ChannelNamingRule) Name() string {
	return "channel-naming-convention"
}

func (r *ChannelNamingRule) Category() Category {
	return CategoryOLMBestPractice
}

func (r *ChannelNamingRule) Severity() Severity {
	return SeverityWarning
}

func (r *ChannelNamingRule) Description() string {
	return "Channel names should follow recommended conventions using prefixes like 'stable', 'fast', or 'candidate' to indicate the support level and maturity. This provides a consistent user experience across operators."
}

func (r *ChannelNamingRule) Fixable() bool {
	return false
}

func (r *ChannelNamingRule) Validate(bundle *Bundle) []Violation {
	var violations []Violation

	if bundle.Annotations == nil {
		return violations
	}

	recommendedPrefixes := []string{"stable", "fast", "candidate", "preview", "alpha", "beta"}

	for _, channel := range bundle.Annotations.Channels {
		hasRecommendedPrefix := false
		for _, prefix := range recommendedPrefixes {
			if strings.HasPrefix(strings.ToLower(channel), prefix) {
				hasRecommendedPrefix = true
				break
			}
		}

		if !hasRecommendedPrefix {
			violations = append(violations, Violation{
				RuleID:   r.ID(),
				RuleName: r.Name(),
				Category: r.Category(),
				Severity: r.Severity(),
				Message: fmt.Sprintf("Channel '%s' does not follow recommended naming conventions", channel),
				File:    bundle.Annotations.FilePath,
				Description: fmt.Sprintf("Consider using a channel name starting with: %s. This helps users understand the support level and maturity.",
					strings.Join(recommendedPrefixes, ", ")),
				Fixable: r.Fixable(),
			})
		}
	}

	return violations
}

