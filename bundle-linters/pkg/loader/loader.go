package loader

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/opendatahub-io/odh-linter/bundle-linters/pkg/rules"
	"gopkg.in/yaml.v3"
)

// LoadBundle loads an operator bundle from a directory
func LoadBundle(bundlePath string) (*rules.Bundle, error) {
	// Normalize path
	absPath, err := filepath.Abs(bundlePath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve bundle path: %w", err)
	}

	// Check if path exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("bundle path does not exist: %s", absPath)
	}

	bundle := &rules.Bundle{
		Path:          absPath,
		ManifestsPath: filepath.Join(absPath, "manifests"),
		MetadataPath:  filepath.Join(absPath, "metadata"),
	}

	// Load bundle annotations
	if err := loadAnnotations(bundle); err != nil {
		return nil, fmt.Errorf("failed to load annotations: %w", err)
	}

	// Load manifests
	if err := loadManifests(bundle); err != nil {
		return nil, fmt.Errorf("failed to load manifests: %w", err)
	}

	return bundle, nil
}

// loadAnnotations loads the bundle annotations from metadata/annotations.yaml
func loadAnnotations(bundle *rules.Bundle) error {
	annotationsPath := filepath.Join(bundle.MetadataPath, "annotations.yaml")
	
	if _, err := os.Stat(annotationsPath); os.IsNotExist(err) {
		// Annotations file is optional in some cases
		return nil
	}

	data, err := os.ReadFile(annotationsPath)
	if err != nil {
		return fmt.Errorf("failed to read annotations file: %w", err)
	}

	var raw struct {
		Annotations map[string]string `yaml:"annotations"`
	}

	if err := yaml.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("failed to parse annotations YAML: %w", err)
	}

	bundle.Annotations = &rules.BundleAnnotations{
		FilePath:      annotationsPath,
		MediaType:     raw.Annotations["operators.operatorframework.io.bundle.mediatype.v1"],
		Manifests:     raw.Annotations["operators.operatorframework.io.bundle.manifests.v1"],
		Metadata:      raw.Annotations["operators.operatorframework.io.bundle.metadata.v1"],
		Package:       raw.Annotations["operators.operatorframework.io.bundle.package.v1"],
		DefaultChannel: raw.Annotations["operators.operatorframework.io.bundle.channel.default.v1"],
	}

	// Parse channels (comma-separated)
	if channelsStr := raw.Annotations["operators.operatorframework.io.bundle.channels.v1"]; channelsStr != "" {
		channels := strings.Split(channelsStr, ",")
		for i, ch := range channels {
			channels[i] = strings.TrimSpace(ch)
		}
		bundle.Annotations.Channels = channels
	}

	return nil
}

