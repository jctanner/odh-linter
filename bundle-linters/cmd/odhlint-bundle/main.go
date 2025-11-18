package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/opendatahub-io/odh-linter/bundle-linters/pkg/loader"
	"github.com/opendatahub-io/odh-linter/bundle-linters/pkg/reporter"
	"github.com/opendatahub-io/odh-linter/bundle-linters/pkg/rules"
)

const version = "1.0.0"

func main() {
	// Command line flags
	listRules := flag.Bool("list-rules", false, "List all available rules")
	enableRules := flag.String("enable", "", "Comma-separated list of rule IDs to enable (default: all)")
	disableRules := flag.String("disable", "", "Comma-separated list of rule IDs to disable")
	showVersion := flag.Bool("version", false, "Show version information")
	noWarnings := flag.Bool("no-warnings", false, "Treat warnings as passing (exit 0)")
	
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <bundle-path>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "odhlint-bundle validates Operator Lifecycle Manager (OLM) bundles against best practices and requirements.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s ./bundle/\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --list-rules\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --enable ODH-OLM-001,ODH-OLM-002 ./bundle/\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --disable ODH-OLM-007 ./bundle/\n", os.Args[0])
	}

	flag.Parse()

	// Handle --version
	if *showVersion {
		fmt.Printf("odhlint-bundle version %s\n", version)
		os.Exit(0)
	}

	// Handle --list-rules
	if *listRules {
		printRules()
		os.Exit(0)
	}

	// Validate arguments
	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Error: bundle path is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	bundlePath := flag.Arg(0)

	// Load the bundle
	fmt.Printf("Loading bundle from: %s\n", bundlePath)
	bundle, err := loader.LoadBundle(bundlePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading bundle: %v\n", err)
		os.Exit(1)
	}

	// Determine which rules to run
	rulesToRun := selectRules(*enableRules, *disableRules)
	fmt.Printf("Running %d validation rule(s)...\n\n", len(rulesToRun))

	// Validate the bundle
	violations := rules.ValidateBundle(bundle, rulesToRun)

	// Report results
	rep := reporter.New(os.Stdout)
	if err := rep.Report(violations); err != nil {
		fmt.Fprintf(os.Stderr, "Error reporting results: %v\n", err)
		os.Exit(1)
	}

	// Exit with appropriate code
	exitCode := 0
	if hasErrors(violations) {
		exitCode = 1
	} else if !*noWarnings && hasWarnings(violations) {
		exitCode = 0 // Warnings don't cause failure by default
	}

	if err := rep.ReportSummary(violations); err != nil {
		if exitCode == 0 {
			exitCode = 1
		}
	}

	os.Exit(exitCode)
}

// printRules prints all available rules
func printRules() {
	allRules := rules.GetAllRules()
	
	fmt.Println("Available validation rules:")
	fmt.Println()

	// Group by category
	categories := make(map[rules.Category][]rules.Rule)
	for _, rule := range allRules {
		cat := rule.Category()
		categories[cat] = append(categories[cat], rule)
	}

	// Print by category
	for _, cat := range []rules.Category{
		rules.CategoryOLMRequirement,
		rules.CategoryOLMBestPractice,
		rules.CategorySecurity,
		rules.CategoryUpgrade,
	} {
		if ruleList, ok := categories[cat]; ok && len(ruleList) > 0 {
			fmt.Printf("=== %s ===\n\n", cat)
			for _, rule := range ruleList {
				fmt.Printf("  %s: %s\n", rule.ID(), rule.Name())
				fmt.Printf("    Severity: %s\n", rule.Severity())
				fmt.Printf("    %s\n", rule.Description())
				fmt.Println()
			}
		}
	}

	fmt.Printf("Total: %d rules\n", len(allRules))
}

// selectRules determines which rules to run based on enable/disable flags
func selectRules(enable, disable string) []rules.Rule {
	allRules := rules.GetAllRules()

	// If enable is specified, start with empty set
	if enable != "" {
		enabledIDs := parseRuleList(enable)
		var selected []rules.Rule
		for _, rule := range allRules {
			if _, ok := enabledIDs[rule.ID()]; ok {
				selected = append(selected, rule)
			}
		}
		return selected
	}

	// Otherwise start with all rules
	selected := allRules

	// Remove disabled rules
	if disable != "" {
		disabledIDs := parseRuleList(disable)
		var filtered []rules.Rule
		for _, rule := range selected {
			if _, ok := disabledIDs[rule.ID()]; !ok {
				filtered = append(filtered, rule)
			}
		}
		selected = filtered
	}

	return selected
}

// parseRuleList parses a comma-separated list of rule IDs
func parseRuleList(list string) map[string]bool {
	result := make(map[string]bool)
	if list == "" {
		return result
	}

	parts := strings.Split(list, ",")
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result[trimmed] = true
		}
	}

	return result
}

// hasErrors checks if there are any error-level violations
func hasErrors(violations []rules.Violation) bool {
	for _, v := range violations {
		if v.Severity == rules.SeverityError {
			return true
		}
	}
	return false
}

// hasWarnings checks if there are any warning-level violations
func hasWarnings(violations []rules.Violation) bool {
	for _, v := range violations {
		if v.Severity == rules.SeverityWarning {
			return true
		}
	}
	return false
}

