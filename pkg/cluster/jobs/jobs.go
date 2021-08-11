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

package jobs

import (
	"context"
	"fmt"
	"os"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/erda-project/dice-operator/pkg/cluster/check"
	"github.com/erda-project/dice-operator/pkg/cluster/deployment"
	"github.com/erda-project/dice-operator/pkg/spec"
	"github.com/erda-project/dice-operator/pkg/utils"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	EnableAffinity    = "ENABLE_AFFINITY"
	InjectClusterInfo = "INJECT_CLUSTER_INFO"
)

func GenName(dicejobname string, clus *spec.DiceCluster) string {
	return strutil.Concat(clus.Name, "-", dicejobname)
}

func CreateAndWait(
	client kubernetes.Interface,
	dicejobs diceyml.Jobs,
	clus *spec.DiceCluster,
	ownerRefs []metav1.OwnerReference) error {
	jobs, err := buildJobs(dicejobs, clus, ownerRefs)
	if err != nil {
		return err
	}

	for _, job := range jobs {
		if _, err := client.BatchV1().Jobs(clus.Namespace).Create(context.Background(), &job, metav1.CreateOptions{}); err != nil {
			// 如果 job 已经存在, 则直接返回(视为成功, 因为已经运行过了这个job)
			if errors.IsAlreadyExists(err) {
				return nil
			}
			return err
		}
		ctx, _ := context.WithTimeout(context.Background(), 300*time.Second)
		if err := check.UntilJobFinished(ctx, client, clus.Namespace, job.Name); err != nil {
			return err
		}
	}
	return nil
}

func buildJobs(dicejobs diceyml.Jobs, clus *spec.DiceCluster, ownerRefs []metav1.OwnerReference) ([]batchv1.Job, error) {
	jobs := []batchv1.Job{}
	for name, j := range dicejobs {
		vols, volmounts, err := volumes(j)
		if err != nil {
			return nil, err
		}
		k8sjob := batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:            GenName(name, clus),
				Namespace:       clus.Namespace,
				OwnerReferences: ownerRefs,
			},
			Spec: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name,
						Namespace: clus.Namespace,
					},
					Spec: corev1.PodSpec{
						ServiceAccountName: utils.GenSAName(""),
						RestartPolicy:      "Never",
						Containers: []corev1.Container{{
							Name:            name,
							Env:             composeEnvFromDiceJob(j),
							Image:           j.Image,
							ImagePullPolicy: corev1.PullAlways,
							Command: map[bool][]string{
								true:  {"/bin/sh", "-c", j.Cmd},
								false: nil}[j.Cmd != ""],
							VolumeMounts: volmounts,
						}},
						ImagePullSecrets: []corev1.LocalObjectReference{{Name: "aliyun-registry"}},
						Volumes:          vols,
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
						HostAliases: utils.ConvertToHostAlias(j.Hosts),
					},
				},
			},
		}
		if os.Getenv(EnableAffinity) != "false" {
			k8sjob.Spec.Template.Spec.Affinity = ComposeAffinity()
		}
		if os.Getenv(InjectClusterInfo) != "false" {
			for index, container := range k8sjob.Spec.Template.Spec.Containers {
				container.EnvFrom = deployment.EnvsFrom(clus)
				k8sjob.Spec.Template.Spec.Containers[index] = container
			}
		}
		jobs = append(jobs, k8sjob)
	}
	return jobs, nil
}

func volumes(dicejob *diceyml.Job) (volumes []corev1.Volume, volumeMounts []corev1.VolumeMount, err error) {
	var binds []diceyml.Bind
	binds, err = diceyml.ParseBinds(dicejob.Binds)
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
	return
}

func composeEnvFromDiceJob(job *diceyml.Job) []corev1.EnvVar {
	envs := []corev1.EnvVar{}
	for k, v := range job.Envs {
		envs = append(envs, corev1.EnvVar{Name: k, Value: v})
	}
	return envs
}

func ComposeAffinity() *corev1.Affinity {
	return &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Key:      "dice/platform",
							Operator: "Exists",
						},
					},
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
	}
}
