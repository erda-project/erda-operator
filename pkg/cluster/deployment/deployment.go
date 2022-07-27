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

package deployment

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"

	"github.com/erda-project/dice-operator/pkg/spec"
	"github.com/erda-project/dice-operator/pkg/utils"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/strutil"
	clusterutils "github.com/erda-project/dice-operator/pkg/cluster/utils"
)

const (
	EnableAffinity           = "ENABLE_AFFINITY"
	EnableSpecifiedNamespace = "ENABLE_SPECIFIED_NAMESPACE"
	EnableEtcdSecret         = "ENABLE_ETCD_SECRET"
	EtcdSecretName           = "ETCD_SECRET_NAME"
	DefaultSecretName        = "erda-etcd-client-secret"
	CPUBound                 = "cpu_bound"
	IOBound                  = "io_bound"

	ErdaClusterCredential = "erda-cluster-credential"

	EnableDatabaseTLS     = "ENABLE_DATABASE_TLS"
	DatabaseTlsSecretName = "erda-database-tls"
)

func GenName(dicesvcname string, clus *spec.DiceCluster) string {
	return strutil.Concat(clus.Name, "-", dicesvcname)
}

func ExtractDiceSvcName(deployName string) string {
	return strings.SplitN(deployName, "-", 2)[1]
}

func CreateIfNotExists(
	client kubernetes.Interface,
	dicesvcname string,
	dicesvc *diceyml.Service,
	clus *spec.DiceCluster,
	ownerRefs []metav1.OwnerReference) (*appsv1.Deployment, error) {
	generatedDeploy, err := BuildDeployment(dicesvcname, dicesvc, clus, ownerRefs)
	if err != nil {
		return nil, err
	}
	deploy, err := client.AppsV1().Deployments(clus.Namespace).Get(context.Background(), generatedDeploy.Name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return client.AppsV1().Deployments(clus.Namespace).Create(context.Background(), generatedDeploy, metav1.CreateOptions{})
	}
	return deploy, nil
}

