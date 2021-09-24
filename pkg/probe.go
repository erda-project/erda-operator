// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pkg

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	erdav1beta1 "github.com/erda-project/erda-operator/api/v1beta1"
)

const (
	DefaultFailureThreshold    int32 = 9
	DefaultPeriodSeconds       int32 = 15
	DefaultTimeoutSeconds      int32 = 10
	DefaultInitialDelaySeconds int32 = 1
	DefaultSuccessThreshold    int32 = 1
)

const (
	TCPTimeoutSeconds      = DefaultTimeoutSeconds
	TCPFailureThreshold    = DefaultFailureThreshold
	TCPInitialDelaySeconds = DefaultInitialDelaySeconds
	TCPPeriodSeconds       = DefaultPeriodSeconds
	TCPSuccessThreshold    = DefaultSuccessThreshold
)

const (
	LivenessFailureThreshold    int32 = 9
	LivenessPeriodSeconds             = DefaultPeriodSeconds
	LivenessInitialDelaySeconds       = DefaultInitialDelaySeconds
	LivenessTimeoutSeconds            = DefaultTimeoutSeconds
	LivenessSuccessThreshold          = DefaultSuccessThreshold
)

const (
	ReadinessFailureThreshold    int32 = 3
	ReadinessInitialDelaySeconds int32 = 10
	ReadinessPeriodSeconds       int32 = 10
	ReadinessTimeoutSeconds            = DefaultTimeoutSeconds
	ReadinessSuccessThreshold          = DefaultSuccessThreshold
)

// composeBaseProbe compose Kubernetes k8s.io/api/core/v1 Probe pointer struct
// from erda HealthCheck
func composeBaseProbe(component *erdav1beta1.Component) *corev1.Probe {
	probe := &corev1.Probe{}
	if component.HealthCheck != nil {

		if component.HealthCheck.ExecCheck != nil {
			probe.Exec = &corev1.ExecAction{
				Command: append([]string{"/bin/sh", "-c"}, component.HealthCheck.ExecCheck.Command...),
			}
			return probe
		}
		if component.HealthCheck.HTTPCheck != nil {
			probe.HTTPGet = &corev1.HTTPGetAction{
				Path:   component.HealthCheck.HTTPCheck.Path,
				Port:   intstr.FromInt(component.HealthCheck.HTTPCheck.Port),
				Scheme: corev1.URISchemeHTTP,
			}
			return probe
		}
		if component.Network != nil && len(component.Network.ServiceDiscovery) != 0 {
			probe = &corev1.Probe{
				TimeoutSeconds:      TCPTimeoutSeconds,
				FailureThreshold:    TCPFailureThreshold,
				InitialDelaySeconds: TCPInitialDelaySeconds,
				PeriodSeconds:       TCPPeriodSeconds,
				SuccessThreshold:    TCPSuccessThreshold,
			}
			probe.TCPSocket = &corev1.TCPSocketAction{
				Port: intstr.FromInt(int(component.Network.ServiceDiscovery[0].Port)),
			}
		}
		return probe
	}
	return nil
}

func ComposeLivenessProbe(component *erdav1beta1.Component) *corev1.Probe {
	probe := composeBaseProbe(component)
	if probe != nil && probe.TCPSocket == nil {
		failureThreshold := LivenessFailureThreshold
		if component.HealthCheck.Duration/LivenessPeriodSeconds > LivenessFailureThreshold {
			failureThreshold = component.HealthCheck.Duration / LivenessPeriodSeconds
		}
		probe.InitialDelaySeconds = LivenessInitialDelaySeconds
		probe.TimeoutSeconds = LivenessTimeoutSeconds
		probe.PeriodSeconds = LivenessPeriodSeconds
		probe.SuccessThreshold = LivenessSuccessThreshold
		probe.FailureThreshold = failureThreshold
	}

	return probe
}

func ComposeReadinessProbe(component *erdav1beta1.Component) *corev1.Probe {

	probe := composeBaseProbe(component)
	if probe != nil && probe.TCPSocket == nil {
		failureThreshold := ReadinessFailureThreshold
		if component.HealthCheck.Duration/ReadinessPeriodSeconds > ReadinessFailureThreshold {
			failureThreshold = component.HealthCheck.Duration / ReadinessPeriodSeconds
		}
		probe.InitialDelaySeconds = ReadinessInitialDelaySeconds
		probe.TimeoutSeconds = ReadinessTimeoutSeconds
		probe.PeriodSeconds = ReadinessPeriodSeconds
		probe.SuccessThreshold = ReadinessSuccessThreshold
		probe.FailureThreshold = failureThreshold
	}
	return probe
}
