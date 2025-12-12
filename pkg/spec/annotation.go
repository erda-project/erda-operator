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

package spec

import "fmt"

const (
	rootDomain = "erda.cloud"
)

type AnnotationSpec struct {
	SubSystem string
	Key       string
}

func (a AnnotationSpec) String() string {
	return fmt.Sprintf("%s.%s/%s", a.SubSystem, rootDomain, a.Key)
}

var (
	AnnotationAppUpdateStrategy = AnnotationSpec{
		SubSystem: "app",
		Key:       "update-strategy",
	}
)
