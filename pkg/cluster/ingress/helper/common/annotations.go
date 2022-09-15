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
	"os"

	"github.com/erda-project/dice-operator/pkg/cluster/ingress/helper/types"
)

func Annotations(svcName string) map[string]string {
	enableAccessLog := "false"

	if os.Getenv(types.EnableComponentAccessLog) == "true" {
		enableAccessLog = "true"
	}

	annotation := map[string]string{
		"nginx.ingress.kubernetes.io/enable-access-log": enableAccessLog,
	}

	switch svcName {
	case "gittar", "erda-server", "ui":
		annotation["nginx.ingress.kubernetes.io/proxy-body-size"] = "0"
	default:
	}

	return annotation
}
