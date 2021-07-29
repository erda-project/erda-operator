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

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func GetClusterInfoConfigMap(client kubernetes.Interface, clus *DiceCluster) (map[string]string, error) {
	cmname := GetClusterInfoConfigMapName(clus)
	cm, err := client.CoreV1().ConfigMaps(clus.Namespace).Get(context.Background(), cmname, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return cm.Data, nil
}

func GetClusterInfoConfigMapName(clus *DiceCluster) string {
	if clus.Spec.ClusterinfoConfigMap != "" {
		return clus.Spec.ClusterinfoConfigMap
	}
	return "dice-cluster-info"
}

func GetAddonConfigMap(client kubernetes.Interface, clus *DiceCluster) (map[string]string, error) {
	cmname := GetAddonConfigMapName(clus)
	cm, err := client.CoreV1().ConfigMaps(clus.Namespace).Get(context.Background(), cmname, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return cm.Data, nil
}

func GetAddonConfigMapName(clus *DiceCluster) string {
	if clus.Spec.AddonConfigMap != "" {
		return clus.Spec.AddonConfigMap
	}
	return "dice-addons-info"
}
