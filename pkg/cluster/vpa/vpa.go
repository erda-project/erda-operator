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

package vpa

import (
	"context"
	"fmt"
	"math"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
	autoscaling "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	autoscalingv1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"

	"github.com/erda-project/dice-operator/pkg/spec"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	DefaultVPAUpdateMode autoscalingv1.UpdateMode = "Auto"

	DefaultVPAUpdateMinAvailableReplicas int32   = 1
	DefaultVPAUMaxAllowedCPU             float64 = 8
	DefaultVPAUMaxAllowedMemory          int32   = 32
	DefaultVPAScaleFactor                int     = 5
	EnvVPAScaleFactor                    string  = "ERDA_VPA_SCALE_FACTOR"
)

func GenName(dicesvcname string, clus *spec.DiceCluster) string {
	return strutil.Concat(clus.Name, "-", dicesvcname)
}

func GenVPACheckpointName(dicesvcname string, clus *spec.DiceCluster) string {
	return strutil.Concat(clus.Name, "-", dicesvcname, "-", dicesvcname)
}

func ListVPAInNamespace(vpaClientSet vpa_clientset.Interface, clus *spec.DiceCluster) (*autoscalingv1.VerticalPodAutoscalerList, error) {
	vpaList, err := vpaClientSet.AutoscalingV1().VerticalPodAutoscalers(clus.Namespace).List(context.Background(),
		metav1.ListOptions{LabelSelector: "dice/koperator=true," + fmt.Sprintf("dice/cluster-name=%s", clus.Name)})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return vpaList, nil
}

func CreateIfNotExists(
	vpaClientSet vpa_clientset.Interface,
	dicesvcname string,
	dicesvc *diceyml.Service,
	clus *spec.DiceCluster,
	ownerRefs []metav1.OwnerReference,
	serviceKind string) (*autoscalingv1.VerticalPodAutoscaler, error) {
	generatedVPA := BuildVPA(dicesvcname, dicesvc, clus, ownerRefs, serviceKind)

	vpa, err := vpaClientSet.AutoscalingV1().VerticalPodAutoscalers(clus.Namespace).Get(context.Background(), generatedVPA.Name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return vpaClientSet.AutoscalingV1().VerticalPodAutoscalers(clus.Namespace).Create(context.Background(), generatedVPA, metav1.CreateOptions{})
	}
	return vpa, nil
}

