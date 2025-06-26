/*
Copyright 2025 The Kubernetes Authors.

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

package scorer

import (
	"context"
	"encoding/json"

	"sigs.k8s.io/gateway-api-inference-extension/pkg/epp/plugins"
	"sigs.k8s.io/gateway-api-inference-extension/pkg/epp/scheduling/framework"
	"sigs.k8s.io/gateway-api-inference-extension/pkg/epp/scheduling/types"
)

const (
	DefaultLoraAffinityScorerWeight = 1
	LoraAffinityScorerType          = "lora-affinity"
)

// compile-time type assertion
var _ framework.Scorer = &LoraAffinityScorer{}

// LoraAffinityScorerFactory defines the factory function for LoraAffinityScorer.
func LoraAffinityScorerFactory(name string, _ json.RawMessage, _ plugins.Handle) (plugins.Plugin, error) {
	return NewLoraAffinityScorer().WithName(name), nil
}

// NewLoraAffinityScorer initializes a new LoraAffinityScorer and returns its pointer.
func NewLoraAffinityScorer() *LoraAffinityScorer {
	return &LoraAffinityScorer{
		name: LoraAffinityScorerType,
	}
}

// LoraAffinityScorer scores list of candidate pods based on KV cache utilization.
type LoraAffinityScorer struct {
	name string
}

// Type returns the type of the scorer.
func (s *LoraAffinityScorer) Type() string {
	return LoraAffinityScorerType
}

// Name returns the name of the scorer.
func (s *LoraAffinityScorer) Name() string {
	return s.name
}

// WithName sets the name of the scorer.
func (s *LoraAffinityScorer) WithName(name string) *LoraAffinityScorer {
	s.name = name
	return s
}

func (s *LoraAffinityScorer) Score(_ context.Context, _ *types.CycleState, request *types.LLMRequest, pods []types.Pod) map[types.Pod]float64 {
	scores := make(map[types.Pod]float64, len(pods))

	// Categorize pods based on affinity and availability
	for _, pod := range pods {
		_, active := pod.GetMetrics().ActiveModels[request.TargetModel]
		_, waiting := pod.GetMetrics().WaitingModels[request.TargetModel]

		if active {
			scores[pod] = 1
		} else if len(pod.GetMetrics().ActiveModels)+len(pod.GetMetrics().WaitingModels) < pod.GetMetrics().MaxActiveModels {
			scores[pod] = 0.8
		} else if waiting {
			scores[pod] = 0.6
		} else {
			scores[pod] = 0.0
		}
	}

	return scores
}
