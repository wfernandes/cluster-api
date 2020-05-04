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
package repository

import logf "sigs.k8s.io/cluster-api/cmd/clusterctl/log"

type defaultTemplateProcessor struct {
	skipVariables   bool
	variablesGetter VariablesGetter
}

func newDefaultTemplateProcessor(skipVariables bool, variablesGetter VariablesGetter) *defaultTemplateProcessor {
	return &defaultTemplateProcessor{
		skipVariables:   skipVariables,
		variablesGetter: variablesGetter,
	}
}

func (tp *defaultTemplateProcessor) GetVariables(t *template) error {
	variables := inspectVariables(t.rawYaml)
	t.variables = variables
	log := logf.Log

	log.V(5).Info("GetVariables", "number of vars found", len(variables), "number of vars set", len(t.variables))
	return nil
}

func (tp *defaultTemplateProcessor) ProcessYAML(t *template) ([]byte, error) {
	log := logf.Log

	log.V(5).Info("Processing yaml", "SkipVariables", tp.skipVariables)
	processedYaml, err := replaceVariables(
		t.rawYaml,
		t.variables,
		tp.variablesGetter,
		tp.skipVariables,
	)
	if err != nil {
		return nil, err
	}
	return processedYaml, nil
}
