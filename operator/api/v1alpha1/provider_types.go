/*


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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha3"
	ctrlcfg "sigs.k8s.io/controller-runtime/pkg/config/v1alpha1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ProviderSpec defines the desired state of CoreProvider
type ProviderSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Version indicates the provider version.
	// +optional
	Version string `json:"version,omitempty"`

	// Manager defines the properties that can be enabled on the controller manager for the provider.
	// +optional
	Manager ManagerSpec `json:"manager,omitempty"`

	// Deployment defines the properties that can be enabled on the deployment for the provider.
	// +optional
	Deployment DeploymentSpec `json:"deployment,omitempty"`

	// SecretName is the name of the Secret providing the configuration
	// variables for the current provider instance, like e.g. credentials.
	// Such configurations will be used when creating or upgrading provider components.
	// +optional
	SecretName *string

	// FetchConfig determines how the operator will fetch the components and metadata for the provider.
	// If nil, the operator will try to fetch components according to default
	// settings embedded in the operator and in clusterctl.
	// +optional
	FetchConfig *FetchConfiguration `json:"fetchConfig,omitempty"`

	// Paused prevents the operator from reconciling the provider. This can be
	// used when doing an upgrade or move action manually.
	// +optional
	Paused bool `json:"paused,omitempty"`
}

// ManagerSpec defines the properties that can be enabled on the controller manager for the provider.
type ManagerSpec struct {
	// ControllerManagerConfigurationSpec defines the desired state of GenericControllerManagerConfiguration.
	ctrlcfg.ControllerManagerConfigurationSpec

	// ProfilerAddress defines the bind address to expose the pprof profiler (e.g. localhost:6060).
	// Default empty, meaning the profiler is disabled.
	// +optional
	ProfilerAddress string `json:"profilerAddress,omitempty"`

	// MaxConcurrentReconciles is the maximum number of concurrent Reconciles
	// which can be run. Defaults to 10.
	// +optional
	MaxConcurrentReconciles *int `json:"maxConcurrentReconciles,omitempty"`

	// Verbosity set the logs verbosity. Defaults to 1.
	// +optional
	Verbosity int `json:"verbosity,omitempty"`

	// Debug, if set, will override a set of fields with opinionated values for
	// a debugging session. (Verbosity=5, ProfilerAddress=localhost:6060)
	// +optional
	Debug bool `json:"debug, omitempty"`

	// FeatureGates define provider specific feature flags that will be passed
	// in as container args to the provider's controller manager.
	FeatureGates map[string]bool `json:"featureGates, omitempty"`
}

// DeploymentSpec defines the properties that can be enabled on the Deployment for the provider.
type DeploymentSpec struct {
	// Number of desired pods. This is a pointer to distinguish between explicit zero and not specified. Defaults to 1.
	// +optional
	Replicas int `json:"replicas,omitempty"`

	// List of containers specified in the Deployment
	// +optional
	Containers []ContainerSpec `json:"containers"`
}

// ContainerSpec defines the properties available to override for each
// container in a provider deployment such as Image and Args to the container’s
// entrypoint.
type ContainerSpec struct {
	// Name of the container. Cannot be updated.
	Name string `json:"name"`

	// Docker image name
	// +optional
	Image string `json:"image,omitempty"`

	// Args represents extra provider specific flags that are not encoded as fields in this API.
	// +optional
	Args map[string]string `json:"args,omitempty"`

	// Compute resources required by this container.
	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
}

// FetchConfiguration determines the way to fetch the components and metadata for the provider.
type FetchConfiguration struct {
	// URL to be used for fetching provider’s components and metadata from a remote repository.
	// +optional
	URL *string `json:"url,omitempty"`

	// Selector to be used for fetching provider’s components and metadata from
	// ConfigMaps stored inside the cluster. Each ConfigMap is expected to contain
	// components and metadata for a specific version only.
	// +optional
	Selector *metav1.LabelSelector `json:"selector,omitempty"`
}

// ProviderStatus defines the observed state of CoreProvider
type ProviderStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Contract will contain the core provider contract that the provider is
	// abiding by, like e.g. v1alpha3.
	Contract string `json:"contract,omitempty"`

	// Conditions define the current service state of the cluster.
	// +optional
	Conditions clusterv1.Conditions `json:"conditions,omitempty"`

	// ObservedGeneration is the latest generation observed by the controller.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}
