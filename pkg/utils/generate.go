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
	"os"
	"strings"

	"github.com/erda-project/dice-operator/pkg/cluster/diff"
	"github.com/erda-project/dice-operator/pkg/crd"
)

const (
	CRDKindSpecified = "CRD_KIND_SPECIFIED"
)

// GenSAName generate launch app's service account
func GenSAName(diceSvcName string) string {
	if diceSvcName == diff.ClusterAgent {
		return diceSvcName
	}

	if os.Getenv(CRDKindSpecified) != "" {
		return strings.ToLower(os.Getenv(CRDKindSpecified)) + "-operator"
	}
	return crd.CRDSingular + "-operator"
}