func CreateOrUpdate(
	client kubernetes.Interface,
	dicesvcname string,
	dicesvc *diceyml.Service,
	clus *spec.DiceCluster,
	ownerRefs []metav1.OwnerReference) (*appsv1.Deployment, error) {

	generatedDeploy, err := BuildDeployment(dicesvcname, dicesvc, clus, ownerRefs)
	if err != nil {
		return nil, err
	}
	deploy, err := client.AppsV1().Deployments(clus.Namespace).Get(context.Background(), generatedDeploy.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			deploy, err = client.AppsV1().Deployments(clus.Namespace).Create(context.Background(), generatedDeploy, metav1.CreateOptions{})
		} else {
			return nil, err
		}
	} else {
		deploy, err = client.AppsV1().Deployments(clus.Namespace).Update(context.Background(), generatedDeploy, metav1.UpdateOptions{})
		if err != nil {
			if errors.IsForbidden(err) || errors.IsInvalid(err) {
				err = client.AppsV1().Deployments(clus.Namespace).Delete(context.Background(), generatedDeploy.Name, metav1.DeleteOptions{})
				if err != nil && !errors.IsNotFound(err) {
					return nil, err
				}
				deploy, err = client.AppsV1().Deployments(clus.Namespace).Create(context.Background(), generatedDeploy, metav1.CreateOptions{})
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return deploy, nil
}

func Delete(
	client kubernetes.Interface,
	dicesvcname string,
	clus *spec.DiceCluster) error {
	return client.AppsV1().Deployments(clus.Namespace).Delete(context.Background(), GenName(dicesvcname, clus), metav1.DeleteOptions{})
}

func BuildDeployment(
	dicesvcname string,
	dicesvc *diceyml.Service,
	clus *spec.DiceCluster,
	ownerRefs []metav1.OwnerReference) (*appsv1.Deployment, error) {

	replica := int32(dicesvc.Deployments.Replicas)
	livenessprobe, err := LivenessProbe(dicesvcname, dicesvc)
	if err != nil {
		return nil, err
	}
	readinessprobe, err := ReadinessProbe(dicesvcname, dicesvc)
	if err != nil {
		return nil, err
	}
	vols, volmounts, err := Volumes(dicesvc)
	if err != nil {
		return nil, err
	}
	privilegedValue := true
	v, privileged := dicesvc.Deployments.Labels["PRIVILEGED"]
	if v != "true" {
		privileged = false
	}
	affinity := []corev1.NodeSelectorRequirement{
		{
			Key:      "dice/platform",
			Operator: "Exists",
		},
	}
	if label, ok := clus.Spec.CustomAffinity[dicesvcname]; ok {
		affinity = append(affinity, corev1.NodeSelectorRequirement{
			Key:      label,
			Operator: "Exists",
		})
	}

	// set default value for revisionHistoryLimit to 3, to limit data in ETCD
	revisionHistoryLimit := int32(3)

	deploy := &appsv1.Deployment{
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
		Spec: appsv1.DeploymentSpec{
			RevisionHistoryLimit: &revisionHistoryLimit,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"dice/component":    dicesvcname,
					"dice/koperator":    "true",
					"dice/cluster-name": clus.Name,
				},
			},
			Replicas: &replica,
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
					Containers: []corev1.Container{
						{
							Name:            dicesvcname,
							Env:             Envs(dicesvcname, dicesvc, clus),
							EnvFrom:         EnvsFromWithSrv(clus, dicesvcname),
							Image:           dicesvc.Image,
							ImagePullPolicy: corev1.PullIfNotPresent,
							LivenessProbe:   &livenessprobe,
							ReadinessProbe:  &readinessprobe,
							Ports:           Ports(dicesvc),
							Resources:       Resources(dicesvc, clus),
							VolumeMounts:    volmounts,
							Command: map[bool][]string{
								true:  {"/bin/sh", "-c", dicesvc.Cmd},
								false: nil}[dicesvc.Cmd != ""],
							SecurityContext: map[bool]*corev1.SecurityContext{
								true:  {Privileged: &privilegedValue},
								false: nil}[privileged],
						},
					},
					ImagePullSecrets: []corev1.LocalObjectReference{{Name: "aliyun-registry"}},
					Volumes:          vols,
					HostNetwork:      dicesvc.Resources.Network["mode"] == "host",
					// 默认容忍 master & lb 污点，有场景存在机器复用
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
		},
	}

	if os.Getenv(EnableAffinity) != "false" {
		deploy.Spec.Template.Spec.Affinity = ComposeAffinity(affinity, dicesvcname, dicesvc)
		SetBoundLabels(CPUBound, dicesvc, deploy)
		SetBoundLabels(IOBound, dicesvc, deploy)
	}

	if deploy.Spec.Template.Spec.HostNetwork {
		deploy.Spec.Template.Spec.DNSPolicy = corev1.DNSClusterFirstWithHostNet
	} else {
		deploy.Spec.Template.Spec.DNSPolicy = corev1.DNSClusterFirst
	}

	if dicesvc.K8SSnippet != nil && dicesvc.K8SSnippet.Container != nil {
		dicesvc.K8SSnippet.Container.Name = dicesvcname
		newContainer, err := PatchContainer(deploy.Spec.Template.Spec.Containers[0], (corev1.Container)(*dicesvc.K8SSnippet.Container))
		if err != nil {
			return nil, err
		}
		deploy.Spec.Template.Spec.Containers[0] = *newContainer
	}

	// TODO: erda.yaml support mount configmap/secret to directory
	// telegraf, filebeat also need. pkg/cluster/daemonset/daemonset.go
	if strings.Contains(dicesvcname, "telegraf") {
		deploy.Spec.Template.Spec.Volumes = append(deploy.Spec.Template.Spec.Volumes, corev1.Volume{
			Name: ErdaClusterCredential,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: ErdaClusterCredential,
				},
			},
		})

		deploy.Spec.Template.Spec.Containers[0].VolumeMounts = append(deploy.Spec.Template.Spec.Containers[0].VolumeMounts,
			corev1.VolumeMount{
				Name:      ErdaClusterCredential,
				ReadOnly:  true,
				MountPath: ErdaClusterCredential,
			})
	}

	return deploy, nil
}