// loadManifests loads all manifest files from the manifests directory
func loadManifests(bundle *rules.Bundle) error {
	if _, err := os.Stat(bundle.ManifestsPath); os.IsNotExist(err) {
		return fmt.Errorf("manifests directory not found: %s", bundle.ManifestsPath)
	}

	files, err := os.ReadDir(bundle.ManifestsPath)
	if err != nil {
		return fmt.Errorf("failed to read manifests directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Only process YAML files
		if !strings.HasSuffix(file.Name(), ".yaml") && !strings.HasSuffix(file.Name(), ".yml") {
			continue
		}

		filePath := filepath.Join(bundle.ManifestsPath, file.Name())
		if err := loadManifestFile(bundle, filePath); err != nil {
			return fmt.Errorf("failed to load manifest %s: %w", file.Name(), err)
		}
	}

	return nil
}

// loadManifestFile loads a single manifest file and adds it to the bundle
func loadManifestFile(bundle *rules.Bundle, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse basic resource structure to determine kind
	var basic struct {
		APIVersion string `yaml:"apiVersion"`
		Kind       string `yaml:"kind"`
	}

	if err := yaml.Unmarshal(data, &basic); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Route to specific parser based on kind
	switch basic.Kind {
	case "ClusterServiceVersion":
		csv, err := parseCSV(filePath, data)
		if err != nil {
			return fmt.Errorf("failed to parse CSV: %w", err)
		}
		bundle.CSV = csv

	case "CustomResourceDefinition":
		crd, err := parseCRD(filePath, data)
		if err != nil {
			return fmt.Errorf("failed to parse CRD: %w", err)
		}
		bundle.CRDs = append(bundle.CRDs, crd)

	default:
		// Parse as generic resource
		resource, err := parseResource(filePath, data)
		if err != nil {
			return fmt.Errorf("failed to parse resource: %w", err)
		}
		bundle.OtherResources = append(bundle.OtherResources, resource)
	}

	return nil
}

// parseCSV parses a ClusterServiceVersion YAML file
func parseCSV(filePath string, data []byte) (*rules.ClusterServiceVersion, error) {
	var raw struct {
		APIVersion string `yaml:"apiVersion"`
		Kind       string `yaml:"kind"`
		Metadata   struct {
			Name        string            `yaml:"name"`
			Namespace   string            `yaml:"namespace"`
			Annotations map[string]string `yaml:"annotations"`
			Labels      map[string]string `yaml:"labels"`
		} `yaml:"metadata"`
		Spec struct {
			MinKubeVersion string `yaml:"minKubeVersion"`
			InstallModes   []struct {
				Type      string `yaml:"type"`
				Supported bool   `yaml:"supported"`
			} `yaml:"installModes"`
			WebhookDefinitions []struct {
				Type                    string   `yaml:"type"`
				AdmissionReviewVersions []string `yaml:"admissionReviewVersions"`
				DeploymentName          string   `yaml:"deploymentName"`
				FailurePolicy           string   `yaml:"failurePolicy"`
				GenerateName            string   `yaml:"generateName"`
				SideEffects             string   `yaml:"sideEffects"`
				WebhookPath             string   `yaml:"webhookPath"`
				ConversionCRDs          []string `yaml:"conversionCRDs"`
				Rules                   []struct {
					APIGroups   []string `yaml:"apiGroups"`
					APIVersions []string `yaml:"apiVersions"`
					Operations  []string `yaml:"operations"`
					Resources   []string `yaml:"resources"`
				} `yaml:"rules"`
			} `yaml:"webhookdefinitions"`
			CustomResourceDefinitions struct {
				Owned []struct {
					Name    string `yaml:"name"`
					Version string `yaml:"version"`
					Kind    string `yaml:"kind"`
				} `yaml:"owned"`
				Required []struct {
					Name    string `yaml:"name"`
					Version string `yaml:"version"`
					Kind    string `yaml:"kind"`
				} `yaml:"required"`
			} `yaml:"customresourcedefinitions"`
			Install struct {
				Strategy string `yaml:"strategy"`
				Spec     struct {
					Deployments []struct {
						Name string `yaml:"name"`
						Spec struct {
							Template struct {
								Spec struct {
									Containers []struct {
										Name    string   `yaml:"name"`
										Image   string   `yaml:"image"`
										Command []string `yaml:"command"`
										Args    []string `yaml:"args"`
									} `yaml:"containers"`
								} `yaml:"spec"`
							} `yaml:"template"`
						} `yaml:"spec"`
					} `yaml:"deployments"`
				} `yaml:"spec"`
			} `yaml:"install"`
		} `yaml:"spec"`
	}

	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	csv := &rules.ClusterServiceVersion{
		FilePath:   filePath,
		APIVersion: raw.APIVersion,
		Kind:       raw.Kind,
		Metadata: rules.Metadata{
			Name:        raw.Metadata.Name,
			Namespace:   raw.Metadata.Namespace,
			Annotations: raw.Metadata.Annotations,
			Labels:      raw.Metadata.Labels,
		},
		Spec: rules.CSVSpec{
			MinKubeVersion: raw.Spec.MinKubeVersion,
		},
	}

	// Parse install modes
	for _, im := range raw.Spec.InstallModes {
		csv.Spec.InstallModes = append(csv.Spec.InstallModes, rules.InstallMode{
			Type:      im.Type,
			Supported: im.Supported,
		})
	}

	// Parse webhook definitions
	for _, wd := range raw.Spec.WebhookDefinitions {
		webhook := rules.WebhookDefinition{
			Type:                    wd.Type,
			AdmissionReviewVersions: wd.AdmissionReviewVersions,
			DeploymentName:          wd.DeploymentName,
			FailurePolicy:           wd.FailurePolicy,
			GenerateName:            wd.GenerateName,
			SideEffects:             wd.SideEffects,
			WebhookPath:             wd.WebhookPath,
			ConversionCRDs:          wd.ConversionCRDs,
		}

		for _, rule := range wd.Rules {
			webhook.Rules = append(webhook.Rules, rules.WebhookRule{
				APIGroups:   rule.APIGroups,
				APIVersions: rule.APIVersions,
				Operations:  rule.Operations,
				Resources:   rule.Resources,
			})
		}

		csv.Spec.WebhookDefinitions = append(csv.Spec.WebhookDefinitions, webhook)
	}

	// Parse CRD references
	for _, owned := range raw.Spec.CustomResourceDefinitions.Owned {
		csv.Spec.CustomResourceDefinitions.Owned = append(
			csv.Spec.CustomResourceDefinitions.Owned,
			rules.CRDReference{
				Name:    owned.Name,
				Version: owned.Version,
				Kind:    owned.Kind,
			},
		)
	}

	for _, required := range raw.Spec.CustomResourceDefinitions.Required {
		csv.Spec.CustomResourceDefinitions.Required = append(
			csv.Spec.CustomResourceDefinitions.Required,
			rules.CRDReference{
				Name:    required.Name,
				Version: required.Version,
				Kind:    required.Kind,
			},
		)
	}

	// Parse install spec
	csv.Spec.Install.Strategy = raw.Spec.Install.Strategy
	for _, dep := range raw.Spec.Install.Spec.Deployments {
		deployment := rules.Deployment{
			Name: dep.Name,
		}

		for _, container := range dep.Spec.Template.Spec.Containers {
			deployment.Spec.Template.Spec.Containers = append(
				deployment.Spec.Template.Spec.Containers,
				rules.Container{
					Name:    container.Name,
					Image:   container.Image,
					Command: container.Command,
					Args:    container.Args,
				},
			)
		}

		csv.Spec.Install.Spec.Deployments = append(csv.Spec.Install.Spec.Deployments, deployment)
	}

	return csv, nil
}

// parseCRD parses a CustomResourceDefinition YAML file
func parseCRD(filePath string, data []byte) (*rules.CustomResourceDefinition, error) {
	var raw struct {
		APIVersion string `yaml:"apiVersion"`
		Kind       string `yaml:"kind"`
		Metadata   struct {
			Name        string            `yaml:"name"`
			Namespace   string            `yaml:"namespace"`
			Annotations map[string]string `yaml:"annotations"`
			Labels      map[string]string `yaml:"labels"`
		} `yaml:"metadata"`
		Spec struct {
			Group                 string `yaml:"group"`
			PreserveUnknownFields *bool  `yaml:"preserveUnknownFields"`
			Names                 struct {
				Kind     string `yaml:"kind"`
				Plural   string `yaml:"plural"`
				Singular string `yaml:"singular"`
			} `yaml:"names"`
			Versions []struct {
				Name    string `yaml:"name"`
				Served  bool   `yaml:"served"`
				Storage bool   `yaml:"storage"`
			} `yaml:"versions"`
			Conversion *struct {
				Strategy string `yaml:"strategy"`
				Webhook  *struct {
					ClientConfig *struct {
						Service *struct {
							Name      string `yaml:"name"`
							Namespace string `yaml:"namespace"`
							Path      string `yaml:"path"`
						} `yaml:"service"`
					} `yaml:"clientConfig"`
				} `yaml:"webhook"`
			} `yaml:"conversion"`
		} `yaml:"spec"`
	}

	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	crd := &rules.CustomResourceDefinition{
		FilePath:   filePath,
		APIVersion: raw.APIVersion,
		Kind:       raw.Kind,
		Metadata: rules.Metadata{
			Name:        raw.Metadata.Name,
			Namespace:   raw.Metadata.Namespace,
			Annotations: raw.Metadata.Annotations,
			Labels:      raw.Metadata.Labels,
		},
		Spec: rules.CRDSpec{
			Group:                 raw.Spec.Group,
			PreserveUnknownFields: raw.Spec.PreserveUnknownFields,
			Names: rules.CRDNames{
				Kind:     raw.Spec.Names.Kind,
				Plural:   raw.Spec.Names.Plural,
				Singular: raw.Spec.Names.Singular,
			},
		},
	}

	// Parse versions
	for _, v := range raw.Spec.Versions {
		crd.Spec.Versions = append(crd.Spec.Versions, rules.CRDVersion{
			Name:    v.Name,
			Served:  v.Served,
			Storage: v.Storage,
		})
	}

	// Parse conversion
	if raw.Spec.Conversion != nil {
		crd.Spec.Conversion = &rules.CRDConversion{
			Strategy: raw.Spec.Conversion.Strategy,
		}

		if raw.Spec.Conversion.Webhook != nil {
			crd.Spec.Conversion.Webhook = &rules.CRDConversionWebhook{}

			if raw.Spec.Conversion.Webhook.ClientConfig != nil {
				crd.Spec.Conversion.Webhook.ClientConfig = &rules.WebhookClientConfig{}

				if raw.Spec.Conversion.Webhook.ClientConfig.Service != nil {
					crd.Spec.Conversion.Webhook.ClientConfig.Service = &rules.ServiceReference{
						Name:      raw.Spec.Conversion.Webhook.ClientConfig.Service.Name,
						Namespace: raw.Spec.Conversion.Webhook.ClientConfig.Service.Namespace,
						Path:      raw.Spec.Conversion.Webhook.ClientConfig.Service.Path,
					}
				}
			}
		}
	}

	return crd, nil
}

// parseResource parses a generic Kubernetes resource YAML file
func parseResource(filePath string, data []byte) (*rules.Resource, error) {
	var raw struct {
		APIVersion string                 `yaml:"apiVersion"`
		Kind       string                 `yaml:"kind"`
		Metadata   struct {
			Name        string            `yaml:"name"`
			Namespace   string            `yaml:"namespace"`
			Annotations map[string]string `yaml:"annotations"`
			Labels      map[string]string `yaml:"labels"`
		} `yaml:"metadata"`
		Spec map[string]interface{} `yaml:"spec"`
	}

	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	return &rules.Resource{
		FilePath:   filePath,
		APIVersion: raw.APIVersion,
		Kind:       raw.Kind,
		Metadata: rules.Metadata{
			Name:        raw.Metadata.Name,
			Namespace:   raw.Metadata.Namespace,
			Annotations: raw.Metadata.Annotations,
			Labels:      raw.Metadata.Labels,
		},
		Spec: raw.Spec,
	}, nil
}

