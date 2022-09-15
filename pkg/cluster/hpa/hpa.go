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

package hpa

import (
	"context"
	"fmt"
	"math"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
	"k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/erda-project/dice-operator/pkg/cluster/vpa"
	"github.com/erda-project/dice-operator/pkg/spec"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	DefaultHPAScaleFactor                int32  = 10
	DefaultHPAAverageUtilization         int32  = 85
	DefaultHPAStabilizationWindowSeconds int32  = 300
	DefaultHPAScaleStepSize              int32  = 2
	DefaultHPAScaleStepPercent           int32  = 50
	DefaultHPAScalePeriodSeconds         int32  = 30
	EnvHPAScaleFactor                    string = "ERDA_HPA_SCALE_FACTOR"
	EnvHPAMaxLimitsToRequestRatio        string = "ERDA_HPA_LIMIT_REQUEST_RATIO"
	DefaultMaxLimitsToRequestRatio       int    = 5

	// DefaultScalingPolicySelect selects the policy with the highest possible change.
	DefaultScalingPolicySelect v2beta2.ScalingPolicySelect = "Max"
	// HPAPodsScalingPolicy is a policy used to specify a change in absolute number of pods.
	HPAPodsScalingPolicy v2beta2.HPAScalingPolicyType = "Pods"
	// HPAPercentScalingPolicy is a policy used to specify a relative amount of change with respect to
	// the current number of pods.
	HPAPercentScalingPolicy v2beta2.HPAScalingPolicyType = "Percent"
)

func GenName(dicesvcname string, clus *spec.DiceCluster) string {
	return strutil.Concat(clus.Name, "-", dicesvcname)
}

func getHPAScaleFactorFromEnv() int32 {
	hpaScaleFactorStr := os.Getenv(EnvHPAScaleFactor)
	if hpaScaleFactorStr == "" {
		return DefaultHPAScaleFactor
	}

	hpaScaleFactor, err := strconv.Atoi(hpaScaleFactorStr)
	if err != nil {
		logrus.Errorf("get hpaScaleFactor form Env %s failed: %v, use default scalef actor: %v", EnvHPAScaleFactor, err, DefaultHPAScaleFactor)
		return DefaultHPAScaleFactor
	}

	return int32(hpaScaleFactor)
}

func CreateIfNotExists(
	client kubernetes.Interface,
	dicesvcname string,
	dicesvc *diceyml.Service,
	clus *spec.DiceCluster,
	ownerRefs []metav1.OwnerReference) (*v2beta2.HorizontalPodAutoscaler, error) {

	generatedHPA := BuildHPA(dicesvcname, dicesvc, clus, ownerRefs)

	hpa, err := client.AutoscalingV2beta2().HorizontalPodAutoscalers(clus.Namespace).Get(context.Background(), generatedHPA.Name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		if !isSuitableForCreateHPA(dicesvc) {
			return nil, fmt.Errorf("service %s with high limit to request ratio, not suitable for create hpa, please adjust its resource", dicesvcname)
		}
		return client.AutoscalingV2beta2().HorizontalPodAutoscalers(clus.Namespace).Create(context.Background(), generatedHPA, metav1.CreateOptions{})
	}
	return hpa, nil
}

func ListHPAInNamespace(client kubernetes.Interface, clus *spec.DiceCluster) (*v2beta2.HorizontalPodAutoscalerList, error) {
	hpaList, err := client.AutoscalingV2beta2().HorizontalPodAutoscalers(clus.Namespace).List(context.Background(),
		metav1.ListOptions{LabelSelector: "dice/koperator=true," + fmt.Sprintf("dice/cluster-name=%s", clus.Name)})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return hpaList, nil
}

