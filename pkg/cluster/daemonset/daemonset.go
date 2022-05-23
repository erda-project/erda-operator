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

package daemonset

import (
	"context"
	"fmt"
	"os"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"

	"github.com/erda-project/dice-operator/pkg/cluster/deployment"
	"github.com/erda-project/dice-operator/pkg/spec"
	"github.com/erda-project/dice-operator/pkg/utils"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/strutil"
	clusterutils "github.com/erda-project/dice-operator/pkg/cluster/utils"
)

const (
	EnableAffinity        = "ENABLE_AFFINITY"
	ErdaClusterCredential = "erda-cluster-credential"
)

func GenName(dicesvcname string, clus *spec.DiceCluster) string {
	return strutil.Concat(clus.Name, "-", dicesvcname)
}

func ExtractDiceSvcName(dsName string) string {
	return strings.SplitN(dsName, "-", 2)[1]
}

func CreateIfNotExists(
	client kubernetes.Interface,
	dicesvcname string,
	dicesvc *diceyml.Service,
	clus *spec.DiceCluster,
	ownerRefs []metav1.OwnerReference) (*appsv1.DaemonSet, error) {
	generatedDS, err := BuildDaemonSet(dicesvcname, dicesvc, clus, ownerRefs)
	if err != nil {
		return nil, err
	}
	ds, err := client.AppsV1().DaemonSets(clus.Namespace).Get(context.Background(), generatedDS.Name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return client.AppsV1().DaemonSets(clus.Namespace).Create(context.Background(), generatedDS, metav1.CreateOptions{})
	}
	return ds, nil
}

func CreateOrUpdate(
	client kubernetes.Interface,
	dicesvcname string,
	dicesvc *diceyml.Service,
	clus *spec.DiceCluster,
	ownerRefs []metav1.OwnerReference) (*appsv1.DaemonSet, error) {
	generatedDS, err := BuildDaemonSet(dicesvcname, dicesvc, clus, ownerRefs)
	if err != nil {
		return nil, err
	}
	_, err = client.AppsV1().DaemonSets(clus.Namespace).Get(context.Background(), generatedDS.Name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return client.AppsV1().DaemonSets(clus.Namespace).Create(context.Background(), generatedDS, metav1.CreateOptions{})
	}
	ds, err := client.AppsV1().DaemonSets(clus.Namespace).Update(context.Background(), generatedDS, metav1.UpdateOptions{})
	if errors.IsForbidden(err) || errors.IsInvalid(err) {
		client.AppsV1().DaemonSets(clus.Namespace).Delete(context.Background(), generatedDS.Name, metav1.DeleteOptions{})
		return client.AppsV1().DaemonSets(clus.Namespace).Create(context.Background(), generatedDS, metav1.CreateOptions{})
	}
	return ds, nil
}

func Delete(
	client kubernetes.Interface,
	dicesvcname string,
	clus *spec.DiceCluster) error {
	return client.AppsV1().DaemonSets(clus.Namespace).Delete(context.Background(), GenName(dicesvcname, clus), metav1.DeleteOptions{})
}

