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
	"fmt"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/erda-project/erda/pkg/strutil"
)

type DaemonsetListDiff struct {
	currentDaemonsets map[string]appsv1.DaemonSet
	targetDaemonsets  map[string]appsv1.DaemonSet
}

func NewDaemonsetListDiff(current, target []appsv1.DaemonSet) *DaemonsetListDiff {
	currentDaemonsets := map[string]appsv1.DaemonSet{}
	targetDaemonsets := map[string]appsv1.DaemonSet{}

	for _, ds := range current {
		currentDaemonsets[ds.Name] = ds
	}
	for _, ds := range target {
		targetDaemonsets[ds.Name] = ds
	}
	return &DaemonsetListDiff{
		currentDaemonsets: currentDaemonsets,
		targetDaemonsets:  targetDaemonsets,
	}
}

type DaemonsetListActions struct {
	AddedDaemonsets         map[string]appsv1.DaemonSet
	UpdatedDaemonsets       map[string]appsv1.DaemonSet
	DeletedDaemonsets       map[string]appsv1.DaemonSet
	UpdatedDaemonsetsForVPA map[string]appsv1.DaemonSet
}

func (a *DaemonsetListActions) String() string {
	add, update, delete := []string{}, []string{}, []string{}
	for _, ds := range a.AddedDaemonsets {
		add = append(add, ds.Name)
	}
	for _, ds := range a.UpdatedDaemonsets {
		update = append(update, ds.Name)
	}
	for _, ds := range a.DeletedDaemonsets {
		delete = append(delete, ds.Name)
	}

	return fmt.Sprintf("DaemonsetListActions: ADD: [%s], UPDATE: [%s], DELETE: [%s]",
		strutil.Join(add, ", "), strutil.Join(update, ", "), strutil.Join(delete, ", "))
}

func (d *DaemonsetListDiff) GetActions() *DaemonsetListActions {
	r := &DaemonsetListActions{}
	missingInSet1, missingInSet2, shared := diffDSSet(d.currentDaemonsets, d.targetDaemonsets)
	r.AddedDaemonsets = missingInSet1
	r.DeletedDaemonsets = missingInSet2
	// For change on turn on/off pod autoscaler, need create or delete pod autoscaler object, so take all daemonsets as need update to trigger create/delete pod autoscaler action
	r.UpdatedDaemonsetsForVPA = shared
	r.UpdatedDaemonsets = getNotEqualDSs(d.currentDaemonsets, d.targetDaemonsets, shared)
	return r
}
func diffDSSet(set1, set2 map[string]appsv1.DaemonSet) (
	missingInSet1, missingInSet2, shared map[string]appsv1.DaemonSet) {
	missingInSet1 = make(map[string]appsv1.DaemonSet)
	missingInSet2 = make(map[string]appsv1.DaemonSet)
	shared = make(map[string]appsv1.DaemonSet)

	for k, v := range set1 {
		if v2, ok := set2[k]; !ok {
			missingInSet2[k] = v
		} else {
			shared[k] = v2
		}
	}

	for k, v := range set2 {
		if _, ok := set1[k]; !ok {
			missingInSet1[k] = v
		}
	}
	return
}

func isDaemonsetEqual(ds1, ds2 appsv1.DaemonSet) bool {
	a1, a2 := getDiceAnnotations(ds1.Annotations, ds2.Annotations)
	if !isAnnotationsEqual(a1, a2) {
		logrus.Infof("diff annotations, %s/%s: %v -> %v", ds1.Namespace, ds1.Name, a1, a2)
		return false
	}

	containerSet1 := map[string]corev1.Container{}
	containerSet2 := map[string]corev1.Container{}

	for _, c := range ds1.Spec.Template.Spec.Containers {
		containerSet1[c.Name] = c
	}
	for _, c := range ds2.Spec.Template.Spec.Containers {
		containerSet2[c.Name] = c
		if _, ok := containerSet1[c.Name]; !ok {
			return false
		}
	}
	for name := range containerSet1 {
		// reuse isContainerEqual in deploymentdiff.go
		if !isContainerEqual(containerSet1[name], containerSet2[name],
			fmt.Sprintf("%s/%s", ds1.Namespace, ds1.Name), false) {
			return false
		}
	}
	return true
}

func getDiceAnnotations(a1, a2 map[string]string) (map[string]string, map[string]string) {
	target1 := make(map[string]string)
	target2 := make(map[string]string)
	for k, v := range a1 {
		if strings.HasPrefix(k, "dice") {
			target1[k] = v
		}
	}
	for k, v := range a2 {
		if strings.HasPrefix(k, "dice") {
			target2[k] = v
		}
	}
	return target1, target2
}

func isAnnotationsEqual(a1, a2 map[string]string) bool {
	if a1 == nil && a2 == nil {
		return true
	}
	if a1 == nil || a2 == nil {
		return false
	}
	return cmp.Equal(a1, a2)
}

func getNotEqualDSs(set1, set2, shared map[string]appsv1.DaemonSet) map[string]appsv1.DaemonSet {
	r := map[string]appsv1.DaemonSet{}
	for k := range shared {
		if !isDaemonsetEqual(set1[k], set2[k]) {
			r[k] = set2[k]
		}
	}
	return r
}