func EnvsFrom(clus *spec.DiceCluster) []corev1.EnvFromSource {
	clusterinfocm := spec.GetClusterInfoConfigMapName(clus)
	addoninfo := spec.GetAddonConfigMapName(clus)
	return []corev1.EnvFromSource{
		{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: clusterinfocm,
				},
			},
		},
		{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: addoninfo,
				},
			},
		},
	}
}

func EnvsFromWithSrv(clus *spec.DiceCluster, dicesvcname string) []corev1.EnvFromSource {
	r := EnvsFrom(clus)
	if len(clus.Spec.MainPlatform) == 0 {
		return r
	}
	if utils.IsPipelineEdgeEnabled() && dicesvcname == "pipeline" {
		r = append(r, corev1.EnvFromSource{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: ErdaClusterCredential,
				},
			},
		})
	}
	return r
}

func PatchContainer(originContainer, patchContainer corev1.Container) (*corev1.Container, error) {
	newContainer := originContainer
	patchBytes, err := json.Marshal(patchContainer)
	if err != nil {
		errMsg := fmt.Sprintf("marshal patch container failed: %v", err)
		logrus.Error(errMsg)
		return nil, fmt.Errorf(errMsg)
	}
	err = json.Unmarshal(patchBytes, &newContainer)
	if err != nil {
		errMsg := fmt.Sprintf("unmarshal patch container failed: %v", err)
		logrus.Error(errMsg)
		return nil, fmt.Errorf(errMsg)
	}
	return &newContainer, nil
}

func ComposeAffinity(affinity []corev1.NodeSelectorRequirement, dicesvcname string, dicesvc *diceyml.Service) *corev1.Affinity {
	newAffinity := &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{{
					MatchExpressions: affinity,
				}},
			},
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
				{
					Weight: 100,
					Preference: corev1.NodeSelectorTerm{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{
								Key:      "dice/master",
								Operator: corev1.NodeSelectorOpDoesNotExist,
							},
						},
					},
				},
				{
					Weight: 100,
					Preference: corev1.NodeSelectorTerm{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{
								Key:      "node-role.kubernetes.io/master",
								Operator: corev1.NodeSelectorOpDoesNotExist,
							},
						},
					},
				},
				{
					Weight: 80,
					Preference: corev1.NodeSelectorTerm{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{
								Key:      "dice/lb",
								Operator: corev1.NodeSelectorOpDoesNotExist,
							},
						},
					},
				},
				{
					Weight: 80,
					Preference: corev1.NodeSelectorTerm{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{
								Key:      "node-role.kubernetes.io/lb",
								Operator: corev1.NodeSelectorOpDoesNotExist,
							},
						},
					},
				},
			},
		},
		PodAntiAffinity: &corev1.PodAntiAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
				{
					Weight: 100,
					PodAffinityTerm: corev1.PodAffinityTerm{
						TopologyKey: "kubernetes.io/hostname",
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{{
								Key:      "dice/component",
								Operator: "In",
								Values:   []string{dicesvcname},
							}},
						},
					},
				},
			},
		},
	}

	preferredTerm := newAffinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution
	_, cpuOK := dicesvc.Labels[CPUBound]
	_, ioOK := dicesvc.Labels[IOBound]
	if cpuOK && ioOK {
		preferredTerm = append(preferredTerm,
			corev1.WeightedPodAffinityTerm{
				Weight: 100,
				PodAffinityTerm: corev1.PodAffinityTerm{
					LabelSelector: &metav1.LabelSelector{
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      "erda/" + CPUBound,
								Operator: metav1.LabelSelectorOpExists,
							},
							{
								Key:      "erda/" + IOBound,
								Operator: metav1.LabelSelectorOpExists,
							},
						},
					},
					TopologyKey: "kubernetes.io/hostname",
				},
			},
		)
	} else {
		preferredTerm = append(preferredTerm,
			[]corev1.WeightedPodAffinityTerm{
				{
					Weight: 50,
					PodAffinityTerm: corev1.PodAffinityTerm{
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "erda/" + CPUBound,
									Operator: metav1.LabelSelectorOpExists,
								},
							},
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
				{
					Weight: 50,
					PodAffinityTerm: corev1.PodAffinityTerm{
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "erda/" + IOBound,
									Operator: metav1.LabelSelectorOpExists,
								},
							},
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
			}...,
		)
	}
	newAffinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution = preferredTerm

	return newAffinity
}