func CreateOrUpdate(
	client kubernetes.Interface,
	dicesvcname string,
	dicesvc *diceyml.Service,
	clus *spec.DiceCluster,
	ownerRefs []metav1.OwnerReference) (*v2beta2.HorizontalPodAutoscaler, error) {

	generatedHPA := BuildHPA(dicesvcname, dicesvc, clus, ownerRefs)

	hpa, err := client.AutoscalingV2beta2().HorizontalPodAutoscalers(clus.Namespace).Get(context.Background(), generatedHPA.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			if !isSuitableForCreateHPA(dicesvc) {
				return nil, fmt.Errorf("service %s with high limit to request ratio, not suitable for create hpa, please adjust its resource", dicesvcname)
			}
			hpa, err = client.AutoscalingV2beta2().HorizontalPodAutoscalers(clus.Namespace).Create(context.Background(), generatedHPA, metav1.CreateOptions{})
		} else {
			return nil, err
		}
	} else {
		hpa, err = client.AutoscalingV2beta2().HorizontalPodAutoscalers(clus.Namespace).Update(context.Background(), generatedHPA, metav1.UpdateOptions{})
		if err != nil {
			if errors.IsForbidden(err) || errors.IsInvalid(err) {
				err = client.AutoscalingV2beta2().HorizontalPodAutoscalers(clus.Namespace).Delete(context.Background(), generatedHPA.Name, metav1.DeleteOptions{})
				if err != nil && !errors.IsNotFound(err) {
					return nil, err
				}
				hpa, err = client.AutoscalingV2beta2().HorizontalPodAutoscalers(clus.Namespace).Create(context.Background(), generatedHPA, metav1.CreateOptions{})
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return hpa, nil
}
func Get(
	client kubernetes.Interface,
	dicesvcname string,
	clus *spec.DiceCluster) (*v2beta2.HorizontalPodAutoscaler, error) {

	hpaName := GenName(dicesvcname, clus)
	hpa, err := client.AutoscalingV2beta2().HorizontalPodAutoscalers(clus.Namespace).Get(context.Background(), hpaName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, fmt.Errorf("not found hpa %s/%s", clus.Namespace, hpaName)
		} else {
			return nil, err
		}
	}
	return hpa, nil
}

func Delete(
	client kubernetes.Interface,
	dicesvcname string,
	clus *spec.DiceCluster) error {
	err := client.AutoscalingV2beta2().HorizontalPodAutoscalers(clus.Namespace).Delete(context.Background(), GenName(dicesvcname, clus), metav1.DeleteOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	return nil
}

func isSuitableForCreateHPA(dicesvc *diceyml.Service) bool {
	ratio := DefaultMaxLimitsToRequestRatio
	rat := os.Getenv(EnvHPAMaxLimitsToRequestRatio)
	if rat != "" {
		ratNum, err := strconv.Atoi(rat)
		if err != nil {
			logrus.Warnf("get limits to request ratio form Env %s failed: %v, use default limits to request ratio: %v", EnvHPAMaxLimitsToRequestRatio, err, DefaultMaxLimitsToRequestRatio)
			ratio = DefaultMaxLimitsToRequestRatio
		} else {
			ratio = ratNum
		}
	}

	cpu := math.Max(dicesvc.Resources.CPU, dicesvc.Resources.MaxCPU) * 1000
	memory := vpa.MaxInt(dicesvc.Resources.Mem, dicesvc.Resources.MaxMem)

	cpuMax := math.Max(dicesvc.Resources.CPU, dicesvc.Resources.MaxCPU) * 1000
	memoryMax := vpa.MaxInt(dicesvc.Resources.Mem, dicesvc.Resources.MaxMem)

	if (memoryMax/memory > ratio) || (int(cpuMax)/int(cpu) > ratio) {
		return false
	}
	return true
}

func BuildHPA(
	dicesvcname string,
	dicesvc *diceyml.Service,
	clus *spec.DiceCluster,
	ownerRefs []metav1.OwnerReference) *v2beta2.HorizontalPodAutoscaler {

	replica := int32(dicesvc.Deployments.Replicas)
	targetAverageUtilization := DefaultHPAAverageUtilization
	stabilizationWindowSeconds := DefaultHPAStabilizationWindowSeconds
	selectPolicy := DefaultScalingPolicySelect
	hpaScaleFactor := getHPAScaleFactorFromEnv()

	// adjust max replicas avoid too much replicas
	switch {
	case replica <= DefaultHPAScaleFactor/5:
		hpaScaleFactor = 5
	case replica < DefaultHPAScaleFactor/2:
		hpaScaleFactor = 3
	case replica >= DefaultHPAScaleFactor/2:
		hpaScaleFactor = 2
	}

	hpa := &v2beta2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:            GenName(dicesvcname, clus),
			Namespace:       clus.Namespace,
			OwnerReferences: ownerRefs,
			Labels: map[string]string{
				"dice/component":    dicesvcname,
				"dice/koperator":    "true",
				"dice/cluster-name": clus.Name,
			},
			Annotations: dicesvc.Annotations,
		},
		Spec: v2beta2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: v2beta2.CrossVersionObjectReference{
				Kind:       "Deployment",
				Name:       GenName(dicesvcname, clus),
				APIVersion: "apps/v1",
			},
			MinReplicas: &replica,
			MaxReplicas: hpaScaleFactor * replica,
			Metrics: []v2beta2.MetricSpec{
				{
					Type: v2beta2.ResourceMetricSourceType,
					Resource: &v2beta2.ResourceMetricSource{
						Name: corev1.ResourceCPU,
						Target: v2beta2.MetricTarget{
							Type:               v2beta2.UtilizationMetricType,
							AverageUtilization: &targetAverageUtilization,
						},
					},
				},
				{
					Type: v2beta2.ResourceMetricSourceType,
					Resource: &v2beta2.ResourceMetricSource{
						Name: corev1.ResourceMemory,
						Target: v2beta2.MetricTarget{
							Type:               v2beta2.UtilizationMetricType,
							AverageUtilization: &targetAverageUtilization,
						},
					},
				},
			},
			Behavior: &v2beta2.HorizontalPodAutoscalerBehavior{
				ScaleUp: &v2beta2.HPAScalingRules{
					StabilizationWindowSeconds: &stabilizationWindowSeconds,
					SelectPolicy:               &selectPolicy,
					Policies: []v2beta2.HPAScalingPolicy{
						{
							Type:          HPAPodsScalingPolicy,
							Value:         DefaultHPAScaleStepSize,
							PeriodSeconds: DefaultHPAScalePeriodSeconds,
						},
						{
							Type:          HPAPercentScalingPolicy,
							Value:         DefaultHPAScaleStepPercent,
							PeriodSeconds: DefaultHPAScalePeriodSeconds,
						},
					},
				},
				ScaleDown: &v2beta2.HPAScalingRules{
					StabilizationWindowSeconds: &stabilizationWindowSeconds,
					SelectPolicy:               &selectPolicy,
					Policies: []v2beta2.HPAScalingPolicy{
						{
							Type:          HPAPodsScalingPolicy,
							Value:         DefaultHPAScaleStepSize,
							PeriodSeconds: DefaultHPAScalePeriodSeconds,
						},
						{
							Type:          HPAPercentScalingPolicy,
							Value:         DefaultHPAScaleStepPercent,
							PeriodSeconds: DefaultHPAScalePeriodSeconds,
						},
					},
				},
			},
		},
	}

	return hpa
}
