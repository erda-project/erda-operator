// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package helper

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
)

// ConvertStringSliceToHostAlias convert []string struct to
// Kubernetes k8s.io/api/core/v1 []HostAlias struct
func ConvertStringSliceToHostAlias(hosts []string) []corev1.HostAlias {
	var hostAlias []corev1.HostAlias
	for _, host := range hosts {
		splitRes := strings.Fields(host)
		if len(splitRes) < 2 {
			continue
		}
		hostAlias = append(hostAlias, corev1.HostAlias{
			IP:        splitRes[0],
			Hostnames: splitRes[1:],
		})
	}
	return hostAlias
}