func BuildDaemonSet(
	dicesvcname string,
	dicesvc *diceyml.Service,
	clus *spec.DiceCluster,
	ownerRefs []metav1.OwnerReference) (*appsv1.DaemonSet, error) {

	vols, volmounts, err := deployment.Volumes(dicesvc)
	if err != nil {
		return nil, err
	}

	livenessProbe, err := deployment.LivenessProbe(dicesvcname, dicesvc)
	if err != nil {
		return nil, err
	}
	readinessProbe, err := deployment.ReadinessProbe(dicesvcname, dicesvc)
	if err != nil {
		return nil, err
	}
	privilegedValue := true
	v, privileged := dicesvc.Deployments.Labels["PRIVILEGED"]
	if v != "true" {
		privileged = false
	}

	// set default value for revisionHistoryLimit to 3, to limit data in ETCD
	revisionHistoryLimit := int32(3)

	ds := &appsv1.DaemonSet{
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
		Spec: appsv1.DaemonSetSpec{
			RevisionHistoryLimit: &revisionHistoryLimit,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"dice/component":    dicesvcname,
					"dice/koperator":    "true",
					"dice/cluster-name": clus.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"dice/component":    dicesvcname,
						"dice/koperator":    "true",
						"dice/cluster-name": clus.Name,
					},
					Annotations: utils.ConvertAnnotations(dicesvc.Annotations),
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: clusterutils.GenSAName(dicesvcname),
					Containers: []corev1.Container{{
						Name:            dicesvcname,
						EnvFrom:         deployment.EnvsFrom(clus),
						Env:             deployment.Envs(dicesvcname, dicesvc, clus),
						Image:           dicesvc.Image,
						ImagePullPolicy: "IfNotPresent",
						LivenessProbe:   &livenessProbe,
						ReadinessProbe:  &readinessProbe,
						Ports:           deployment.Ports(dicesvc),
						Resources:       Resources(dicesvc, clus),
						VolumeMounts:    volmounts,
						Command: map[bool][]string{
							true:  {"/bin/sh", "-c", dicesvc.Cmd},
							false: nil}[dicesvc.Cmd != ""],
						SecurityContext: map[bool]*corev1.SecurityContext{
							true:  {Privileged: &privilegedValue},
							false: nil}[privileged],
					}},
					ImagePullSecrets: []corev1.LocalObjectReference{{Name: "aliyun-registry"}},
					Volumes:          vols,
					HostNetwork:      dicesvc.Resources.Network["mode"] == "host",
					Tolerations: []corev1.Toleration{
						{
							Key:    "node-role.kubernetes.io/master",
							Effect: corev1.TaintEffectNoSchedule,
						},
						{
							Key:    "node-role.kubernetes.io/lb",
							Effect: corev1.TaintEffectNoSchedule,
						},
					},
					HostAliases: utils.ConvertToHostAlias(dicesvc.Hosts),
				},
			},
			UpdateStrategy: appsv1.DaemonSetUpdateStrategy{
				Type: appsv1.RollingUpdateDaemonSetStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDaemonSet{
					MaxUnavailable: &intstr.IntOrString{Type: intstr.Int, IntVal: int32(5)},
				},
			},
		},
	}

	// TODO: erda.yaml support mount configmap/secret to directory
	// telegraf-platform also need. pkg/cluster/deployment/deployment.go
	if strings.Contains(dicesvcname, "fluent-bit") || strings.Contains(dicesvcname, "telegraf") {
		ds.Spec.Template.Spec.Volumes = append(ds.Spec.Template.Spec.Volumes, corev1.Volume{
			Name: ErdaClusterCredential,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: ErdaClusterCredential,
				},
			},
		})

		ds.Spec.Template.Spec.Containers[0].VolumeMounts = append(ds.Spec.Template.Spec.Containers[0].VolumeMounts,
			corev1.VolumeMount{
				Name:      ErdaClusterCredential,
				ReadOnly:  true,
				MountPath: ErdaClusterCredential,
			})
	}

	if os.Getenv(EnableAffinity) != "false" {
		ds.Spec.Template.Spec.Affinity = ComposeAffinity(dicesvcname)
	}

	if ds.Spec.Template.Spec.HostNetwork {
		ds.Spec.Template.Spec.DNSPolicy = corev1.DNSClusterFirstWithHostNet
	} else {
		ds.Spec.Template.Spec.DNSPolicy = corev1.DNSClusterFirst
	}
	return ds, nil
}

func maxInt(i, j int) int {
	if i > j {
		return i
	}
	return j
}

func Resources(dicesvc *diceyml.Service, clus *spec.DiceCluster) corev1.ResourceRequirements {
	return corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			"cpu": resource.MustParse(fmt.Sprintf("%.fm",
				max(dicesvc.Resources.CPU, dicesvc.Resources.MaxCPU)*1000)),
			"memory": resource.MustParse(fmt.Sprintf("%.dMi",
				maxInt(dicesvc.Resources.Mem, dicesvc.Resources.MaxMem))),
		},
		Requests: corev1.ResourceList{
			"cpu":    resource.MustParse("10m"),
			"memory": resource.MustParse("10Mi"),
		},
	}
}

func ComposeAffinity(svcName string) *corev1.Affinity {

	affinity := &corev1.Affinity{}

	kubeEdgeComponent := []string{
		"telegraf-edge",
		"telegraf-app-edge",
	}

	diceDSComponent := []string{
		"telegraf",
		"telegraf-app",
	}

	var requirements []corev1.NodeSelectorRequirement

	for _, com := range kubeEdgeComponent {
		if com == svcName {
			requirements = append(requirements, corev1.NodeSelectorRequirement{
				Key:      "node.kubernetes.io/edge",
				Operator: "Exists",
			})
		}
	}

	for _, com := range diceDSComponent {
		if com == svcName {
			requirements = append(requirements, corev1.NodeSelectorRequirement{
				Key:      "node.kubernetes.io/edge",
				Operator: "DoesNotExist",
			})
		}
	}

	if len(requirements) > 0 {
		affinity = &corev1.Affinity{
			NodeAffinity: &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{{
						MatchExpressions: requirements,
					}},
				},
			},
		}
	}

	return affinity
}

func max(i float64, j ...float64) float64 {
	maxv := i
	for _, j_ := range j {
		if j_ > maxv {
			maxv = j_
		}
	}
	return maxv
}
