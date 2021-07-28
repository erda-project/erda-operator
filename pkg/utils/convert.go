// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package utils

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
)

// ConvertToHostAlias 转换 hosts []string 格式为 k8s.io/api/core/v1.HostAlias 格式
func ConvertToHostAlias(hosts []string) []corev1.HostAlias {
	var r []corev1.HostAlias
	for _, host := range hosts {
		splitRes := strings.Fields(host)
		if len(splitRes) < 2 {
			continue
		}
		r = append(r, corev1.HostAlias{
			IP:        splitRes[0],
			Hostnames: splitRes[1:],
		})
	}
	return r
}
