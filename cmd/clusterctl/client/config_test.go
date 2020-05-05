/*
Copyright 2019 The Kubernetes Authors.

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

package client

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client/cluster"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client/config"
)

func Test_clusterctlClient_GetProvidersConfig(t *testing.T) {
	customProviderConfig := config.NewProvider("custom", "url", clusterctlv1.BootstrapProviderType)

	type field struct {
		client Client
	}
	tests := []struct {
		name          string
		field         field
		wantProviders []string
		wantErr       bool
	}{
		{
			name: "Returns default providers",
			field: field{
				client: newFakeClient(newFakeConfig()),
			},
			wantProviders: []string{
				config.ClusterAPIProviderName,
				config.KubeadmBootstrapProviderName,
				config.KubeadmControlPlaneProviderName,
				config.AWSProviderName,
				config.AzureProviderName,
				config.Metal3ProviderName,
				config.OpenStackProviderName,
				config.VSphereProviderName,
			},
			wantErr: false,
		},
		{
			name: "Returns default providers and custom providers if defined",
			field: field{
				client: newFakeClient(newFakeConfig().WithProvider(customProviderConfig)),
			},
			wantProviders: []string{
				config.ClusterAPIProviderName,
				customProviderConfig.Name(),
				config.KubeadmBootstrapProviderName,
				config.KubeadmControlPlaneProviderName,
				config.AWSProviderName,
				config.AzureProviderName,
				config.Metal3ProviderName,
				config.OpenStackProviderName,
				config.VSphereProviderName,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			got, err := tt.field.client.GetProvidersConfig()
			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
				return
			}

			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(got).To(HaveLen(len(tt.wantProviders)))

			for i, gotProvider := range got {
				w := tt.wantProviders[i]
				g.Expect(gotProvider.Name()).To(Equal(w))
			}
		})
	}
}

func Test_clusterctlClient_GetProviderComponents(t *testing.T) {
	config1 := newFakeConfig().
		WithProvider(capiProviderConfig)

	repository1 := newFakeRepository(capiProviderConfig, config1).
		WithPaths("root", "components.yaml").
		WithDefaultVersion("v1.0.0").
		WithFile("v1.0.0", "components.yaml", componentsYAML("ns1"))

	client := newFakeClient(config1).
		WithRepository(repository1)

	type args struct {
		provider          string
		targetNameSpace   string
		watchingNamespace string
	}
	type want struct {
		provider config.Provider
		version  string
	}
	tests := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "Pass",
			args: args{
				provider:          capiProviderConfig.Name(),
				targetNameSpace:   "ns2",
				watchingNamespace: "",
			},
			want: want{
				provider: capiProviderConfig,
				version:  "v1.0.0",
			},
			wantErr: false,
		},
		{
			name: "Fail",
			args: args{
				provider:          fmt.Sprintf("%s:v0.2.0", capiProviderConfig.Name()),
				targetNameSpace:   "ns2",
				watchingNamespace: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			options := ComponentsInput{
				TargetNamespace:   tt.args.targetNameSpace,
				WatchingNamespace: tt.args.watchingNamespace,
			}
			got, err := client.GetProviderComponents(tt.args.provider, capiProviderConfig.Type(), options)
			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
				return
			}
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(got.Name()).To(Equal(tt.want.provider.Name()))
			g.Expect(got.Version()).To(Equal(tt.want.version))
		})
	}
}

func Test_getComponentsByName_withEmptyVariables(t *testing.T) {
	g := NewWithT(t)

	// Create a fake config with a provider named P1 and a variable named foo.
	repository1Config := config.NewProvider("p1", "url", clusterctlv1.InfrastructureProviderType)

	config1 := newFakeConfig().
		WithProvider(repository1Config)

	repository1 := newFakeRepository(repository1Config, config1).
		WithPaths("root", "components.yaml").
		WithDefaultVersion("v1.0.0").
		WithFile("v1.0.0", "components.yaml", componentsYAML("${FOO}")).
		WithMetadata("v1.0.0", &clusterctlv1.Metadata{
			ReleaseSeries: []clusterctlv1.ReleaseSeries{
				{Major: 1, Minor: 0, Contract: "v1alpha3"},
			},
		})

	// Create a fake cluster, eventually adding some existing runtime objects to it.
	cluster1 := newFakeCluster(cluster.Kubeconfig{Path: "kubeconfig", Context: "mgmt-context"}, config1).WithObjs()

	// Create a new fakeClient that allows to execute tests on the fake config,
	// the fake repositories and the fake cluster.
	client := newFakeClient(config1).
		WithRepository(repository1).
		WithCluster(cluster1)

	options := ComponentsInput{
		TargetNamespace:   "ns1",
		WatchingNamespace: "",
		SkipVariables:     true,
	}
	components, err := client.GetProviderComponents(repository1Config.Name(), repository1Config.Type(), options)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(len(components.Variables())).To(Equal(1))
	g.Expect(components.Name()).To(Equal("p1"))
}

func Test_clusterctlClient_templateOptionsToVariables(t *testing.T) {
	type args struct {
		options GetClusterTemplateOptions
	}
	tests := []struct {
		name     string
		args     args
		wantVars map[string]string
		wantErr  bool
	}{
		{
			name: "pass (using KubernetesVersion from template options)",
			args: args{
				options: GetClusterTemplateOptions{
					ClusterName:              "foo",
					TargetNamespace:          "bar",
					KubernetesVersion:        "v1.2.3",
					ControlPlaneMachineCount: pointer.Int64Ptr(1),
					WorkerMachineCount:       pointer.Int64Ptr(2),
				},
			},
			wantVars: map[string]string{
				"CLUSTER_NAME":                "foo",
				"NAMESPACE":                   "bar",
				"KUBERNETES_VERSION":          "v1.2.3",
				"CONTROL_PLANE_MACHINE_COUNT": "1",
				"WORKER_MACHINE_COUNT":        "2",
			},
			wantErr: false,
		},
		{
			name: "pass (using KubernetesVersion from env variables)",
			args: args{
				options: GetClusterTemplateOptions{
					ClusterName:              "foo",
					TargetNamespace:          "bar",
					KubernetesVersion:        "", // empty means to use value from env variables/config file
					ControlPlaneMachineCount: pointer.Int64Ptr(1),
					WorkerMachineCount:       pointer.Int64Ptr(2),
				},
			},
			wantVars: map[string]string{
				"CLUSTER_NAME":                "foo",
				"NAMESPACE":                   "bar",
				"KUBERNETES_VERSION":          "v3.4.5",
				"CONTROL_PLANE_MACHINE_COUNT": "1",
				"WORKER_MACHINE_COUNT":        "2",
			},
			wantErr: false,
		},
		{
			name: "pass (using defaults for machine counts)",
			args: args{
				options: GetClusterTemplateOptions{
					ClusterName:       "foo",
					TargetNamespace:   "bar",
					KubernetesVersion: "v1.2.3",
				},
			},
			wantVars: map[string]string{
				"CLUSTER_NAME":                "foo",
				"NAMESPACE":                   "bar",
				"KUBERNETES_VERSION":          "v1.2.3",
				"CONTROL_PLANE_MACHINE_COUNT": "1",
				"WORKER_MACHINE_COUNT":        "0",
			},
			wantErr: false,
		},
		{
			name: "fails for invalid cluster Name",
			args: args{
				options: GetClusterTemplateOptions{
					ClusterName: "A!££%",
				},
			},
			wantErr: true,
		},
		{
			name: "fails for invalid namespace Name",
			args: args{
				options: GetClusterTemplateOptions{
					ClusterName:     "foo",
					TargetNamespace: "A!££%",
				},
			},
			wantErr: true,
		},
		{
			name: "fails for invalid version",
			args: args{
				options: GetClusterTemplateOptions{
					ClusterName:       "foo",
					TargetNamespace:   "bar",
					KubernetesVersion: "A!££%",
				},
			},
			wantErr: true,
		},
		{
			name: "fails for invalid control plane machine count",
			args: args{
				options: GetClusterTemplateOptions{
					ClusterName:              "foo",
					TargetNamespace:          "bar",
					KubernetesVersion:        "v1.2.3",
					ControlPlaneMachineCount: pointer.Int64Ptr(-1),
				},
			},
			wantErr: true,
		},
		{
			name: "fails for invalid worker machine count",
			args: args{
				options: GetClusterTemplateOptions{
					ClusterName:              "foo",
					TargetNamespace:          "bar",
					KubernetesVersion:        "v1.2.3",
					ControlPlaneMachineCount: pointer.Int64Ptr(1),
					WorkerMachineCount:       pointer.Int64Ptr(-1),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			config := newFakeConfig().
				WithVar("KUBERNETES_VERSION", "v3.4.5") // with this line we are simulating an env var

			c := &clusterctlClient{
				configClient: config,
			}
			err := c.templateOptionsToVariables(tt.args.options)
			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
				return
			}
			g.Expect(err).NotTo(HaveOccurred())

			for name, wantValue := range tt.wantVars {
				gotValue, err := config.Variables().Get(name)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(gotValue).To(Equal(wantValue))
			}
		})
	}
}

func Test_clusterctlClient_templateOptionsToVariables_withExistingMachineCountVariables(t *testing.T) {
	configClient := newFakeConfig().
		WithVar("CONTROL_PLANE_MACHINE_COUNT", "3").
		WithVar("WORKER_MACHINE_COUNT", "10")

	c := &clusterctlClient{
		configClient: configClient,
	}
	options := GetClusterTemplateOptions{
		ClusterName:       "foo",
		TargetNamespace:   "bar",
		KubernetesVersion: "v1.2.3",
	}

	wantVars := map[string]string{
		"CLUSTER_NAME":                "foo",
		"NAMESPACE":                   "bar",
		"KUBERNETES_VERSION":          "v1.2.3",
		"CONTROL_PLANE_MACHINE_COUNT": "3",
		"WORKER_MACHINE_COUNT":        "10",
	}

	if err := c.templateOptionsToVariables(options); err != nil {
		t.Fatalf("error = %v", err)
	}

	for name, wantValue := range wantVars {
		gotValue, err := configClient.Variables().Get(name)
		if err != nil {
			t.Fatalf("variable %s is not definied in config variables", name)
		}
		if gotValue != wantValue {
			t.Errorf("variable %s, got = %v, want %v", name, gotValue, wantValue)
		}
	}
}

func Test_clusterctlClient_GetClusterTemplate(t *testing.T) {
	g := NewWithT(t)

	rawTemplate := templateYAML("ns3", "${ CLUSTER_NAME }")

	// Template on a file
	tmpDir, err := ioutil.TempDir("", "cc")
	g.Expect(err).NotTo(HaveOccurred())
	defer os.RemoveAll(tmpDir)

	path := filepath.Join(tmpDir, "cluster-template.yaml")
	g.Expect(ioutil.WriteFile(path, rawTemplate, 0644)).To(Succeed())

	// Template on a repository & in a ConfigMap
	configMap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ns1",
			Name:      "my-template",
		},
		Data: map[string]string{
			"prod": string(rawTemplate),
		},
	}

	config1 := newFakeConfig().
		WithProvider(infraProviderConfig)

	repository1 := newFakeRepository(infraProviderConfig, config1).
		WithPaths("root", "components").
		WithDefaultVersion("v3.0.0").
		WithFile("v3.0.0", "cluster-template.yaml", rawTemplate)

	cluster1 := newFakeCluster(cluster.Kubeconfig{Path: "kubeconfig", Context: "mgmt-context"}, config1).
		WithProviderInventory(infraProviderConfig.Name(), infraProviderConfig.Type(), "v3.0.0", "foo", "bar").
		WithObjs(configMap)

	client := newFakeClient(config1).
		WithCluster(cluster1).
		WithRepository(repository1)

	type args struct {
		options GetClusterTemplateOptions
	}

	type templateValues struct {
		variables       []string
		targetNamespace string
		yaml            []byte
	}

	tests := []struct {
		name    string
		args    args
		want    templateValues
		wantErr bool
	}{
		{
			name: "repository source - pass",
			args: args{
				options: GetClusterTemplateOptions{
					Kubeconfig: Kubeconfig{Path: "kubeconfig", Context: "mgmt-context"},
					ProviderRepositorySource: &ProviderRepositorySourceOptions{
						InfrastructureProvider: "infra:v3.0.0",
						Flavor:                 "",
					},
					ClusterName:              "test",
					TargetNamespace:          "ns1",
					ControlPlaneMachineCount: pointer.Int64Ptr(1),
				},
			},
			want: templateValues{
				variables:       []string{"CLUSTER_NAME"}, // variable detected
				targetNamespace: "ns1",
				yaml:            templateYAML("ns1", "test"), // original template modified with target namespace and variable replacement
			},
		},
		{
			name: "repository source - detects provider name/version if missing",
			args: args{
				options: GetClusterTemplateOptions{
					Kubeconfig: Kubeconfig{Path: "kubeconfig", Context: "mgmt-context"},
					ProviderRepositorySource: &ProviderRepositorySourceOptions{
						InfrastructureProvider: "", // empty triggers auto-detection of the provider name/version
						Flavor:                 "",
					},
					ClusterName:              "test",
					TargetNamespace:          "ns1",
					ControlPlaneMachineCount: pointer.Int64Ptr(1),
				},
			},
			want: templateValues{
				variables:       []string{"CLUSTER_NAME"}, // variable detected
				targetNamespace: "ns1",
				yaml:            templateYAML("ns1", "test"), // original template modified with target namespace and variable replacement
			},
		},
		{
			name: "repository source - use current namespace if targetNamespace is missing",
			args: args{
				options: GetClusterTemplateOptions{
					Kubeconfig: Kubeconfig{Path: "kubeconfig", Context: "mgmt-context"},
					ProviderRepositorySource: &ProviderRepositorySourceOptions{
						InfrastructureProvider: "infra:v3.0.0",
						Flavor:                 "",
					},
					ClusterName:              "test",
					TargetNamespace:          "", // empty triggers usage of the current namespace
					ControlPlaneMachineCount: pointer.Int64Ptr(1),
				},
			},
			want: templateValues{
				variables:       []string{"CLUSTER_NAME"}, // variable detected
				targetNamespace: "default",
				yaml:            templateYAML("default", "test"), // original template modified with target namespace and variable replacement
			},
		},
		{
			name: "URL source - pass",
			args: args{
				options: GetClusterTemplateOptions{
					Kubeconfig: Kubeconfig{Path: "kubeconfig", Context: "mgmt-context"},
					URLSource: &URLSourceOptions{
						URL: path,
					},
					ClusterName:              "test",
					TargetNamespace:          "ns1",
					ControlPlaneMachineCount: pointer.Int64Ptr(1),
				},
			},
			want: templateValues{
				variables:       []string{"CLUSTER_NAME"}, // variable detected
				targetNamespace: "ns1",
				yaml:            templateYAML("ns1", "test"), // original template modified with target namespace and variable replacement
			},
		},
		{
			name: "ConfigMap source - pass",
			args: args{
				options: GetClusterTemplateOptions{
					Kubeconfig: Kubeconfig{Path: "kubeconfig", Context: "mgmt-context"},
					ConfigMapSource: &ConfigMapSourceOptions{
						Namespace: "ns1",
						Name:      "my-template",
						DataKey:   "prod",
					},
					ClusterName:              "test",
					TargetNamespace:          "ns1",
					ControlPlaneMachineCount: pointer.Int64Ptr(1),
				},
			},
			want: templateValues{
				variables:       []string{"CLUSTER_NAME"}, // variable detected
				targetNamespace: "ns1",
				yaml:            templateYAML("ns1", "test"), // original template modified with target namespace and variable replacement
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gs := NewWithT(t)

			got, err := client.GetClusterTemplate(tt.args.options)
			if tt.wantErr {
				gs.Expect(err).To(HaveOccurred())
				return
			}
			gs.Expect(err).NotTo(HaveOccurred())

			gs.Expect(got.Variables()).To(Equal(tt.want.variables))
			gs.Expect(got.TargetNamespace()).To(Equal(tt.want.targetNamespace))

			gotYaml, err := got.Yaml()
			gs.Expect(err).NotTo(HaveOccurred())
			gs.Expect(gotYaml).To(Equal(tt.want.yaml))
		})
	}
}
