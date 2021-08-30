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

package launch

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	extensions "k8s.io/api/extensions/v1beta1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/erda-project/dice-operator/pkg/cluster/check"
	"github.com/erda-project/dice-operator/pkg/cluster/daemonset"
	"github.com/erda-project/dice-operator/pkg/cluster/deployment"
	"github.com/erda-project/dice-operator/pkg/cluster/diff"
	"github.com/erda-project/dice-operator/pkg/cluster/ingress"
	"github.com/erda-project/dice-operator/pkg/cluster/service"
	cluStatus "github.com/erda-project/dice-operator/pkg/cluster/status"
	"github.com/erda-project/dice-operator/pkg/spec"
	"github.com/erda-project/dice-operator/pkg/status"
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

const (
	UpdateDaemonSet = "update DaemonSet"
	DeleteDaemonSet = "delete DaemonSet"
	AddDaemonSet    = "add DaemonSet"
	UpdateService   = "update Service"
	DeleteService   = "delete Service"
	AddService      = "add Service"
)

type Launcher struct {
	*diff.Actions
	targetspec *spec.DiceCluster
	ownerRefs  []metav1.OwnerReference
	client     kubernetes.Interface
	restclient rest.Interface
	phase      spec.ClusterPhase
}

type result struct {
	SvcName  string
	Msg      string
	Complete bool
	Phase    spec.ClusterPhase
}

type launchFunc func(c chan result, svcName string, dicSvc *diceyml.Service)

func NewLauncher(
	actions *diff.Actions,
	targetspec *spec.DiceCluster,
	ownerRefs []metav1.OwnerReference,
	client kubernetes.Interface,
	restclient rest.Interface,
	phase spec.ClusterPhase) *Launcher {
	return &Launcher{actions, targetspec, ownerRefs,
		client, restclient, phase}
}