func SetBoundLabels(key string, service *diceyml.Service, deploy *appsv1.Deployment) {
	if val, ok := service.Labels[key]; ok {
		deploy.Labels["erda/"+key] = val
		deploy.Spec.Selector.MatchLabels["erda/"+key] = val
		deploy.Spec.Template.Labels["erda/"+key] = val
	}
}

// dicesvc 中的 env 已经包括了 global env
// 额外注入ENV: DICE_COMPONENT, POD_IP, HOST_IP
func Envs(dicesvcname string, dicesvc *diceyml.Service, clus *spec.DiceCluster) []corev1.EnvVar {
	r := []corev1.EnvVar{}
	for k, v := range dicesvc.Envs {
		r = append(r, corev1.EnvVar{Name: k, Value: v})
	}

	if _, ok := dicesvc.Envs["SELF_ADDR"]; !ok && len(dicesvc.Ports) > 0 {
		defaultPort := dicesvc.Ports[0].Port
		for _, svcPort := range dicesvc.Ports {
			if svcPort.Default {
				defaultPort = svcPort.Port
			}
		}
		var namespace = metav1.NamespaceDefault
		if os.Getenv(EnableSpecifiedNamespace) != "" {
			namespace = os.Getenv(EnableSpecifiedNamespace)
		}

		r = append(r, corev1.EnvVar{
			Name:  "SELF_ADDR",
			Value: fmt.Sprintf("%s.%s.svc.cluster.local:%d", dicesvcname, namespace, defaultPort),
		})
	}

	r = append(r, corev1.EnvVar{Name: "DICE_COMPONENT", Value: dicesvcname})
	// 跟 DICE_CLUSTER_NAME 相同, 一些组件目前还是用的这个 env
	r = append(r, corev1.EnvVar{Name: "DICE_CLUSTER", ValueFrom: &corev1.EnvVarSource{
		ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: spec.GetClusterInfoConfigMapName(clus),
			},
			Key: "DICE_CLUSTER_NAME",
		},
	}})
	r = append(r, corev1.EnvVar{
		Name: "POD_IP",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "status.podIP",
			},
		}})
	r = append(r, corev1.EnvVar{
		Name: "HOST_IP",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "status.hostIP",
			},
		},
	})
	r = append(r, corev1.EnvVar{
		Name: "NODE_NAME",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "spec.nodeName",
			},
		},
	})
	r = append(r, corev1.EnvVar{
		Name: "POD_NAME",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "metadata.name",
			},
		},
	})
	r = append(r, corev1.EnvVar{
		Name: "POD_UUID",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "metadata.uid",
			},
		},
	})
	r = append(r, corev1.EnvVar{
		Name: "DICE_NAMESPACE",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "metadata.namespace",
			},
		},
	})

	r = append(r, corev1.EnvVar{
		Name:  "DICE_CPU_ORIGIN",
		Value: fmt.Sprintf("%f", max(dicesvc.Resources.CPU, dicesvc.Resources.MaxCPU)),
	}, corev1.EnvVar{
		Name:  "DICE_MEM_ORIGIN",
		Value: fmt.Sprintf("%d", maxInt(dicesvc.Resources.Mem, dicesvc.Resources.MaxMem)),
	}, corev1.EnvVar{
		Name:  "DICE_CPU_REQUEST",
		Value: fmt.Sprintf("%f", dicesvc.Resources.CPU),
	}, corev1.EnvVar{
		Name:  "DICE_MEM_REQUEST",
		Value: fmt.Sprintf("%d", dicesvc.Resources.Mem),
	}, corev1.EnvVar{
		Name:  "DICE_CPU_LIMIT",
		Value: fmt.Sprintf("%f", max(dicesvc.Resources.CPU, dicesvc.Resources.MaxCPU)),
	}, corev1.EnvVar{
		Name:  "DICE_MEM_LIMIT",
		Value: fmt.Sprintf("%d", maxInt(dicesvc.Resources.Mem, dicesvc.Resources.MaxMem)),
	})

	if os.Getenv(EnableDatabaseTLS) == "true" {
		r = append(r, corev1.EnvVar{
			Name:  "MYSQL_CACERTPATH",
			Value: fmt.Sprintf("/%s", DatabaseTlsSecretName),
		})
	}

	return r
}

