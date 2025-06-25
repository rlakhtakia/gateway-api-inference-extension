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

package picker

import (
	"context"
	"testing"

	// Import config for thresholds
	"github.com/google/go-cmp/cmp"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api-inference-extension/pkg/epp/backend"
	"sigs.k8s.io/gateway-api-inference-extension/pkg/epp/scheduling/framework"
	"sigs.k8s.io/gateway-api-inference-extension/pkg/epp/scheduling/types"
)

var _ framework.Picker = &TestRandomPicker{}

// NewTestRandomPicker initializes a new NewTestRandomPicker and returns its pointer.
func NewTestRandomPicker(pickRes string) *TestRandomPicker {
	return &TestRandomPicker{
		Picked: pickRes,
		name:   "random-test",
	}
}

// NewTestRandomPicker picks the selected pod from the list of candidates.
type TestRandomPicker struct {
	Picked string
	name   string
}

// Name returns the name of the picker.
func (p *TestRandomPicker) Name() string {
	return p.name
}

// Type returns the type of the picker.
func (p *TestRandomPicker) Type() string {
	return RandomPickerType
}

func (tp *TestRandomPicker) Pick(_ context.Context, _ *types.CycleState, scoredPods []*types.ScoredPod) *types.ProfileRunResult {
	fallbackPods := []*types.ScoredPod{}

	var winnerPod types.Pod
	for _, scoredPod := range scoredPods {
		if scoredPod.GetPod().NamespacedName.String() == tp.Picked {
			winnerPod = scoredPod
		} else {
			fallbackPods = append(fallbackPods, scoredPod)
		}
	}

	return &types.ProfileRunResult{TargetPod: winnerPod, FallbackPods: fallbackPods}
}

