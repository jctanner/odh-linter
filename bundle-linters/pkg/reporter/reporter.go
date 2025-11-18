package reporter

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/opendatahub-io/odh-linter/bundle-linters/pkg/rules"
)

// Reporter formats and outputs validation results
type Reporter struct {
	writer io.Writer
}

// New creates a new Reporter
func New(writer io.Writer) *Reporter {
	return &Reporter{writer: writer}
}

// Report outputs validation violations
func (r *Reporter) Report(violations []rules.Violation) error {
	if len(violations) == 0 {
		_, err := fmt.Fprintln(r.writer, "✓ No issues found")
		return err
	}

	// Sort violations by severity, then by file, then by rule ID
	sort.Slice(violations, func(i, j int) bool {
		if violations[i].Severity != violations[j].Severity {
			return severityWeight(violations[i].Severity) > severityWeight(violations[j].Severity)
		}
		if violations[i].File != violations[j].File {
			return violations[i].File < violations[j].File
		}
		return violations[i].RuleID < violations[j].RuleID
	})

	// Count by severity
	errorCount := 0
	warningCount := 0
	infoCount := 0

	for _, v := range violations {
		switch v.Severity {
		case rules.SeverityError:
			errorCount++
		case rules.SeverityWarning:
			warningCount++
		case rules.SeverityInfo:
			infoCount++
		}
	}

	// Print summary header
	fmt.Fprintf(r.writer, "\nFound %d issue(s):\n", len(violations))
	if errorCount > 0 {
		fmt.Fprintf(r.writer, "  - %d error(s)\n", errorCount)
	}
	if warningCount > 0 {
		fmt.Fprintf(r.writer, "  - %d warning(s)\n", warningCount)
	}
	if infoCount > 0 {
		fmt.Fprintf(r.writer, "  - %d info\n", infoCount)
	}
	fmt.Fprintln(r.writer, "")

	// Print violations
	for _, v := range violations {
		fmt.Fprintln(r.writer, r.formatViolation(v))
		fmt.Fprintln(r.writer, "")
	}

	return nil
}

// formatViolation formats a single violation for display
func (r *Reporter) formatViolation(v rules.Violation) string {
	var sb strings.Builder

	// Format header with severity emoji
	severityIcon := getSeverityIcon(v.Severity)
	fmt.Fprintf(&sb, "%s [%s] %s\n", severityIcon, v.RuleID, v.Message)

	// Add file location
	if v.File != "" {
		if v.Line > 0 {
			fmt.Fprintf(&sb, "   File: %s:%d\n", v.File, v.Line)
		} else {
			fmt.Fprintf(&sb, "   File: %s\n", v.File)
		}
	}

	// Add category
	fmt.Fprintf(&sb, "   Category: %s\n", v.Category)

	// Add description if available
	if v.Description != "" {
		fmt.Fprintf(&sb, "   %s\n", v.Description)
	}

	// Add fixable status
	if v.Fixable {
		fmt.Fprintf(&sb, "   ℹ️  This issue is potentially auto-fixable\n")
	}

	return sb.String()
}

// getSeverityIcon returns an emoji icon for the severity level
func getSeverityIcon(severity rules.Severity) string {
	switch severity {
	case rules.SeverityError:
		return "❌"
	case rules.SeverityWarning:
		return "⚠️ "
	case rules.SeverityInfo:
		return "ℹ️ "
	default:
		return "  "
	}
}

// severityWeight returns a numeric weight for sorting
func severityWeight(severity rules.Severity) int {
	switch severity {
	case rules.SeverityError:
		return 3
	case rules.SeverityWarning:
		return 2
	case rules.SeverityInfo:
		return 1
	default:
		return 0
	}
}

// ReportSummary outputs a summary of violations
func (r *Reporter) ReportSummary(violations []rules.Violation) error {
	errorCount := 0
	warningCount := 0

	for _, v := range violations {
		switch v.Severity {
		case rules.SeverityError:
			errorCount++
		case rules.SeverityWarning:
			warningCount++
		}
	}

	if errorCount > 0 {
		fmt.Fprintf(r.writer, "\n❌ Validation failed: %d error(s), %d warning(s)\n", errorCount, warningCount)
		return fmt.Errorf("validation failed with %d error(s)", errorCount)
	}

	if warningCount > 0 {
		fmt.Fprintf(r.writer, "\n⚠️  Validation passed with %d warning(s)\n", warningCount)
	} else {
		fmt.Fprintln(r.writer, "\n✓ All checks passed!")
	}

	return nil
}

