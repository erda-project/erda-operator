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

package size

import (
	"github.com/erda-project/dice-operator/pkg/spec"
)

func MemOverCommit(size spec.ClusterSize) int {
	switch size {
	case spec.ClusterSizeTest:
		return 2
	case spec.ClusterSizeProd:
		return 1
	default:
		return 2
	}
}

func CpuOverCommit(size spec.ClusterSize) int {
	switch size {
	case spec.ClusterSizeTest:
		return 6
	case spec.ClusterSizeProd:
		return 3
	default:
		return 6
	}
}