func TestMaxScorePicker(t *testing.T) {
	tests := []struct {
		name        string
		scoredPods  []*types.ScoredPod
		wantProfile *types.ProfileRunResult
		picker      framework.Picker
		err         bool
	}{

		{
			name: "Single pod in list",
			scoredPods: []*types.ScoredPod{
				{
					Pod: &types.PodMetrics{
						Pod: &backend.Pod{
							Address:        "192.168.1.100",
							NamespacedName: k8stypes.NamespacedName{Name: "pod1", Namespace: "default"},
						},
					},
					Score: 0.8,
				},
			},
			wantProfile: &types.ProfileRunResult{
				TargetPod: &types.ScoredPod{
					Pod: &types.PodMetrics{
						Pod: &backend.Pod{
							Address:        "192.168.1.100",
							NamespacedName: k8stypes.NamespacedName{Name: "pod1", Namespace: "default"},
						},
					},
					Score: 0.8,
				},
				FallbackPods: []*types.ScoredPod{},
			},
			picker: NewMaxScorePicker(),
		},
		{
			name: "Multiple pods with different scores in list",
			scoredPods: []*types.ScoredPod{
				{
					Pod: &types.PodMetrics{
						Pod: &backend.Pod{
							Address:        "192.168.1.100",
							NamespacedName: k8stypes.NamespacedName{Name: "pod1", Namespace: "default"},
						},
					},
					Score: 8.0,
				},
				{
					Pod: &types.PodMetrics{
						Pod: &backend.Pod{
							Address:        "192.168.2.100",
							NamespacedName: k8stypes.NamespacedName{Name: "pod2", Namespace: "default"},
						},
					},
					Score: 12.0,
				},
				{
					Pod: &types.PodMetrics{
						Pod: &backend.Pod{
							Address:        "192.168.3.100",
							NamespacedName: k8stypes.NamespacedName{Name: "pod3", Namespace: "default"},
						},
					},
					Score: 5.0,
				},
				{
					Pod: &types.PodMetrics{
						Pod: &backend.Pod{
							Address:        "192.168.4.100",
							NamespacedName: k8stypes.NamespacedName{Name: "pod4", Namespace: "default"},
						},
					},
					Score: 9.0,
				},
			},
			wantProfile: &types.ProfileRunResult{
				TargetPod: &types.ScoredPod{
					Pod: &types.PodMetrics{
						Pod: &backend.Pod{
							Address:        "192.168.2.100",
							NamespacedName: k8stypes.NamespacedName{Name: "pod2", Namespace: "default"},
						},
					},
					Score: 12.0,
				},
				FallbackPods: []*types.ScoredPod{
					{
						Pod: &types.PodMetrics{
							Pod: &backend.Pod{
								Address:        "192.168.4.100",
								NamespacedName: k8stypes.NamespacedName{Name: "pod4", Namespace: "default"},
							},
						},
						Score: 9.0,
					},
					{
						Pod: &types.PodMetrics{
							Pod: &backend.Pod{
								Address:        "192.168.1.100",
								NamespacedName: k8stypes.NamespacedName{Name: "pod1", Namespace: "default"},
							},
						},
						Score: 8.0,
					},
					{
						Pod: &types.PodMetrics{
							Pod: &backend.Pod{
								Address:        "192.168.3.100",
								NamespacedName: k8stypes.NamespacedName{Name: "pod3", Namespace: "default"},
							},
						},
						Score: 5.0,
					},
				},
			},
			picker: NewMaxScorePicker(),
		},
		{
			name: "Multiple pods with same score",
			scoredPods: []*types.ScoredPod{
				{
					Pod: &types.PodMetrics{
						Pod: &backend.Pod{
							Address:        "192.168.1.100",
							NamespacedName: k8stypes.NamespacedName{Name: "pod1", Namespace: "default"},
						},
					},
					Score: 8.0,
				},
				{
					Pod: &types.PodMetrics{
						Pod: &backend.Pod{
							Address:        "192.168.2.100",
							NamespacedName: k8stypes.NamespacedName{Name: "pod2", Namespace: "default"},
						},
					},
					Score: 8.0,
				},
				{
					Pod: &types.PodMetrics{
						Pod: &backend.Pod{
							Address:        "192.168.3.100",
							NamespacedName: k8stypes.NamespacedName{Name: "pod3", Namespace: "default"},
						},
					},
					Score: 8.0,
				},
				{
					Pod: &types.PodMetrics{
						Pod: &backend.Pod{
							Address:        "192.168.4.100",
							NamespacedName: k8stypes.NamespacedName{Name: "pod4", Namespace: "default"},
						},
					},
					Score: 8.0,
				},
			},
			wantProfile: &types.ProfileRunResult{
				TargetPod: &types.ScoredPod{
					Pod: &types.PodMetrics{
						Pod: &backend.Pod{
							Address:        "192.168.2.100",
							NamespacedName: k8stypes.NamespacedName{Name: "pod2", Namespace: "default"},
						},
					},
					Score: 8.0,
				},
				FallbackPods: []*types.ScoredPod{
					{
						Pod: &types.PodMetrics{
							Pod: &backend.Pod{
								Address:        "192.168.1.100",
								NamespacedName: k8stypes.NamespacedName{Name: "pod1", Namespace: "default"},
							},
						},
						Score: 8.0,
					},
					{
						Pod: &types.PodMetrics{
							Pod: &backend.Pod{
								Address:        "192.168.3.100",
								NamespacedName: k8stypes.NamespacedName{Name: "pod3", Namespace: "default"},
							},
						},
						Score: 8.0,
					},
					{
						Pod: &types.PodMetrics{
							Pod: &backend.Pod{
								Address:        "192.168.4.100",
								NamespacedName: k8stypes.NamespacedName{Name: "pod4", Namespace: "default"},
							},
						},
						Score: 8.0,
					},
				},
			},
			picker: NewTestRandomPicker(k8stypes.NamespacedName{Name: "pod2", Namespace: "default"}.String()),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			got := test.picker.Pick(ctx, nil, test.scoredPods)

			if diff := cmp.Diff(test.wantProfile, got); diff != "" {
				t.Errorf("Unexpected output (-want +got): %v", diff)
			}
		})
	}
}
