package rules

import "fmt"

// Severity levels for rule violations
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

// Category groups related rules
type Category string

const (
	CategoryOLMRequirement Category = "OLM-Requirement"
	CategoryOLMBestPractice Category = "OLM-Best-Practice"
	CategorySecurity       Category = "OLM-Security"
	CategoryUpgrade        Category = "OLM-Upgrade"
)

// Violation represents a rule violation found in a bundle
type Violation struct {
	RuleID      string   // e.g., "ODH-OLM-001"
	RuleName    string   // e.g., "missing-minkubeversion"
	Category    Category
	Severity    Severity
	Message     string
	File        string
	Line        int // 0 if not applicable
	Description string
	Fixable     bool
}

// Rule defines a validation rule for operator bundles
type Rule interface {
	// ID returns the rule identifier (e.g., "ODH-OLM-001")
	ID() string
	
	// Name returns a short name for the rule
	Name() string
	
	// Category returns the rule category
	Category() Category
	
	// Severity returns the severity level
	Severity() Severity
	
	// Description returns a detailed description
	Description() string
	
	// Validate checks the rule against a bundle
	Validate(bundle *Bundle) []Violation
	
	// Fixable returns whether the issue can be auto-fixed
	Fixable() bool
}

// Bundle represents an operator bundle structure
type Bundle struct {
	Path            string
	ManifestsPath   string
	MetadataPath    string
	CSV             *ClusterServiceVersion
	CRDs            []*CustomResourceDefinition
	OtherResources  []*Resource
	Annotations     *BundleAnnotations
}

// ClusterServiceVersion represents parsed CSV data
type ClusterServiceVersion struct {
	FilePath           string
	APIVersion         string
	Kind               string
	Metadata           Metadata
	Spec               CSVSpec
}

// CSVSpec contains the CSV specification
type CSVSpec struct {
	MinKubeVersion     string
	InstallModes       []InstallMode
	WebhookDefinitions []WebhookDefinition
	CustomResourceDefinitions CSVCustomResourceDefinitions
	Install            CSVInstall
}

// CSVCustomResourceDefinitions contains owned and required CRDs
type CSVCustomResourceDefinitions struct {
	Owned    []CRDReference
	Required []CRDReference
}

// CRDReference references a CRD
type CRDReference struct {
	Name    string
	Version string
	Kind    string
}

// CSVInstall defines the install strategy
type CSVInstall struct {
	Strategy string
	Spec     InstallSpec
}

// InstallSpec contains deployment information
type InstallSpec struct {
	Deployments []Deployment
}

// Deployment represents a deployment in the CSV
type Deployment struct {
	Name string
	Spec DeploymentSpec
}

// DeploymentSpec contains deployment details
type DeploymentSpec struct {
	Template PodTemplateSpec
}

// PodTemplateSpec contains pod template
type PodTemplateSpec struct {
	Spec PodSpec
}

// PodSpec contains pod specification
type PodSpec struct {
	Containers []Container
}

// Container represents a container
type Container struct {
	Name    string
	Image   string
	Command []string
	Args    []string
}

// InstallMode defines how the operator can be installed
type InstallMode struct {
	Type      string
	Supported bool
}

// WebhookDefinition defines a webhook in the CSV
type WebhookDefinition struct {
	Type                    string // ValidatingAdmissionWebhook, MutatingAdmissionWebhook, ConversionWebhook
	AdmissionReviewVersions []string
	DeploymentName          string
	FailurePolicy           string
	GenerateName            string
	Rules                   []WebhookRule
	SideEffects             string
	WebhookPath             string
	ConversionCRDs          []string
}

// WebhookRule defines rules for a webhook
type WebhookRule struct {
	APIGroups   []string
	APIVersions []string
	Operations  []string
	Resources   []string
}

// Metadata contains resource metadata
type Metadata struct {
	Name        string
	Namespace   string
	Annotations map[string]string
	Labels      map[string]string
}

// CustomResourceDefinition represents a CRD
type CustomResourceDefinition struct {
	FilePath   string
	APIVersion string
	Kind       string
	Metadata   Metadata
	Spec       CRDSpec
}

// CRDSpec contains CRD specification
type CRDSpec struct {
	Group                 string
	Names                 CRDNames
	Versions              []CRDVersion
	PreserveUnknownFields *bool
	Conversion            *CRDConversion
}

// CRDNames contains CRD names
type CRDNames struct {
	Kind     string
	Plural   string
	Singular string
}

// CRDVersion represents a CRD version
type CRDVersion struct {
	Name   string
	Served bool
	Storage bool
}

// CRDConversion defines conversion webhook for CRD
type CRDConversion struct {
	Strategy string
	Webhook  *CRDConversionWebhook
}

// CRDConversionWebhook defines conversion webhook details
type CRDConversionWebhook struct {
	ClientConfig *WebhookClientConfig
}

// WebhookClientConfig contains webhook client configuration
type WebhookClientConfig struct {
	Service *ServiceReference
}

// ServiceReference references a service
type ServiceReference struct {
	Name      string
	Namespace string
	Path      string
}

// Resource represents a generic Kubernetes resource
type Resource struct {
	FilePath   string
	APIVersion string
	Kind       string
	Metadata   Metadata
	Spec       map[string]interface{}
}

// BundleAnnotations contains bundle metadata annotations
type BundleAnnotations struct {
	FilePath     string
	MediaType    string
	Manifests    string
	Metadata     string
	Package      string
	Channels     []string
	DefaultChannel string
}

// String returns a formatted string representation of a violation
func (v Violation) String() string {
	loc := v.File
	if v.Line > 0 {
		loc = fmt.Sprintf("%s:%d", v.File, v.Line)
	}
	return fmt.Sprintf("[%s] %s: %s", v.RuleID, loc, v.Message)
}