func (l *Launcher) Launch() error {
	if len(l.AddedDaemonSet)+len(l.AddedServices)+len(l.DeletedDaemonSet)+len(l.DeletedServices)+
		len(l.UpdatedDaemonSet)+len(l.UpdatedServices) > 0 {
		logrus.Infof("launch actions: %s", l.Actions.String())
	}

	if err := l.launchTmpStuff(); err != nil {
		logrus.Errorf("launch tmp stuff err: %v", err)
		return err
	}

	var failed bool
	var errMsg string
	if err := l.MultiLaunch(l.Actions.UpdatedServices, l.launchUpdatedService, UpdateService); err != nil {
		errMsg = errMsg + fmt.Sprintf("update service err %v;", err)
		failed = true
	}
	if err := l.MultiLaunch(l.Actions.DeletedServices, l.launchDeletedService, DeleteService); err != nil {
		errMsg = errMsg + fmt.Sprintf("delete service err %v;", err)
		failed = true
	}
	if err := l.MultiLaunch(l.Actions.AddedServices, l.launchAddedService, AddService); err != nil {
		errMsg = errMsg + fmt.Sprintf("add service err %v;", err)
		failed = true
	}
	if err := l.MultiLaunch(l.Actions.UpdatedDaemonSet, l.launchUpdatedDS, UpdateDaemonSet); err != nil {
		errMsg = errMsg + fmt.Sprintf("update daemonset err %v;", err)
		failed = true
	}
	if err := l.MultiLaunch(l.Actions.DeletedDaemonSet, l.launchDeletedDS, DeleteDaemonSet); err != nil {
		errMsg = errMsg + fmt.Sprintf("delete daemonset err %v;", err)
		failed = true
	}
	if err := l.MultiLaunch(l.Actions.AddedDaemonSet, l.LaunchAddedDS, AddDaemonSet); err != nil {
		errMsg = errMsg + fmt.Sprintf("add daemonset err %v;", err)
		failed = true
	}

	if failed {
		return errors.New("launch failed " + errMsg)
	}

	return nil
}
func (l *Launcher) launchTmpStuff() error {
	if len(l.targetspec.Spec.MainPlatform) == 0 { // center cluster
		if _, err := l.client.ExtensionsV1beta1().
			Ingresses(l.targetspec.Namespace).Get(context.Background(), "collector-publish", metav1.GetOptions{}); err != nil {
			if !kerrors.IsNotFound(err) {
				return err
			}

			ing := extensions.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "collector-publish",
					Namespace:       l.targetspec.Namespace,
					OwnerReferences: l.ownerRefs,
				},
				Spec: extensions.IngressSpec{
					Rules: []extensions.IngressRule{
						{
							Host: fmt.Sprintf("%s.%s", "collector",
								l.targetspec.Spec.PlatformDomain),
							IngressRuleValue: extensions.IngressRuleValue{
								HTTP: &extensions.HTTPIngressRuleValue{
									Paths: []extensions.HTTPIngressPath{
										{
											Path: "/api/publish-items/security/status",
											Backend: extensions.IngressBackend{
												ServiceName: "dicehub",
												ServicePort: intstr.FromInt(10000),
											},
										},
										{
											Path: "/api/publish-items/erase/status",
											Backend: extensions.IngressBackend{
												ServiceName: "dicehub",
												ServicePort: intstr.FromInt(10000),
											},
										},
									},
								},
							},
						},
					},
					TLS: []extensions.IngressTLS{{
						Hosts: []string{fmt.Sprintf("%s.%s", "collector",
							l.targetspec.Spec.PlatformDomain)},
					}},
				},
			}

			if _, err := l.client.ExtensionsV1beta1().Ingresses(l.targetspec.Namespace).Create(context.Background(), &ing, metav1.CreateOptions{}); err != nil {
				return err
			}
		}

	}
	return nil

}
func (l *Launcher) MultiLaunch(components map[string]*diceyml.Service, fun launchFunc, opt string) error {
	statusAdapter := cluStatus.New(l.client, l.restclient, l.targetspec)
	count := len(components)
	c := make(chan result, count)

	for svcName, svc := range components {
		if strings.Index(opt, "delete") <= 0 {
			_ = status.UpdateComponentStatus(l.restclient, l.targetspec.Namespace, l.targetspec.Name,
				map[string]spec.ComponentStatus{svcName: spec.ComponentStatusDeploying})
		}
		go fun(c, svcName, svc)
	}

	var failedComponents []string
	for range make([]int, count) {
		result, ok := <-c
		if !ok {
			logrus.Info("no data to get.")
			break
		}

		if !result.Complete {
			failedComponents = append(failedComponents, result.SvcName)
			logrus.Errorf("%s %s failed. error: %s", opt, result.SvcName, result.Msg)
		}

		if result.Phase != "" {
			// update component phase
			_ = status.UpdateConditionAndPhase(l.restclient, l.targetspec, l.targetspec.Namespace,
				l.targetspec.Name, spec.Condition{Reason: result.Msg}, result.Phase)

			// update component status
			_ = statusAdapter.Update(l.targetspec.Name)
		}
	}

	// return error, when execute failed
	if len(failedComponents) != 0 {
		msg := fmt.Sprintf("%s: %s failed", opt, strings.Join(failedComponents, ","))
		return errors.New(msg)
	}

	return nil
}