func LivenessProbe(dicesvcname string, dicesvc *diceyml.Service) (corev1.Probe, error) {
	failureThreshold := 9
	if dicesvc.HealthCheck.HTTP != nil && dicesvc.HealthCheck.HTTP.Path != "" {
		if dicesvc.HealthCheck.HTTP.Duration/15 > 9 {
			failureThreshold = dicesvc.HealthCheck.HTTP.Duration / 15
		}
		return corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:   dicesvc.HealthCheck.HTTP.Path,
					Port:   intstr.FromInt(dicesvc.HealthCheck.HTTP.Port),
					Scheme: corev1.URISchemeHTTP,
				},
			},
			InitialDelaySeconds: 1,
			TimeoutSeconds:      10,
			PeriodSeconds:       15,
			SuccessThreshold:    1,
			FailureThreshold:    int32(failureThreshold),
		}, nil
	}
	if dicesvc.HealthCheck.Exec != nil && dicesvc.HealthCheck.Exec.Cmd != "" {
		if dicesvc.HealthCheck.Exec.Duration/15 > 9 {
			failureThreshold = dicesvc.HealthCheck.Exec.Duration / 15
		}
		return corev1.Probe{
			Handler: corev1.Handler{
				Exec: &corev1.ExecAction{
					Command: []string{"/bin/sh", "-c", dicesvc.HealthCheck.Exec.Cmd},
				},
			},
			InitialDelaySeconds: 1,
			TimeoutSeconds:      10,
			PeriodSeconds:       15,
			SuccessThreshold:    1,
			FailureThreshold:    int32(failureThreshold),
		}, nil
	}
	if len(dicesvc.Ports) > 0 {
		var port = dicesvc.Ports[0].Port

		return corev1.Probe{
			FailureThreshold:    9,
			InitialDelaySeconds: 1,
			PeriodSeconds:       15,
			SuccessThreshold:    1,
			Handler: corev1.Handler{
				TCPSocket: &corev1.TCPSocketAction{
					Port: intstr.FromInt(port),
				},
			},
		}, nil
	}
	return corev1.Probe{}, fmt.Errorf("Not provide http healthcheck, dice service: %s", dicesvcname)
}

