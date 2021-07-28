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

package diff

import (
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_DeploymentListDiff(t *testing.T) {
	current := []*appsv1.Deployment{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "bbb"},
		},
	}
	target := []*appsv1.Deployment{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "aaa"},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "bbb"},
		},
	}
	d := NewDeploymentListDiff(current, target)
	a := d.GetActions()
	assert.True(t, len(a.AddedDeployments) == 1)
}

func Test_Annotations(t *testing.T) {
	current := appsv1.Deployment{

		ObjectMeta: metav1.ObjectMeta{
			Name: "bbb",
		},
	}
	target := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "bbb",
			Annotations: map[string]string{
				"dice/a1": "a1",
				"a1":      "a2",
			},
		},
	}
	for k, v := range current.Annotations {
		t.Logf("k:%s, v:%s", k, v)
	}
	a1, a2 := getDiceAnnotations(current.Annotations, target.Annotations)
	if !isAnnotationsEqual(a1, a2) {
		t.Logf("diff annotations, %s/%s: %v -> %v", current.Namespace, current.Name, a1, a2)
	}
}
