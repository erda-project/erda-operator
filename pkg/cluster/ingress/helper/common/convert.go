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

package common

import (
	"github.com/erda-project/dice-operator/pkg/spec"
	"github.com/erda-project/dice-operator/pkg/cluster/ingress/helper/types"
	"fmt"
	"github.com/erda-project/erda/pkg/strutil"
)

func ConvertHost(svcName string, cluster *spec.DiceCluster) []string {
	customDomain := cluster.Spec.CustomDomain

	// Rename, logic will in tools render future.
	// erda-server.* will convert to openapi.*
	if svcName == types.ErdaServer {
		svcName = types.Openapi
	}

	if domains, ok := customDomain[svcName]; ok {
		return strutil.Map(strutil.Split(domains, ",", true), func(s string) string { return strutil.Trim(s) })
	}
	r, ok := map[string][]string{
		"ui": {
			fmt.Sprintf("dice.%s", cluster.Spec.PlatformDomain),
			fmt.Sprintf("*.%s", cluster.Spec.PlatformDomain),
		},
	}[svcName]
	if !ok {
		return []string{fmt.Sprintf("%s.%s", svcName, cluster.Spec.PlatformDomain)}
	}
	return r
}