func (l *Launcher) launchAddedService(c chan result, svcName string, diceSvc *diceyml.Service) {
	if _, err := service.CreateIfNotExists(l.client, svcName, diceSvc, l.targetspec, l.ownerRefs); err != nil {
		msg := fmt.Sprintf("Failed to deploy service: err: %v, dicesvc: %s", err, svcName)
		c <- result{svcName, msg, false, spec.ClusterPhaseFailed}
		return
	}

	if ingress.HasIngress(diceSvc) {
		if _, err := ingress.CreateIfNotExists(l.client, svcName, diceSvc, l.targetspec, l.ownerRefs); err != nil {
			msg := fmt.Sprintf("Failed to deploy ingress: dicesvc: %s, err: %v", svcName, err)
			c <- result{svcName, msg, false, spec.ClusterPhaseFailed}
			return
		}
	}

	if _, err := deployment.CreateIfNotExists(l.client, svcName, diceSvc, l.targetspec, l.ownerRefs); err != nil {
		msg := fmt.Sprintf("Failed to deploy deployment: err: %v, dicesvc: %s", err, svcName)
		c <- result{svcName, msg, false, spec.ClusterPhaseFailed}
		return
	}

	// check deployment ready, timeout 60m
	ctx, _ := context.WithTimeout(context.Background(), 60*time.Minute)
	err := check.UntilDeploymentReady(ctx, l.client, l.targetspec.Namespace, deployment.GenName(svcName, l.targetspec))
	if err == context.DeadlineExceeded {
		msg := fmt.Sprintf("dicesvc: %s, check deployment available timeout(60m)", svcName)
		c <- result{svcName, msg, false, l.phase}
		return
	}
	if err != nil {
		msg := fmt.Sprintf("Failed to deploy deployment: dicesvc: %s, err: %v", svcName, err)
		c <- result{svcName, msg, false, spec.ClusterPhaseFailed}
		return
	}

	msg := fmt.Sprintf("check %s done", svcName)
	c <- result{svcName, msg, true, l.phase}
}

func (l *Launcher) launchUpdatedService(c chan result, svcName string, diceSvc *diceyml.Service) {

	if _, err := service.CreateOrUpdate(l.client, svcName, diceSvc, l.targetspec, l.ownerRefs); err != nil {
		msg := fmt.Sprintf("Failed to update service: err: %v, dicesvc: %s", err, svcName)
		c <- result{svcName, msg, false, spec.ClusterPhaseFailed}
		return
	}

	if ingress.HasIngress(diceSvc) {
		if _, err := ingress.CreateOrUpdate(l.client, svcName, diceSvc, l.targetspec, l.ownerRefs); err != nil {
			msg := fmt.Sprintf("Failed to update ingress: dicesvc: %s, err: %v", svcName, err)
			c <- result{svcName, msg, false, spec.ClusterPhaseFailed}
			return
		}
	}

	if _, err := deployment.CreateOrUpdate(l.client, svcName, diceSvc, l.targetspec, l.ownerRefs); err != nil {
		msg := fmt.Sprintf("Failed to update deployment: dicesvc: %s, err: %v", svcName, err)
		c <- result{svcName, msg, false, spec.ClusterPhaseFailed}
		return
	}

	// check deployment ready, timeout 60m
	ctx, _ := context.WithTimeout(context.Background(), 60*time.Minute)
	err := check.UntilDeploymentReady(ctx, l.client, l.targetspec.Namespace, deployment.GenName(svcName, l.targetspec))
	if err == context.DeadlineExceeded {
		msg := fmt.Sprintf("dicesvc: %s, check deployment available timeout(60m)", svcName)
		c <- result{svcName, msg, false, l.phase}
		return
	}
	if err != nil {
		msg := fmt.Sprintf("Failed to update deployment: dicesvc: %s, err: %v", svcName, err)
		c <- result{svcName, msg, false, spec.ClusterPhaseFailed}
		return
	}

	msg := fmt.Sprintf("check %s done", svcName)
	c <- result{svcName, msg, true, l.phase}
}

