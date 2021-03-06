/*
Copyright 2019 Google LLC

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

package resources

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGetPullSubscriptionAnnotations(t *testing.T) {
	testCases := map[string]struct {
		want map[string]string
		got  map[string]string
	}{
		"no cluster name": {
			want: map[string]string{
				"metrics-resource-name":  "my-channel",
				"metrics-resource-group": "channels.messaging.cloud.google.com",
			},
			got: GetPullSubscriptionAnnotations("my-channel", ""),
		},
		"has cluster name": {
			want: map[string]string{
				"metrics-resource-name":  "my-channel",
				"metrics-resource-group": "channels.messaging.cloud.google.com",
				"cluster-name":           "fake-cluster-name",
			},
			got: GetPullSubscriptionAnnotations("my-channel", "fake-cluster-name"),
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			if diff := cmp.Diff(tc.want, tc.got); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
		})
	}
}

func TestGetTopicAnnotations(t *testing.T) {
	testCases := map[string]struct {
		want map[string]string
		got  map[string]string
	}{
		"no cluster name": {
			want: map[string]string{},
			got:  GetTopicAnnotations(""),
		},
		"has cluster name": {
			want: map[string]string{
				"cluster-name": "fake-cluster-name",
			},
			got: GetTopicAnnotations("fake-cluster-name"),
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			if diff := cmp.Diff(tc.want, tc.got); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
		})
	}
}
