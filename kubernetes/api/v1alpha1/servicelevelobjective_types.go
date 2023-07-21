/*
Copyright 2021 Pyrra Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/yaml"

	"github.com/pyrra-dev/pyrra/slo"
)

func init() {
	SchemeBuilder.Register(&ServiceLevelObjective{}, &ServiceLevelObjectiveList{})
}

var _ webhook.Validator = &ServiceLevelObjective{}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +kubebuilder:object:root=true

// ServiceLevelObjectiveList contains a list of ServiceLevelObjective.
type ServiceLevelObjectiveList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceLevelObjective `json:"items"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=slo
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Window",type=string,JSONPath=`.spec.window`
// +kubebuilder:printcolumn:name="Target",type=string,JSONPath=`.spec.target`
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.status.type`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// ServiceLevelObjective is the Schema for the ServiceLevelObjectives API.
type ServiceLevelObjective struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceLevelObjectiveSpec   `json:"spec,omitempty"`
	Status ServiceLevelObjectiveStatus `json:"status,omitempty"`
}

// ServiceLevelObjectiveSpec defines the desired state of ServiceLevelObjective.
type ServiceLevelObjectiveSpec struct {
	// +optional
	// Description describes the ServiceLevelObjective in more detail and
	// gives extra context for engineers that might not directly work on the service.
	Description string `json:"description"`

	// Target is a string that's casted to a float64 between 0 - 100.
	// It represents the desired availability of the service in the given window.
	// float64 are not supported: https://github.com/kubernetes-sigs/controller-tools/issues/245
	Target string `json:"target"`

	// Window within which the Target is supposed to be kept. Usually something like 1d, 7d or 28d.
	Window string `json:"window"`

	// ServiceLevelIndicator is the underlying data source that indicates how the service is doing.
	// This will be a Prometheus metric with specific selectors for your service.
	ServiceLevelIndicator ServiceLevelIndicator `json:"indicator"`

	// +optional
	// Alerting customizes the alerting rules generated by Pyrra.
	Alerting Alerting `json:"alerting"`
}

// ServiceLevelIndicator defines the underlying indicator that is a Prometheus metric.
type ServiceLevelIndicator struct {
	// +optional
	// Ratio is the indicator that measures against errors / total events.
	Ratio *RatioIndicator `json:"ratio,omitempty"`

	// +optional
	// Latency is the indicator that measures a certain percentage to be faster than the expected latency.
	Latency *LatencyIndicator `json:"latency,omitempty"`

	// +optional
	// LatencyNative is the indicator that measures a certain percentage to be faster than the expected latency.
	// This uses the new native histograms in Prometheus.
	LatencyNative *NativeLatencyIndicator `json:"latencyNative,omitempty"`

	// +optional
	// BoolGauge is the indicator that measures whether a boolean gauge is
	// successful.
	BoolGauge *BoolGaugeIndicator `json:"bool_gauge,omitempty"`
}

type Alerting struct {
	// +optional
	// Disabled is used to disable the generation of alerts. Recording rules are still generated.
	Disabled *bool `json:"disabled,omitempty"`

	// +optional
	// Name is used as the name of the alert generated by Pyrra. Defaults to "ErrorBudgetBurn".
	Name string `json:"name,omitempty"`
}

type RatioIndicator struct {
	// Errors is the metric that returns how many errors there are.
	Errors Query `json:"errors"`
	// Total is the metric that returns how many requests there are in total.
	Total Query `json:"total"`
	// +optional
	// Grouping allows an SLO to be defined for many SLI at once, like HTTP handlers for example.
	Grouping []string `json:"grouping"`
}

type LatencyIndicator struct {
	// Success is the metric that returns how many errors there are.
	Success Query `json:"success"`
	// Total is the metric that returns how many requests there are in total.
	Total Query `json:"total"`
	// +optional
	// Grouping allows an SLO to be defined for many SLI at once, like HTTP handlers for example.
	Grouping []string `json:"grouping"`
}

type NativeLatencyIndicator struct {
	// Total is the metric that returns how many requests there are in total.
	Total Query `json:"total"`

	// Latency the requests should be faster than.
	Latency string `json:"latency"`

	// +optional
	// Grouping allows an SLO to be defined for many SLI at once, like HTTP handlers for example.
	Grouping []string `json:"grouping"`
}

type BoolGaugeIndicator struct {
	Query `json:",inline"`
	// Total is the metric that returns how many requests there are in total.
	Grouping []string `json:"grouping"`
}

// Query contains a PromQL metric.
type Query struct {
	Metric string `json:"metric"`
}

// ServiceLevelObjectiveStatus defines the observed state of ServiceLevelObjective.
type ServiceLevelObjectiveStatus struct {
	// Type is the generated resource type, like PrometheusRule or ConfigMap
	Type string `json:"type,omitempty"`
}

func (in *ServiceLevelObjective) ValidateCreate() (admission.Warnings, error) {
	return in.validate()
}

func (in *ServiceLevelObjective) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	return in.validate()
}

func (in *ServiceLevelObjective) ValidateDelete() (admission.Warnings, error) {
	return nil, nil
}

func (in *ServiceLevelObjective) validate() (admission.Warnings, error) {
	var warnings []string

	if in.GetName() == "" {
		return warnings, fmt.Errorf("name must be set")
	}
	if in.GetNamespace() == "" {
		warnings = append(warnings, "namespace must be set")
	}

	if in.Spec.Target == "" {
		return warnings, fmt.Errorf("target must be set")
	}
	target, err := strconv.ParseFloat(in.Spec.Target, 64)
	if err != nil {
		return warnings, err
	}
	if target < 0 || target > 100 {
		return warnings, fmt.Errorf("target must be between 0 and 100")
	}
	if target > 0 && target < 1 {
		warnings = append(warnings, fmt.Sprintf("target is from 0-100 (%v), not 0-1 (%v)", 100*target, target))
	}

	if in.Spec.Window == "" {
		return warnings, fmt.Errorf("window must be set")
	}
	_, err = model.ParseDuration(in.Spec.Window)
	if err != nil {
		return warnings, err
	}

	if in.Spec.ServiceLevelIndicator.Ratio == nil &&
		in.Spec.ServiceLevelIndicator.Latency == nil &&
		in.Spec.ServiceLevelIndicator.LatencyNative == nil &&
		in.Spec.ServiceLevelIndicator.BoolGauge == nil {
		return warnings, fmt.Errorf("one of ratio, latency, latencyNative or bool_gauge must be set")
	}

	if in.Spec.ServiceLevelIndicator.Ratio != nil {
		ratio := in.Spec.ServiceLevelIndicator.Ratio
		if ratio.Total.Metric == "" {
			return warnings, fmt.Errorf("ratio total metric must be set")
		}
		if ratio.Errors.Metric == "" {
			return warnings, fmt.Errorf("ratio errors metric must be set")
		}

		if ratio.Errors.Metric == ratio.Total.Metric {
			warnings = append(warnings, "ratio errors metric should be different from ratio total metric")
		}

		_, err := parser.ParseExpr(ratio.Total.Metric)
		if err != nil {
			return warnings, fmt.Errorf("failed to parse ratio total metric: %w", err)
		}
		_, err = parser.ParseExpr(ratio.Errors.Metric)
		if err != nil {
			return warnings, fmt.Errorf("failed to parse ratio error metric: %w", err)
		}
	}

	return warnings, nil
}

func (in *ServiceLevelObjective) Internal() (slo.Objective, error) {
	target, err := strconv.ParseFloat(in.Spec.Target, 64)
	if err != nil {
		return slo.Objective{}, fmt.Errorf("failed to parse objective target: %w", err)
	}

	window, err := model.ParseDuration(in.Spec.Window)
	if err != nil {
		return slo.Objective{}, fmt.Errorf("failed to parse objective window: %w", err)
	}
	var alerting slo.Alerting
	alerting.Disabled = false
	if in.Spec.Alerting.Disabled != nil {
		alerting.Disabled = *in.Spec.Alerting.Disabled
	}

	if in.Spec.Alerting.Name != "" {
		alerting.Name = in.Spec.Alerting.Name
	}

	if in.Spec.ServiceLevelIndicator.Ratio != nil && in.Spec.ServiceLevelIndicator.Latency != nil {
		return slo.Objective{}, fmt.Errorf("cannot have ratio and latency indicators at the same time")
	}

	var ratio *slo.RatioIndicator
	if in.Spec.ServiceLevelIndicator.Ratio != nil {
		totalExpr, err := parser.ParseExpr(in.Spec.ServiceLevelIndicator.Ratio.Total.Metric)
		if err != nil {
			return slo.Objective{}, err
		}

		totalVec, ok := totalExpr.(*parser.VectorSelector)
		if !ok {
			return slo.Objective{}, fmt.Errorf("ratio total metric is not a VectorSelector")
		}

		errorExpr, err := parser.ParseExpr(in.Spec.ServiceLevelIndicator.Ratio.Errors.Metric)
		if err != nil {
			return slo.Objective{}, err
		}

		errorVec, ok := errorExpr.(*parser.VectorSelector)
		if !ok {
			return slo.Objective{}, fmt.Errorf("ratio error metric is not a VectorSelector")
		}

		// Copy the matchers to get rid of the re field for unit testing...
		errorMatchers := make([]*labels.Matcher, len(errorVec.LabelMatchers))
		for i, matcher := range errorVec.LabelMatchers {
			errorMatchers[i] = &labels.Matcher{Type: matcher.Type, Name: matcher.Name, Value: matcher.Value}
		}

		ratio = &slo.RatioIndicator{
			Errors: slo.Metric{
				Name:          errorVec.Name,
				LabelMatchers: errorMatchers,
			},
			Total: slo.Metric{
				Name:          totalVec.Name,
				LabelMatchers: totalVec.LabelMatchers,
			},
			Grouping: in.Spec.ServiceLevelIndicator.Ratio.Grouping,
		}
	}

	var latency *slo.LatencyIndicator
	if in.Spec.ServiceLevelIndicator.Latency != nil {
		totalExpr, err := parser.ParseExpr(in.Spec.ServiceLevelIndicator.Latency.Total.Metric)
		if err != nil {
			return slo.Objective{}, err
		}

		totalVec, ok := totalExpr.(*parser.VectorSelector)
		if !ok {
			return slo.Objective{}, fmt.Errorf("latency total metric is not a VectorSelector")
		}

		// Copy the matchers to get rid of the re field for unit testing...
		totalMatchers := make([]*labels.Matcher, len(totalVec.LabelMatchers))
		for i, matcher := range totalVec.LabelMatchers {
			totalMatchers[i] = &labels.Matcher{Type: matcher.Type, Name: matcher.Name, Value: matcher.Value}
		}

		successExpr, err := parser.ParseExpr(in.Spec.ServiceLevelIndicator.Latency.Success.Metric)
		if err != nil {
			return slo.Objective{}, err
		}

		successVec, ok := successExpr.(*parser.VectorSelector)
		if !ok {
			return slo.Objective{}, fmt.Errorf("latency success metric is not a VectorSelector")
		}

		// Copy the matchers to get rid of the re field for unit testing...
		successMatchers := make([]*labels.Matcher, len(successVec.LabelMatchers))
		for i, matcher := range successVec.LabelMatchers {
			successMatchers[i] = &labels.Matcher{Type: matcher.Type, Name: matcher.Name, Value: matcher.Value}
		}

		latency = &slo.LatencyIndicator{
			Success: slo.Metric{
				Name:          successVec.Name,
				LabelMatchers: successMatchers,
			},
			Total: slo.Metric{
				Name:          totalVec.Name,
				LabelMatchers: totalMatchers,
			},
			Grouping: in.Spec.ServiceLevelIndicator.Latency.Grouping,
		}
	}

	var latencyNative *slo.LatencyNativeIndicator
	if in.Spec.ServiceLevelIndicator.LatencyNative != nil {
		latency, err := model.ParseDuration(in.Spec.ServiceLevelIndicator.LatencyNative.Latency)
		if err != nil {
			return slo.Objective{}, fmt.Errorf("failed to parse objective latency: %w", err)
		}

		totalExpr, err := parser.ParseExpr(in.Spec.ServiceLevelIndicator.LatencyNative.Total.Metric)
		if err != nil {
			return slo.Objective{}, err
		}

		totalVec, ok := totalExpr.(*parser.VectorSelector)
		if !ok {
			return slo.Objective{}, fmt.Errorf("latency total metric is not a VectorSelector")
		}

		// Copy the matchers to get rid of the re field for unit testing...
		totalMatchers := make([]*labels.Matcher, len(totalVec.LabelMatchers))
		for i, matcher := range totalVec.LabelMatchers {
			totalMatchers[i] = &labels.Matcher{Type: matcher.Type, Name: matcher.Name, Value: matcher.Value}
		}

		latencyNative = &slo.LatencyNativeIndicator{
			Latency: latency,
			Total: slo.Metric{
				Name:          totalVec.Name,
				LabelMatchers: totalMatchers,
			},
			Grouping: in.Spec.ServiceLevelIndicator.LatencyNative.Grouping,
		}
	}

	var boolGauge *slo.BoolGaugeIndicator
	if in.Spec.ServiceLevelIndicator.BoolGauge != nil {
		expr, err := parser.ParseExpr(in.Spec.ServiceLevelIndicator.BoolGauge.Metric)
		if err != nil {
			return slo.Objective{}, err
		}

		vec, ok := expr.(*parser.VectorSelector)
		if !ok {
			return slo.Objective{}, fmt.Errorf("bool gauge metric is not a VectorSelector")
		}

		// Copy the matchers to get rid of the re field for unit testing...
		matchers := make([]*labels.Matcher, len(vec.LabelMatchers))
		for i, matcher := range vec.LabelMatchers {
			matchers[i] = &labels.Matcher{Type: matcher.Type, Name: matcher.Name, Value: matcher.Value}
		}

		boolGauge = &slo.BoolGaugeIndicator{
			Metric: slo.Metric{
				Name:          vec.Name,
				LabelMatchers: matchers,
			},
			Grouping: in.Spec.ServiceLevelIndicator.BoolGauge.Grouping,
		}
	}

	inCopy := in.DeepCopy()
	inCopy.ManagedFields = nil
	delete(inCopy.Annotations, "kubectl.kubernetes.io/last-applied-configuration")

	config, err := yaml.Marshal(inCopy)
	if err != nil {
		return slo.Objective{}, fmt.Errorf("failed to marshal resource as config")
	}

	ls := labels.Labels{{Name: labels.MetricName, Value: in.GetName()}}

	if in.GetNamespace() != "" {
		ls = append(ls, labels.Label{
			Name: "namespace", Value: in.GetNamespace(),
		})
	}

	for name, value := range in.GetLabels() {
		if strings.HasPrefix(name, slo.PropagationLabelsPrefix) {
			ls = append(ls, labels.Label{Name: name, Value: value})
		}
	}

	return slo.Objective{
		Labels:      ls,
		Annotations: in.Annotations,
		Description: in.Spec.Description,
		Target:      target / 100,
		Window:      window,
		Config:      string(config),
		Alerting:    alerting,
		Indicator: slo.Indicator{
			Ratio:         ratio,
			Latency:       latency,
			LatencyNative: latencyNative,
			BoolGauge:     boolGauge,
		},
	}, nil
}