func CreateOrUpdate(
	vpaClientSet vpa_clientset.Interface,
	dicesvcname string,
	dicesvc *diceyml.Service,
	clus *spec.DiceCluster,
	ownerRefs []metav1.OwnerReference,
	serviceKind string) (*autoscalingv1.VerticalPodAutoscaler, error) {

	generatedVPA := BuildVPA(dicesvcname, dicesvc, clus, ownerRefs, serviceKind)

	hpa, err := vpaClientSet.AutoscalingV1().VerticalPodAutoscalers(clus.Namespace).Get(context.Background(), generatedVPA.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			hpa, err = vpaClientSet.AutoscalingV1().VerticalPodAutoscalers(clus.Namespace).Create(context.Background(), generatedVPA, metav1.CreateOptions{})
		} else {
			return nil, err
		}
	} else {
		generatedVPA.ResourceVersion = hpa.ResourceVersion
		hpa, err = vpaClientSet.AutoscalingV1().VerticalPodAutoscalers(clus.Namespace).Update(context.Background(), generatedVPA, metav1.UpdateOptions{})
		if err != nil {
			if errors.IsForbidden(err) || errors.IsInvalid(err) {
				err = vpaClientSet.AutoscalingV1().VerticalPodAutoscalers(clus.Namespace).Delete(context.Background(), generatedVPA.Name, metav1.DeleteOptions{})
				if err != nil && !errors.IsNotFound(err) {
					return nil, err
				}
				hpa, err = vpaClientSet.AutoscalingV1().VerticalPodAutoscalers(clus.Namespace).Create(context.Background(), generatedVPA, metav1.CreateOptions{})
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return hpa, nil
}

func Delete(
	vpaClientSet vpa_clientset.Interface,
	dicesvcname string,
	clus *spec.DiceCluster) error {

	// delete vpa
	err := vpaClientSet.AutoscalingV1().VerticalPodAutoscalers(clus.Namespace).Delete(context.Background(), GenName(dicesvcname, clus), metav1.DeleteOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	// delete vpacheckpoint
	err = vpaClientSet.AutoscalingV1().VerticalPodAutoscalerCheckpoints(clus.Namespace).Delete(context.Background(), GenVPACheckpointName(dicesvcname, clus), metav1.DeleteOptions{})
	if err != nil {
		logrus.Errorf("delete vpacheckpoint %s/%s failed: %v", clus.Namespace, GenVPACheckpointName(dicesvcname, clus), err)
	}
	return nil
}

func getVpaScaleFactorFromEnv() int {
	vpaScaleFactorStr := os.Getenv(EnvVPAScaleFactor)
	if vpaScaleFactorStr == "" {
		return DefaultVPAScaleFactor
	}

	vpaScaleFactor, err := strconv.Atoi(vpaScaleFactorStr)
	if err != nil {
		logrus.Errorf("get vpaScaleFactor form Env %s failed: %v, use default scale factor: %v", EnvVPAScaleFactor, err, DefaultVPAScaleFactor)
		return DefaultVPAScaleFactor
	}

	return vpaScaleFactor
}

func BuildVPA(
	dicesvcname string,
	dicesvc *diceyml.Service,
	clus *spec.DiceCluster,
	ownerRefs []metav1.OwnerReference,
	serviceKind string) *autoscalingv1.VerticalPodAutoscaler {

	minReplicas := DefaultVPAUpdateMinAvailableReplicas
	updateMode := DefaultVPAUpdateMode

	cpuMax := math.Max(dicesvc.Resources.CPU, dicesvc.Resources.MaxCPU) * 1000
	memoryMax := MaxInt(dicesvc.Resources.Mem, dicesvc.Resources.MaxMem)
	vpaFactor := getVpaScaleFactorFromEnv()

	vpa := &autoscalingv1.VerticalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			Kind:       "VerticalPodAutoscaler",
			APIVersion: "autoscaling.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      GenName(dicesvcname, clus),
			Namespace: clus.Namespace,
			Labels: map[string]string{
				"dice/component":    dicesvcname,
				"dice/koperator":    "true",
				"dice/cluster-name": clus.Name,
			},
			Annotations:     dicesvc.Annotations,
			OwnerReferences: ownerRefs,
		},
		Spec: autoscalingv1.VerticalPodAutoscalerSpec{
			TargetRef: &autoscaling.CrossVersionObjectReference{
				Kind:       serviceKind,
				Name:       GenName(dicesvcname, clus),
				APIVersion: "apps/v1",
			},
			UpdatePolicy: &autoscalingv1.PodUpdatePolicy{
				UpdateMode:  &updateMode,
				MinReplicas: &minReplicas,
			},
			ResourcePolicy: &autoscalingv1.PodResourcePolicy{
				ContainerPolicies: []autoscalingv1.ContainerResourcePolicy{
					{
						ContainerName: "*",
						MinAllowed: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse(fmt.Sprintf("%.fm", dicesvc.Resources.CPU*1000)),
							corev1.ResourceMemory: resource.MustParse(fmt.Sprintf("%.dMi", dicesvc.Resources.Mem)),
						},
						MaxAllowed: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse(fmt.Sprintf("%.fm", math.Min(float64(vpaFactor)*cpuMax, DefaultVPAUMaxAllowedCPU*1000))),
							corev1.ResourceMemory: resource.MustParse(fmt.Sprintf("%.dMi", minInt(int(vpaFactor)*memoryMax, int(DefaultVPAUMaxAllowedMemory)*1024))),
						},
						ControlledResources: &[]corev1.ResourceName{corev1.ResourceCPU, corev1.ResourceMemory},
					},
				},
			},
		},
	}
	return vpa
}

func MaxInt(i, j int) int {
	if i > j {
		return i
	}
	return j
}

func minInt(i, j int) int {
	if i > j {
		return j
	}
	return i
}