func ReadinessProbe(dicesvcname string, dicesvc *diceyml.Service) (corev1.Probe, error) {
	failureThreshold := 3
	if dicesvc.HealthCheck.HTTP != nil && dicesvc.HealthCheck.HTTP.Path != "" {
		if dicesvc.HealthCheck.HTTP.Duration/10 > 3 {
			failureThreshold = dicesvc.HealthCheck.HTTP.Duration / 10
		}
		return corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:   dicesvc.HealthCheck.HTTP.Path,
					Port:   intstr.FromInt(dicesvc.HealthCheck.HTTP.Port),
					Scheme: corev1.URISchemeHTTP,
				},
			},
			InitialDelaySeconds: 10,
			TimeoutSeconds:      10,
			PeriodSeconds:       10,
			SuccessThreshold:    1,
			FailureThreshold:    int32(failureThreshold),
		}, nil
	}
	if dicesvc.HealthCheck.Exec != nil && dicesvc.HealthCheck.Exec.Cmd != "" {
		if dicesvc.HealthCheck.Exec.Duration/10 > 3 {
			failureThreshold = dicesvc.HealthCheck.Exec.Duration / 10
		}
		return corev1.Probe{
			Handler: corev1.Handler{
				Exec: &corev1.ExecAction{
					Command: []string{"/bin/sh", "-c", dicesvc.HealthCheck.Exec.Cmd},
				},
			},
			InitialDelaySeconds: 10,
			TimeoutSeconds:      10,
			PeriodSeconds:       10,
			SuccessThreshold:    1,
			FailureThreshold:    int32(failureThreshold),
		}, nil
	}
	if len(dicesvc.Ports) > 0 {
		return corev1.Probe{
			InitialDelaySeconds: 10,
			PeriodSeconds:       10,
			FailureThreshold:    3,
			SuccessThreshold:    1,
			Handler: corev1.Handler{
				TCPSocket: &corev1.TCPSocketAction{
					Port: intstr.FromInt(dicesvc.Ports[0].Port),
				},
			},
		}, nil
	}
	return corev1.Probe{}, fmt.Errorf("Not provide http healthcheck, dice service: %s", dicesvcname)
}

func Ports(dicesvc *diceyml.Service) []corev1.ContainerPort {
	r := []corev1.ContainerPort{}
	for _, port := range dicesvc.Ports {
		// TODO: better impl, 目前diceyml中的port字段没有区分 udp, tcp 的配置
		if port.Protocol == "" {
			port.Protocol = "TCP"
			port.L4Protocol = corev1.ProtocolTCP
			if port.Port == 8125 || port.Port == 8094 {
				port.Protocol = "UDP"
				port.L4Protocol = corev1.ProtocolUDP
			}
		}
		r = append(r, corev1.ContainerPort{
			ContainerPort: int32(port.Port),
			Protocol:      port.L4Protocol,
		})
	}
	return r
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
			"cpu":    resource.MustParse(fmt.Sprintf("%.fm", dicesvc.Resources.CPU*1000)),
			"memory": resource.MustParse(fmt.Sprintf("%.dMi", dicesvc.Resources.Mem)),
		},
	}
}

func Volumes(dicesvc *diceyml.Service) (volumes []corev1.Volume, volumeMounts []corev1.VolumeMount, err error) {
	var binds []diceyml.Bind
	binds, err = diceyml.ParseBinds(dicesvc.Binds)
	if err != nil {
		return
	}
	volumes = []corev1.Volume{}
	volumeMounts = []corev1.VolumeMount{}
	for i, bind := range binds {
		volumeName := fmt.Sprintf("bind-%d", i)
		volumes = append(volumes, corev1.Volume{
			Name: volumeName,
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: bind.HostPath,
				},
			},
		})

		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      volumeName,
			MountPath: bind.ContainerPath,
			ReadOnly:  bind.Type == "ro",
		})
	}
	if os.Getenv(EnableEtcdSecret) != "disable" {
		name := DefaultSecretName

		if os.Getenv(EtcdSecretName) != "" {
			name = os.Getenv(EtcdSecretName)
		}
		volumes = append(volumes, corev1.Volume{
			Name: name,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: name,
				},
			},
		})
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      name,
			ReadOnly:  true,
			MountPath: "/certs/",
		})
	}

	if os.Getenv(EnableDatabaseTLS) == "true" {
		volumes = append(volumes, corev1.Volume{
			Name: DatabaseTlsSecretName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: DatabaseTlsSecretName,
				},
			},
		})
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      DatabaseTlsSecretName,
			ReadOnly:  true,
			MountPath: fmt.Sprintf("/%s", DatabaseTlsSecretName),
		})
	}

	return
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
