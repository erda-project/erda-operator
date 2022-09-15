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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/apimachinery/pkg/util/sets"
)

var (
	serverVersions = sets.String{}
)

func InitializeGroups(c kubernetes.Interface) error {
	groups, err := c.Discovery().ServerGroups()
	if err != nil {
		return err
	}

	versions := metav1.ExtractGroupVersions(groups)
	for _, v := range versions {
		serverVersions.Insert(v)
	}

	return err
}

func GetServerVersions() sets.String {
	return serverVersions
}

func VersionHas(v string) bool {
	return serverVersions.Has(v)
}
