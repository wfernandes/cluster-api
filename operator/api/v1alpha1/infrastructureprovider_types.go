/*
Copyright 2020 The Kubernetes Authors.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// InfrastructureProviderSpec defines the desired state of InfrastructureProvider
type InfrastructureProviderSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ProviderSpec `json:",inline"`
}

// InfrastructureProviderStatus defines the observed state of InfrastructureProvider
type InfrastructureProviderStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	ProviderStatus `json:",inline"`
}

// +kubebuilder:object:root=true

// InfrastructureProvider is the Schema for the infrastructureproviders API
type InfrastructureProvider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InfrastructureProviderSpec   `json:"spec,omitempty"`
	Status InfrastructureProviderStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// InfrastructureProviderList contains a list of InfrastructureProvider
type InfrastructureProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InfrastructureProvider `json:"items"`
}

func init() {
	SchemeBuilder.Register(&InfrastructureProvider{}, &InfrastructureProviderList{})
}