func (l *Launcher) launchDeletedService(c chan result, svcName string, diceSvc *diceyml.Service) {
	if err := deployment.Delete(l.client, svcName, l.targetspec); err != nil {
		msg := fmt.Sprintf("Failed to delete deployment: dicesvc: %s, err: %v", svcName, err)
		c <- result{svcName, msg, false, spec.ClusterPhaseFailed}
		return
	}

	if err := service.Delete(l.client, svcName, diceSvc, l.targetspec); err != nil {
		msg := fmt.Sprintf("Failed to delete service: dicesvc: %s, err: %v", svcName, err)
		c <- result{svcName, msg, false, spec.ClusterPhaseFailed}
		return
	}
	if err := ingress.Delete(l.client, svcName, l.targetspec); err != nil {
		msg := fmt.Sprintf("Failed to delete ingress: dicesvc: %s, err: %v", svcName, err)
		c <- result{svcName, msg, false, spec.ClusterPhaseFailed}
		return
	}

	msg := fmt.Sprintf("check %s done", svcName)
	c <- result{svcName, msg, true, ""}
}

func (l *Launcher) LaunchAddedDS(c chan result, svcName string, diceSvc *diceyml.Service) {
	if _, err := daemonset.CreateIfNotExists(l.client, svcName, diceSvc, l.targetspec, l.ownerRefs); err != nil {
		msg := fmt.Sprintf("Failed to deploy daemonset: dicesvc: %s, err: %v", svcName, err)
		c <- result{svcName, msg, false, spec.ClusterPhaseFailed}
		return
	}

	// check DaemonSet ready, timeout 60m
	ctx, _ := context.WithTimeout(context.Background(), 60*time.Minute)
	err := check.UntilDaemonsetReady(ctx, l.client, l.targetspec.Namespace, daemonset.GenName(svcName, l.targetspec))
	if err == context.DeadlineExceeded {
		msg := fmt.Sprintf("Failed to deploy daemonset: dicesvc: %s, check daemonset timeout(60m)", svcName)
		c <- result{svcName, msg, false, l.phase}
		return
	}
	if err != nil {
		msg := fmt.Sprintf("Failed to deploy daemonset: dicesvc: %s, err: %v", svcName, err)
		c <- result{svcName, msg, false, spec.ClusterPhaseFailed}
		return
	}

	msg := fmt.Sprintf("check %s done", svcName)
	c <- result{svcName, msg, true, l.phase}
}

func (l *Launcher) launchUpdatedDS(c chan result, svcName string, diceSvc *diceyml.Service) {
	if _, err := daemonset.CreateOrUpdate(l.client, svcName, diceSvc, l.targetspec, l.ownerRefs); err != nil {
		msg := fmt.Sprintf("Failed to update daemonset: dicesvc: %s, err: %v", svcName, err)
		c <- result{svcName, msg, false, spec.ClusterPhaseFailed}
		return
	}

	// check daemonSet ready, timeout 60m
	ctx, _ := context.WithTimeout(context.Background(), 60*time.Minute)
	err := check.UntilDaemonsetReady(ctx, l.client, l.targetspec.Namespace, daemonset.GenName(svcName, l.targetspec))
	if err == context.DeadlineExceeded {
		msg := fmt.Sprintf("Failed to update daemonset: dicesvc: %s, check daemonset timeout(60m)", svcName)
		c <- result{svcName, msg, false, l.phase}
		return
	}
	if err != nil {
		msg := fmt.Sprintf("Failed to update daemonset: dicesvc: %s, err: %v", svcName, err)
		c <- result{svcName, msg, false, spec.ClusterPhaseFailed}
		return
	}

	msg := fmt.Sprintf("check %s done", svcName)
	c <- result{svcName, msg, true, l.phase}
}

func (l *Launcher) launchDeletedDS(c chan result, svcName string, diceSvc *diceyml.Service) {
	if err := daemonset.Delete(l.client, svcName, diceSvc, l.targetspec, l.ownerRefs); err != nil {
		msg := fmt.Sprintf("Failed to delete daemonset: dicesvc: %s, err: %v", svcName, err)
		c <- result{svcName, msg, false, spec.ClusterPhaseFailed}
		return
	}

	msg := fmt.Sprintf("check %s done", svcName)
	c <- result{svcName, msg, true, ""}
}
