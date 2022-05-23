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
	"sort"

	"github.com/google/go-cmp/cmp"

	"github.com/erda-project/dice-operator/pkg/spec"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	fdpAgent         = "fdp-agent"
	fdpDataService   = "fdp-dataservice"
	fluentbit        = "fluent-bit"
	filebeat         = "filebeat"
	soldier          = "soldier"
	telegraf         = "telegraf"
	telegrafEdge     = "telegraf-edge"
	telegrafApp      = "telegraf-app"
	telegrafAppEdge  = "telegraf-app-edge"
	telegrafPlatform = "telegraf-platform"
	ClusterAgent     = "cluster-agent"
	fdpMetaManager   = "fdp-metadata-manager"
	fdpWorkflow      = "fdp-workflow"
	collector        = "collector"
)

type SpecDiff struct {
	// spec.Dice
	diceGlobalEnvDiff    bool
	currentDiceGlobalEnv map[string]string
	targetDiceGlobalEnv  map[string]string

	diceServiceDiff     bool
	currentDiceServices map[string]*diceyml.Service
	targetDiceServices  map[string]*diceyml.Service

	// spec.AddonPlatform
	addonPlatformGlobalEnvDiff    bool
	currentAddonPlatformGlobalEnv map[string]string
	targetAddonPlatformGlobalEnv  map[string]string

	addonPlatformServiceDiff     bool
	currentAddonPlatformServices map[string]*diceyml.Service
	targetAddonPlatformServices  map[string]*diceyml.Service

	// spec.Gittar
	gittarGlobalEnvDiff    bool
	currentGittarGlobalEnv map[string]string
	targetGittarGlobalEnv  map[string]string

	gittarServiceDiff     bool
	currentGittarServices map[string]*diceyml.Service
	targetGittarServices  map[string]*diceyml.Service

	// spec.Pandora
	pandoraGlobalEnvDiff    bool
	currentPandoraGlobalEnv map[string]string
	targetPandoraGlobalEnv  map[string]string

	pandoraServiceDiff     bool
	currentPandoraServices map[string]*diceyml.Service
	targetPandoraServices  map[string]*diceyml.Service

	// spec.DiceUI
	diceUIGlobalEnvDiff    bool
	currentDiceUIGlobalEnv map[string]string
	targetDiceUIGlobalEnv  map[string]string

	diceUIServiceDiff     bool
	currentDiceUIServices map[string]*diceyml.Service
	targetDiceUIServices  map[string]*diceyml.Service

	// spec.UC
	ucGlobalEnvDiff    bool
	currentUCGlobalEnv map[string]string
	targetUCGlobalEnv  map[string]string

	ucServiceDiff     bool
	currentUCServices map[string]*diceyml.Service
	targetUCServices  map[string]*diceyml.Service

	// spec.spotAnalyzer
	spotAnalyzerGlobalEnvDiff    bool
	currentSpotAnalyzerGlobalEnv map[string]string
	targetSpotAnalyzerGlobalEnv  map[string]string

	spotAnalyzerServiceDiff     bool
	currentSpotAnalyzerServices map[string]*diceyml.Service
	targetSpotAnalyzerServices  map[string]*diceyml.Service

	// spec.spotCollector
	spotCollectorGlobalEnvDiff    bool
	currentSpotCollectorGlobalEnv map[string]string
	targetSpotCollectorGlobalEnv  map[string]string

	spotCollectorServiceDiff     bool
	currentSpotCollectorServices map[string]*diceyml.Service
	targetSpotCollectorServices  map[string]*diceyml.Service

	// spec.spotDashboard
	spotDashboardGlobalEnvDiff    bool
	currentSpotDashboardGlobalEnv map[string]string
	targetSpotDashboardGlobalEnv  map[string]string

	spotDashboardServiceDiff     bool
	currentSpotDashboardServices map[string]*diceyml.Service
	targetSpotDashboardServices  map[string]*diceyml.Service

	// spec.spotFilebeat
	spotFilebeatGlobalEnvDiff    bool
	currentSpotFilebeatGlobalEnv map[string]string
	targetSpotFilebeatGlobalEnv  map[string]string

	spotFilebeatServiceDiff     bool
	currentSpotFilebeatServices map[string]*diceyml.Service
	targetSpotFilebeatServices  map[string]*diceyml.Service

	// spec.spotStatus
	spotStatusGlobalEnvDiff    bool
	currentSpotStatusGlobalEnv map[string]string
	targetSpotStatusGlobalEnv  map[string]string

	spotStatusServiceDiff     bool
	currentSpotStatusServices map[string]*diceyml.Service
	targetSpotStatusServices  map[string]*diceyml.Service

	// spec.spotTelegraf
	spotTelegrafGlobalEnvDiff    bool
	currentSpotTelegrafGlobalEnv map[string]string
	targetSpotTelegrafGlobalEnv  map[string]string

	spotTelegrafServiceDiff     bool
	currentSpotTelegrafServices map[string]*diceyml.Service
	targetSpotTelegrafServices  map[string]*diceyml.Service

	// spec.tmc
	tmcGlobalEnvDiff    bool
	currentTmcGlobalEnv map[string]string
	targetTmcGlobalEnv  map[string]string

	tmcServiceDiff     bool
	currentTmcServices map[string]*diceyml.Service
	targetTmcServices  map[string]*diceyml.Service

	// spec.hepa
	hepaGlobalEnvDiff    bool
	currentHepaGlobalEnv map[string]string
	targetHepaGlobalEnv  map[string]string

	hepaServiceDiff     bool
	currentHepaServices map[string]*diceyml.Service
	targetHepaServices  map[string]*diceyml.Service

	// spec.spotMonitor
	spotMonitorGlobalEnvDiff    bool
	currentSpotMonitorGlobalEnv map[string]string
	targetSpotMonitorGlobalEnv  map[string]string

	spotMonitorServiceDiff     bool
	currentSpotMonitorServices map[string]*diceyml.Service
	targetSpotMonitorServices  map[string]*diceyml.Service

	// spec.fluentBit
	fluentBitGlobalEnvDiff    bool
	currentFluentBitGlobalEnv map[string]string
	targetFluentBitGlobalEnv  map[string]string

	fluentBitServiceDiff     bool
	currentFluentBitServices map[string]*diceyml.Service
	targetFluentBitServices  map[string]*diceyml.Service

	// spec.fdp
	fdpGlobalEnvDiff    bool
	currentFdpGlobalEnv map[string]string
	targetFdpGlobalEnv  map[string]string

	fdpServiceDiff     bool
	currentFdpServices map[string]*diceyml.Service
	targetFdpServices  map[string]*diceyml.Service

	// spec.fdpUI
	fdpUIGlobalEnvDiff    bool
	currentFdpUIGlobalEnv map[string]string
	targetFdpUIGlobalEnv  map[string]string

	fdpUIServiceDiff     bool
	currentFdpUIServices map[string]*diceyml.Service
	targetFdpUIServices  map[string]*diceyml.Service

	meshControllerGlobalEnvDiff    bool
	currentMeshControllerGlobalEnv map[string]string
	targetMeshControllerGlobalEnv  map[string]string

	meshControllerServiceDiff     bool
	currentMeshControllerServices map[string]*diceyml.Service
	targetMeshControllerServices  map[string]*diceyml.Service
}

type Actions struct {
	AddedServices   map[string]*diceyml.Service
	UpdatedServices map[string]*diceyml.Service
	DeletedServices map[string]*diceyml.Service

	AddedDaemonSet   map[string]*diceyml.Service
	UpdatedDaemonSet map[string]*diceyml.Service
	DeletedDaemonSet map[string]*diceyml.Service
}

func (a *Actions) EmptyAction() bool {
	return len(a.AddedServices) == 0 &&
		len(a.UpdatedServices) == 0 &&
		len(a.DeletedServices) == 0 &&
		len(a.AddedDaemonSet) == 0 &&
		len(a.UpdatedDaemonSet) == 0 &&
		len(a.DeletedDaemonSet) == 0
}

func (a *Actions) String() string {
	deploymentAdd, deploymentUpdate, deploymentDelete, daemonsetAdd, daemonsetUpdate, daemonsetDelete := []string{}, []string{}, []string{}, []string{}, []string{}, []string{}
	for name := range a.AddedServices {
		deploymentAdd = append(deploymentAdd, name)
	}
	for name := range a.UpdatedServices {
		deploymentUpdate = append(deploymentUpdate, name)
	}
	for name := range a.DeletedServices {
		deploymentDelete = append(deploymentDelete, name)
	}
	for name := range a.AddedDaemonSet {
		daemonsetAdd = append(daemonsetAdd, name)
	}
	for name := range a.UpdatedDaemonSet {
		daemonsetUpdate = append(daemonsetUpdate, name)
	}
	for name := range a.DeletedDaemonSet {
		daemonsetDelete = append(daemonsetDelete, name)
	}

	return fmt.Sprintf("deployment: ADD: [%s], UPDATE: [%s], DELETE: [%s], daemonset: ADD: [%s], UPDATE: [%s], DELETE: [%s]",
		strutil.Join(deploymentAdd, ", "),
		strutil.Join(deploymentUpdate, ", "),
		strutil.Join(deploymentDelete, ", "),
		strutil.Join(daemonsetAdd, ", "),
		strutil.Join(daemonsetUpdate, ", "),
		strutil.Join(daemonsetDelete, ", "))
}

func NewSpecDiff(current, target *spec.DiceCluster) *SpecDiff {
	r := SpecDiff{
		currentDiceGlobalEnv:           make(map[string]string),
		targetDiceGlobalEnv:            make(map[string]string),
		currentAddonPlatformGlobalEnv:  make(map[string]string),
		targetAddonPlatformGlobalEnv:   make(map[string]string),
		currentGittarGlobalEnv:         make(map[string]string),
		targetGittarGlobalEnv:          make(map[string]string),
		currentPandoraGlobalEnv:        make(map[string]string),
		targetPandoraGlobalEnv:         make(map[string]string),
		currentDiceUIGlobalEnv:         make(map[string]string),
		targetDiceUIGlobalEnv:          make(map[string]string),
		currentUCGlobalEnv:             make(map[string]string),
		targetUCGlobalEnv:              make(map[string]string),
		currentSpotAnalyzerGlobalEnv:   make(map[string]string),
		targetSpotAnalyzerGlobalEnv:    make(map[string]string),
		currentSpotCollectorGlobalEnv:  make(map[string]string),
		targetSpotCollectorGlobalEnv:   make(map[string]string),
		currentSpotDashboardGlobalEnv:  make(map[string]string),
		targetSpotDashboardGlobalEnv:   make(map[string]string),
		currentSpotFilebeatGlobalEnv:   make(map[string]string),
		targetSpotFilebeatGlobalEnv:    make(map[string]string),
		currentSpotStatusGlobalEnv:     make(map[string]string),
		targetSpotStatusGlobalEnv:      make(map[string]string),
		currentSpotTelegrafGlobalEnv:   make(map[string]string),
		targetSpotTelegrafGlobalEnv:    make(map[string]string),
		currentTmcGlobalEnv:            make(map[string]string),
		targetTmcGlobalEnv:             make(map[string]string),
		currentHepaGlobalEnv:           make(map[string]string),
		targetHepaGlobalEnv:            make(map[string]string),
		currentSpotMonitorGlobalEnv:    make(map[string]string),
		targetSpotMonitorGlobalEnv:     make(map[string]string),
		currentFdpGlobalEnv:            make(map[string]string),
		targetFdpGlobalEnv:             make(map[string]string),
		currentFdpUIGlobalEnv:          make(map[string]string),
		targetFdpUIGlobalEnv:           make(map[string]string),
		currentMeshControllerGlobalEnv: make(map[string]string),
		targetMeshControllerGlobalEnv:  make(map[string]string),
		currentFluentBitGlobalEnv:      make(map[string]string),
		targetFluentBitGlobalEnv:       make(map[string]string),

		currentDiceServices:           make(map[string]*diceyml.Service),
		targetDiceServices:            make(map[string]*diceyml.Service),
		currentAddonPlatformServices:  make(map[string]*diceyml.Service),
		targetAddonPlatformServices:   make(map[string]*diceyml.Service),
		currentGittarServices:         make(map[string]*diceyml.Service),
		targetGittarServices:          make(map[string]*diceyml.Service),
		currentPandoraServices:        make(map[string]*diceyml.Service),
		targetPandoraServices:         make(map[string]*diceyml.Service),
		currentDiceUIServices:         make(map[string]*diceyml.Service),
		targetDiceUIServices:          make(map[string]*diceyml.Service),
		currentUCServices:             make(map[string]*diceyml.Service),
		targetUCServices:              make(map[string]*diceyml.Service),
		currentSpotAnalyzerServices:   make(map[string]*diceyml.Service),
		targetSpotAnalyzerServices:    make(map[string]*diceyml.Service),
		currentSpotCollectorServices:  make(map[string]*diceyml.Service),
		targetSpotCollectorServices:   make(map[string]*diceyml.Service),
		currentSpotDashboardServices:  make(map[string]*diceyml.Service),
		targetSpotDashboardServices:   make(map[string]*diceyml.Service),
		currentSpotFilebeatServices:   make(map[string]*diceyml.Service),
		targetSpotFilebeatServices:    make(map[string]*diceyml.Service),
		currentSpotStatusServices:     make(map[string]*diceyml.Service),
		targetSpotStatusServices:      make(map[string]*diceyml.Service),
		currentSpotTelegrafServices:   make(map[string]*diceyml.Service),
		targetSpotTelegrafServices:    make(map[string]*diceyml.Service),
		currentTmcServices:            make(map[string]*diceyml.Service),
		targetTmcServices:             make(map[string]*diceyml.Service),
		currentHepaServices:           make(map[string]*diceyml.Service),
		targetHepaServices:            make(map[string]*diceyml.Service),
		currentSpotMonitorServices:    make(map[string]*diceyml.Service),
		targetSpotMonitorServices:     make(map[string]*diceyml.Service),
		currentFdpServices:            make(map[string]*diceyml.Service),
		targetFdpServices:             make(map[string]*diceyml.Service),
		currentFdpUIServices:          make(map[string]*diceyml.Service),
		targetFdpUIServices:           make(map[string]*diceyml.Service),
		currentMeshControllerServices: make(map[string]*diceyml.Service),
		targetMeshControllerServices:  make(map[string]*diceyml.Service),
		currentFluentBitServices:      make(map[string]*diceyml.Service),
		targetFluentBitServices:       make(map[string]*diceyml.Service),
	}

	if current == nil {
		diffFromBlank(target, &r)
		if len(target.Spec.MainPlatform) > 0 {
			r.filterEdgeClusterServices()
		}
		return &r
	}
	diffDice(*diceyml.CopyObj(&current.Spec.Dice), *diceyml.CopyObj(&target.Spec.Dice), &r)
	diffAddonPlatform(*diceyml.CopyObj(&current.Spec.AddonPlatform), *diceyml.CopyObj(&target.Spec.AddonPlatform), &r)
	diffGittar(*diceyml.CopyObj(&current.Spec.Gittar), *diceyml.CopyObj(&target.Spec.Gittar), &r)
	diffPandora(*diceyml.CopyObj(&current.Spec.Pandora), *diceyml.CopyObj(&target.Spec.Pandora), &r)
	diffDiceUI(*diceyml.CopyObj(&current.Spec.DiceUI), *diceyml.CopyObj(&target.Spec.DiceUI), &r)
	diffUC(*diceyml.CopyObj(&current.Spec.UC), *diceyml.CopyObj(&target.Spec.UC), &r)
	diffSpotAnalyzer(*diceyml.CopyObj(&current.Spec.SpotAnalyzer), *diceyml.CopyObj(&target.Spec.SpotAnalyzer), &r)
	diffSpotCollector(*diceyml.CopyObj(&current.Spec.SpotCollector), *diceyml.CopyObj(&target.Spec.SpotCollector), &r)
	diffSpotDashboard(*diceyml.CopyObj(&current.Spec.SpotDashboard), *diceyml.CopyObj(&target.Spec.SpotDashboard), &r)
	diffSpotFilebeat(*diceyml.CopyObj(&current.Spec.SpotFilebeat), *diceyml.CopyObj(&target.Spec.SpotFilebeat), &r)
	diffSpotStatus(*diceyml.CopyObj(&current.Spec.SpotStatus), *diceyml.CopyObj(&target.Spec.SpotStatus), &r)
	diffSpotTelegraf(*diceyml.CopyObj(&current.Spec.SpotTelegraf), *diceyml.CopyObj(&target.Spec.SpotTelegraf), &r)
	diffTmc(*diceyml.CopyObj(&current.Spec.Tmc), *diceyml.CopyObj(&target.Spec.Tmc), &r)
	diffHepa(*diceyml.CopyObj(&current.Spec.Hepa), *diceyml.CopyObj(&target.Spec.Hepa), &r)
	diffSpotMonitor(*diceyml.CopyObj(&current.Spec.SpotMonitor), *diceyml.CopyObj(&target.Spec.SpotMonitor), &r)
	diffFdp(*diceyml.CopyObj(&current.Spec.Fdp), *diceyml.CopyObj(&target.Spec.Fdp), &r)
	diffFdpUI(*diceyml.CopyObj(&current.Spec.FdpUI), *diceyml.CopyObj(&target.Spec.FdpUI), &r)
	diffMeshController(*diceyml.CopyObj(&current.Spec.MeshController), *diceyml.CopyObj(&target.Spec.MeshController), &r)
	diffFluentBit(*diceyml.CopyObj(&current.Spec.FluentBit), *diceyml.CopyObj(&target.Spec.FluentBit), &r)
	if len(target.Spec.MainPlatform) > 0 {
		r.filterEdgeClusterServices()
	}
	return &r
}

// filterEdgeClusterServices 如果 spec 中设置了 MainPlatform, 那么只保留部署 edge cluster 所需的服务
func (d *SpecDiff) filterEdgeClusterServices() {
	edgeSvcList := []string{
		fdpAgent,
		fdpDataService,
		telegrafApp,
		telegrafAppEdge,
		fluentbit,
		filebeat,
		telegraf,
		telegrafEdge,
		soldier,
		telegrafPlatform,
		ClusterAgent,
		fdpMetaManager,
		fdpWorkflow,
	}
	sort.Strings(edgeSvcList)
	f := func(m map[string]*diceyml.Service) {
		for k := range m {
			ind := sort.SearchStrings(edgeSvcList, k)
			if ind >= len(edgeSvcList) || edgeSvcList[ind] != k {
				delete(m, k)
			}
		}
	}
	f(d.currentDiceServices)
	f(d.targetDiceServices)
	f(d.currentAddonPlatformServices)
	f(d.targetAddonPlatformServices)
	f(d.currentGittarServices)
	f(d.targetGittarServices)
	f(d.currentPandoraServices)
	f(d.targetPandoraServices)
	f(d.currentDiceUIServices)
	f(d.targetDiceUIServices)
	f(d.currentUCServices)
	f(d.targetUCServices)
	f(d.currentSpotAnalyzerServices)
	f(d.targetSpotAnalyzerServices)
	f(d.currentSpotCollectorServices)
	f(d.targetSpotCollectorServices)
	f(d.currentSpotDashboardServices)
	f(d.targetSpotDashboardServices)
	f(d.currentSpotFilebeatServices)
	f(d.targetSpotFilebeatServices)
	f(d.currentSpotStatusServices)
	f(d.targetSpotStatusServices)
	f(d.currentSpotTelegrafServices)
	f(d.targetSpotTelegrafServices)
	f(d.currentTmcServices)
	f(d.targetTmcServices)
	f(d.currentHepaServices)
	f(d.targetHepaServices)
	f(d.currentSpotMonitorServices)
	f(d.targetSpotMonitorServices)
	f(d.currentFdpServices)
	f(d.targetFdpServices)
	f(d.currentFdpUIServices)
	f(d.targetFdpUIServices)
	f(d.currentMeshControllerServices)
	f(d.targetMeshControllerServices)
	f(d.currentFluentBitServices)
	f(d.targetFluentBitServices)

}

func (d *SpecDiff) GetActions() *Actions {
	r := &Actions{
		AddedServices:    make(map[string]*diceyml.Service),
		UpdatedServices:  make(map[string]*diceyml.Service),
		DeletedServices:  make(map[string]*diceyml.Service),
		AddedDaemonSet:   make(map[string]*diceyml.Service),
		UpdatedDaemonSet: make(map[string]*diceyml.Service),
		DeletedDaemonSet: make(map[string]*diceyml.Service),
	}
	// spec.dice
	missingInSet1, missingInSet2, shared := diffServiceset(d.currentDiceServices, d.targetDiceServices)
	differentServices := getDifferentServices(d.currentDiceServices, d.targetDiceServices, shared)
	expandGlobalEnv(d.targetDiceGlobalEnv, missingInSet1)
	expandGlobalEnv(d.targetDiceGlobalEnv, missingInSet2)
	expandGlobalEnv(d.targetDiceGlobalEnv, shared)

	mergemap(r.AddedServices, missingInSet1)
	mergemap(r.DeletedServices, missingInSet2)
	if d.diceGlobalEnvDiff {
		mergemap(r.UpdatedServices, shared)
	} else {
		mergemap(r.UpdatedServices, differentServices)
	}
	// spec.addonplatform
	missingInSet1, missingInSet2, shared = diffServiceset(d.currentAddonPlatformServices, d.targetAddonPlatformServices)
	differentServices = getDifferentServices(d.currentAddonPlatformServices, d.targetAddonPlatformServices, shared)
	expandGlobalEnv(d.targetAddonPlatformGlobalEnv, missingInSet1)
	expandGlobalEnv(d.targetAddonPlatformGlobalEnv, missingInSet2)
	expandGlobalEnv(d.targetAddonPlatformGlobalEnv, shared)

	mergemap(r.AddedServices, missingInSet1)
	mergemap(r.DeletedServices, missingInSet2)
	if d.addonPlatformGlobalEnvDiff {
		mergemap(r.UpdatedServices, shared)
	} else {
		mergemap(r.UpdatedServices, differentServices)
	}
	// spec.gittar
	missingInSet1, missingInSet2, shared = diffServiceset(d.currentGittarServices, d.targetGittarServices)
	differentServices = getDifferentServices(d.currentGittarServices, d.targetGittarServices, shared)
	expandGlobalEnv(d.targetGittarGlobalEnv, missingInSet1)
	expandGlobalEnv(d.targetGittarGlobalEnv, missingInSet2)
	expandGlobalEnv(d.targetGittarGlobalEnv, shared)

	mergemap(r.AddedServices, missingInSet1)
	mergemap(r.DeletedServices, missingInSet2)
	if d.gittarGlobalEnvDiff {
		mergemap(r.UpdatedServices, shared)
	} else {
		mergemap(r.UpdatedServices, differentServices)
	}
	// spec.pandora
	missingInSet1, missingInSet2, shared = diffServiceset(d.currentPandoraServices, d.targetPandoraServices)
	differentServices = getDifferentServices(d.currentPandoraServices, d.targetPandoraServices, shared)
	expandGlobalEnv(d.targetPandoraGlobalEnv, missingInSet1)
	expandGlobalEnv(d.targetPandoraGlobalEnv, missingInSet2)
	expandGlobalEnv(d.targetPandoraGlobalEnv, shared)

	mergemap(r.AddedServices, missingInSet1)
	mergemap(r.DeletedServices, missingInSet2)
	if d.pandoraGlobalEnvDiff {
		mergemap(r.UpdatedServices, shared)
	} else {
		mergemap(r.UpdatedServices, differentServices)
	}
	// spec.diceui
	missingInSet1, missingInSet2, shared = diffServiceset(d.currentDiceUIServices, d.targetDiceUIServices)
	differentServices = getDifferentServices(d.currentDiceUIServices, d.targetDiceUIServices, shared)
	expandGlobalEnv(d.targetDiceUIGlobalEnv, missingInSet1)
	expandGlobalEnv(d.targetDiceUIGlobalEnv, missingInSet2)
	expandGlobalEnv(d.targetDiceUIGlobalEnv, shared)

	mergemap(r.AddedServices, missingInSet1)
	mergemap(r.DeletedServices, missingInSet2)
	if d.diceUIGlobalEnvDiff {
		mergemap(r.UpdatedServices, shared)
	} else {
		mergemap(r.UpdatedServices, differentServices)
	}
	// spec.uc
	missingInSet1, missingInSet2, shared = diffServiceset(d.currentUCServices, d.targetUCServices)
	differentServices = getDifferentServices(d.currentUCServices, d.targetUCServices, shared)
	expandGlobalEnv(d.targetUCGlobalEnv, missingInSet1)
	expandGlobalEnv(d.targetUCGlobalEnv, missingInSet2)
	expandGlobalEnv(d.targetUCGlobalEnv, shared)

	mergemap(r.AddedServices, missingInSet1)
	mergemap(r.DeletedServices, missingInSet2)
	if d.ucGlobalEnvDiff {
		mergemap(r.UpdatedServices, shared)
	} else {
		mergemap(r.UpdatedServices, differentServices)
	}

	// spec.spotAnalyzer
	missingInSet1, missingInSet2, shared = diffServiceset(d.currentSpotAnalyzerServices, d.targetSpotAnalyzerServices)
	differentServices = getDifferentServices(d.currentSpotAnalyzerServices, d.targetSpotAnalyzerServices, shared)
	expandGlobalEnv(d.targetSpotAnalyzerGlobalEnv, missingInSet1)
	expandGlobalEnv(d.targetSpotAnalyzerGlobalEnv, missingInSet2)
	expandGlobalEnv(d.targetSpotAnalyzerGlobalEnv, shared)

	mergemap(r.AddedServices, missingInSet1)
	mergemap(r.DeletedServices, missingInSet2)
	if d.spotAnalyzerGlobalEnvDiff {
		mergemap(r.UpdatedServices, shared)
	} else {
		mergemap(r.UpdatedServices, differentServices)
	}

	// spec.spotCollector
	missingInSet1, missingInSet2, shared = diffServiceset(d.currentSpotCollectorServices, d.targetSpotCollectorServices)
	differentServices = getDifferentServices(d.currentSpotCollectorServices, d.targetSpotCollectorServices, shared)
	expandGlobalEnv(d.targetSpotCollectorGlobalEnv, missingInSet1)
	expandGlobalEnv(d.targetSpotCollectorGlobalEnv, missingInSet2)
	expandGlobalEnv(d.targetSpotCollectorGlobalEnv, shared)

	mergemap(r.AddedServices, missingInSet1)
	mergemap(r.DeletedServices, missingInSet2)
	if d.spotCollectorGlobalEnvDiff {
		mergemap(r.UpdatedServices, shared)
	} else {
		mergemap(r.UpdatedServices, differentServices)
	}

	// spec.spotDashboard
	missingInSet1, missingInSet2, shared = diffServiceset(d.currentSpotDashboardServices, d.targetSpotDashboardServices)
	differentServices = getDifferentServices(d.currentSpotDashboardServices, d.targetSpotDashboardServices, shared)
	expandGlobalEnv(d.targetSpotDashboardGlobalEnv, missingInSet1)
	expandGlobalEnv(d.targetSpotDashboardGlobalEnv, missingInSet2)
	expandGlobalEnv(d.targetSpotDashboardGlobalEnv, shared)

	mergemap(r.AddedServices, missingInSet1)
	mergemap(r.DeletedServices, missingInSet2)
	if d.spotDashboardGlobalEnvDiff {
		mergemap(r.UpdatedServices, shared)
	} else {
		mergemap(r.UpdatedServices, differentServices)
	}

	// spec.fluentBit
	missingInSet1, missingInSet2, shared = diffServiceset(d.currentFluentBitServices, d.targetFluentBitServices)
	differentServices = getDifferentServices(d.currentFluentBitServices, d.targetFluentBitServices, shared)
	expandGlobalEnv(d.targetFluentBitGlobalEnv, missingInSet1)
	expandGlobalEnv(d.targetFluentBitGlobalEnv, missingInSet2)
	expandGlobalEnv(d.targetFluentBitGlobalEnv, shared)

	mergemap(r.AddedDaemonSet, missingInSet1)
	mergemap(r.DeletedDaemonSet, missingInSet2)
	if d.fluentBitGlobalEnvDiff {
		mergemap(r.UpdatedDaemonSet, shared)
	} else {
		mergemap(r.UpdatedDaemonSet, differentServices)
	}

	// spec.spotFilebeat (Daemonset)
	missingInSet1, missingInSet2, shared = diffServiceset(d.currentSpotFilebeatServices, d.targetSpotFilebeatServices)
	differentServices = getDifferentServices(d.currentSpotFilebeatServices, d.targetSpotFilebeatServices, shared)
	expandGlobalEnv(d.targetSpotFilebeatGlobalEnv, missingInSet1)
	expandGlobalEnv(d.targetSpotFilebeatGlobalEnv, missingInSet2)
	expandGlobalEnv(d.targetSpotFilebeatGlobalEnv, shared)

	mergemap(r.AddedDaemonSet, missingInSet1)
	mergemap(r.DeletedDaemonSet, missingInSet2)
	if d.spotFilebeatGlobalEnvDiff {
		mergemap(r.UpdatedDaemonSet, shared)
	} else {
		mergemap(r.UpdatedDaemonSet, differentServices)
	}

	// spec.spotStatus
	missingInSet1, missingInSet2, shared = diffServiceset(d.currentSpotStatusServices, d.targetSpotStatusServices)
	differentServices = getDifferentServices(d.currentSpotStatusServices, d.targetSpotStatusServices, shared)
	expandGlobalEnv(d.targetSpotStatusGlobalEnv, missingInSet1)
	expandGlobalEnv(d.targetSpotStatusGlobalEnv, missingInSet2)
	expandGlobalEnv(d.targetSpotStatusGlobalEnv, shared)

	mergemap(r.AddedServices, missingInSet1)
	mergemap(r.DeletedServices, missingInSet2)
	if d.spotStatusGlobalEnvDiff {
		mergemap(r.UpdatedServices, shared)
	} else {
		mergemap(r.UpdatedServices, differentServices)
	}

	// spec.spotTelegraf (telegraf: daemonset, telegraf-platform: deployment)
	missingInSet1, missingInSet2, shared = diffServiceset(d.currentSpotTelegrafServices, d.targetSpotTelegrafServices)
	differentServices = getDifferentServices(d.currentSpotTelegrafServices, d.targetSpotTelegrafServices, shared)
	expandGlobalEnv(d.targetSpotTelegrafGlobalEnv, missingInSet1)
	expandGlobalEnv(d.targetSpotTelegrafGlobalEnv, missingInSet2)
	expandGlobalEnv(d.targetSpotTelegrafGlobalEnv, shared)

	for name, dicesvc := range missingInSet1 {
		if name == telegraf || name == telegrafApp || name == telegrafEdge || name == telegrafAppEdge || name == fluentbit {
			mergemap(r.AddedDaemonSet, map[string]*diceyml.Service{name: dicesvc})
		} else {
			mergemap(r.AddedServices, map[string]*diceyml.Service{name: dicesvc})
		}
	}
	for name, dicesvc := range missingInSet2 {
		if name == telegraf || name == telegrafApp || name == telegrafEdge || name == telegrafAppEdge || name == fluentbit {
			mergemap(r.DeletedDaemonSet, map[string]*diceyml.Service{name: dicesvc})
		} else {
			mergemap(r.DeletedServices, map[string]*diceyml.Service{name: dicesvc})
		}
	}
	updatedSet := differentServices
	if d.spotTelegrafGlobalEnvDiff {
		updatedSet = shared
	}
	for name, dicesvc := range updatedSet {
		if name == telegraf || name == telegrafApp || name == telegrafEdge || name == telegrafAppEdge || name == fluentbit {
			mergemap(r.UpdatedDaemonSet, map[string]*diceyml.Service{name: dicesvc})
		} else {
			mergemap(r.UpdatedServices, map[string]*diceyml.Service{name: dicesvc})
		}
	}
	// spec.tmc
	missingInSet1, missingInSet2, shared = diffServiceset(d.currentTmcServices, d.targetTmcServices)
	differentServices = getDifferentServices(d.currentTmcServices, d.targetTmcServices, shared)
	expandGlobalEnv(d.targetTmcGlobalEnv, missingInSet1)
	expandGlobalEnv(d.targetTmcGlobalEnv, missingInSet2)
	expandGlobalEnv(d.targetTmcGlobalEnv, shared)

	mergemap(r.AddedServices, missingInSet1)
	mergemap(r.DeletedServices, missingInSet2)
	if d.tmcGlobalEnvDiff {
		mergemap(r.UpdatedServices, shared)
	} else {
		mergemap(r.UpdatedServices, differentServices)
	}

	// spec.hepa
	missingInSet1, missingInSet2, shared = diffServiceset(d.currentHepaServices, d.targetHepaServices)
	differentServices = getDifferentServices(d.currentHepaServices, d.targetHepaServices, shared)
	expandGlobalEnv(d.targetHepaGlobalEnv, missingInSet1)
	expandGlobalEnv(d.targetHepaGlobalEnv, missingInSet2)
	expandGlobalEnv(d.targetHepaGlobalEnv, shared)

	mergemap(r.AddedServices, missingInSet1)
	mergemap(r.DeletedServices, missingInSet2)
	if d.hepaGlobalEnvDiff {
		mergemap(r.UpdatedServices, shared)
	} else {
		mergemap(r.UpdatedServices, differentServices)
	}
	// spec.spotMonitor
	missingInSet1, missingInSet2, shared = diffServiceset(d.currentSpotMonitorServices, d.targetSpotMonitorServices)
	differentServices = getDifferentServices(d.currentSpotMonitorServices, d.targetSpotMonitorServices, shared)
	expandGlobalEnv(d.targetSpotMonitorGlobalEnv, missingInSet1)
	expandGlobalEnv(d.targetSpotMonitorGlobalEnv, missingInSet2)
	expandGlobalEnv(d.targetSpotMonitorGlobalEnv, shared)

	mergemap(r.AddedServices, missingInSet1)
	mergemap(r.DeletedServices, missingInSet2)
	if d.spotMonitorGlobalEnvDiff {
		mergemap(r.UpdatedServices, shared)
	} else {
		mergemap(r.UpdatedServices, differentServices)
	}
	// spec.fdp
	missingInSet1, missingInSet2, shared = diffServiceset(d.currentFdpServices, d.targetFdpServices)
	differentServices = getDifferentServices(d.currentFdpServices, d.targetFdpServices, shared)
	expandGlobalEnv(d.targetFdpGlobalEnv, missingInSet1)
	expandGlobalEnv(d.targetFdpGlobalEnv, missingInSet2)
	expandGlobalEnv(d.targetFdpGlobalEnv, shared)

	mergemap(r.AddedServices, missingInSet1)
	mergemap(r.DeletedServices, missingInSet2)
	if d.fdpGlobalEnvDiff {
		mergemap(r.UpdatedServices, shared)
	} else {
		mergemap(r.UpdatedServices, differentServices)
	}

	// spec.fdpUI
	missingInSet1, missingInSet2, shared = diffServiceset(d.currentFdpUIServices, d.targetFdpUIServices)
	differentServices = getDifferentServices(d.currentFdpUIServices, d.targetFdpUIServices, shared)
	expandGlobalEnv(d.targetFdpUIGlobalEnv, missingInSet1)
	expandGlobalEnv(d.targetFdpUIGlobalEnv, missingInSet2)
	expandGlobalEnv(d.targetFdpUIGlobalEnv, shared)

	mergemap(r.AddedServices, missingInSet1)
	mergemap(r.DeletedServices, missingInSet2)
	if d.fdpUIGlobalEnvDiff {
		mergemap(r.UpdatedServices, shared)
	} else {
		mergemap(r.UpdatedServices, differentServices)
	}
	// spec.meshController
	missingInSet1, missingInSet2, shared = diffServiceset(d.currentMeshControllerServices, d.targetMeshControllerServices)
	differentServices = getDifferentServices(d.currentMeshControllerServices, d.targetMeshControllerServices, shared)
	expandGlobalEnv(d.targetMeshControllerGlobalEnv, missingInSet1)
	expandGlobalEnv(d.targetMeshControllerGlobalEnv, missingInSet2)
	expandGlobalEnv(d.targetMeshControllerGlobalEnv, shared)

	mergemap(r.AddedServices, missingInSet1)
	mergemap(r.DeletedServices, missingInSet2)
	if d.meshControllerGlobalEnvDiff {
		mergemap(r.UpdatedServices, shared)
	} else {
		mergemap(r.UpdatedServices, differentServices)
	}

	// 忽略 '*-action', dice action 服务虽然写在 dice.yml 中但是不需要部署
	// 之后在 dice.yml 中删掉 '*-action' 的描述？
	for name := range r.AddedServices {
		if strutil.HasSuffixes(name, "-action") {
			delete(r.AddedServices, name)
		}
	}

	for name := range r.UpdatedServices {
		if strutil.HasSuffixes(name, "-action") {
			delete(r.UpdatedServices, name)
		}
	}

	for name := range r.DeletedServices {
		if strutil.HasSuffixes(name, "-action") {
			delete(r.DeletedServices, name)
		}
	}

	return r
}

func diffFromBlank(target *spec.DiceCluster, specdiff *SpecDiff) {
	specdiff.diceGlobalEnvDiff = true
	specdiff.addonPlatformGlobalEnvDiff = true
	specdiff.gittarGlobalEnvDiff = true
	specdiff.pandoraGlobalEnvDiff = true
	specdiff.diceUIGlobalEnvDiff = true
	specdiff.ucGlobalEnvDiff = true
	specdiff.spotAnalyzerGlobalEnvDiff = true
	specdiff.spotCollectorGlobalEnvDiff = true
	specdiff.spotDashboardGlobalEnvDiff = true
	specdiff.spotFilebeatGlobalEnvDiff = true
	specdiff.spotStatusGlobalEnvDiff = true
	specdiff.spotTelegrafGlobalEnvDiff = true
	specdiff.tmcGlobalEnvDiff = true
	specdiff.hepaGlobalEnvDiff = true
	specdiff.spotMonitorGlobalEnvDiff = true
	specdiff.fdpGlobalEnvDiff = true
	specdiff.fdpUIGlobalEnvDiff = true
	specdiff.meshControllerGlobalEnvDiff = true
	specdiff.fluentBitGlobalEnvDiff = true

	dice := diceyml.CopyObj(&target.Spec.Dice)
	addonplatform := diceyml.CopyObj(&target.Spec.AddonPlatform)
	gittar := diceyml.CopyObj(&target.Spec.Gittar)
	pandora := diceyml.CopyObj(&target.Spec.Pandora)
	diceui := diceyml.CopyObj(&target.Spec.DiceUI)
	uc := diceyml.CopyObj(&target.Spec.UC)
	spotAnalyzer := diceyml.CopyObj(&target.Spec.SpotAnalyzer)
	spotCollector := diceyml.CopyObj(&target.Spec.SpotCollector)
	spotDashboard := diceyml.CopyObj(&target.Spec.SpotDashboard)
	spotFilebeat := diceyml.CopyObj(&target.Spec.SpotFilebeat)
	spotStatus := diceyml.CopyObj(&target.Spec.SpotStatus)
	spotTelegraf := diceyml.CopyObj(&target.Spec.SpotTelegraf)
	tmc := diceyml.CopyObj(&target.Spec.Tmc)
	hepa := diceyml.CopyObj(&target.Spec.Hepa)
	spotMonitor := diceyml.CopyObj(&target.Spec.SpotMonitor)
	fdp := diceyml.CopyObj(&target.Spec.Fdp)
	fdpUI := diceyml.CopyObj(&target.Spec.FdpUI)
	meshController := diceyml.CopyObj(&target.Spec.MeshController)
	fluentBit := diceyml.CopyObj(&target.Spec.FluentBit)

	specdiff.targetDiceGlobalEnv = dice.Envs
	specdiff.targetAddonPlatformGlobalEnv = addonplatform.Envs
	specdiff.targetGittarGlobalEnv = gittar.Envs
	specdiff.targetPandoraGlobalEnv = pandora.Envs
	specdiff.targetDiceUIGlobalEnv = diceui.Envs
	specdiff.targetUCGlobalEnv = uc.Envs
	specdiff.targetSpotAnalyzerGlobalEnv = spotAnalyzer.Envs
	specdiff.targetSpotCollectorGlobalEnv = spotCollector.Envs
	specdiff.targetSpotDashboardGlobalEnv = spotDashboard.Envs
	specdiff.targetSpotFilebeatGlobalEnv = spotFilebeat.Envs
	specdiff.targetSpotStatusGlobalEnv = spotStatus.Envs
	specdiff.targetSpotTelegrafGlobalEnv = spotTelegraf.Envs
	specdiff.targetTmcGlobalEnv = tmc.Envs
	specdiff.targetHepaGlobalEnv = hepa.Envs
	specdiff.targetSpotMonitorGlobalEnv = spotMonitor.Envs
	specdiff.targetFdpGlobalEnv = fdp.Envs
	specdiff.targetFdpUIGlobalEnv = fdpUI.Envs
	specdiff.targetMeshControllerGlobalEnv = meshController.Envs
	specdiff.targetFluentBitGlobalEnv = fluentBit.Envs

	specdiff.diceServiceDiff = true
	specdiff.addonPlatformServiceDiff = true
	specdiff.gittarServiceDiff = true
	specdiff.pandoraServiceDiff = true
	specdiff.diceUIServiceDiff = true
	specdiff.ucServiceDiff = true
	specdiff.spotAnalyzerServiceDiff = true
	specdiff.spotCollectorServiceDiff = true
	specdiff.spotDashboardServiceDiff = true
	specdiff.spotFilebeatServiceDiff = true
	specdiff.spotStatusServiceDiff = true
	specdiff.spotTelegrafServiceDiff = true
	specdiff.tmcServiceDiff = true
	specdiff.hepaServiceDiff = true
	specdiff.spotMonitorServiceDiff = true
	specdiff.fdpServiceDiff = true
	specdiff.fdpUIServiceDiff = true
	specdiff.meshControllerServiceDiff = true
	specdiff.fluentBitServiceDiff = true

	specdiff.targetDiceServices = dice.Services
	specdiff.targetAddonPlatformServices = addonplatform.Services
	specdiff.targetGittarServices = gittar.Services
	specdiff.targetPandoraServices = pandora.Services
	specdiff.targetDiceUIServices = diceui.Services
	specdiff.targetUCServices = uc.Services
	specdiff.targetSpotAnalyzerServices = spotAnalyzer.Services
	specdiff.targetSpotCollectorServices = spotCollector.Services
	specdiff.targetSpotDashboardServices = spotDashboard.Services
	specdiff.targetSpotFilebeatServices = spotFilebeat.Services
	specdiff.targetSpotStatusServices = spotStatus.Services
	specdiff.targetSpotTelegrafServices = spotTelegraf.Services
	specdiff.targetTmcServices = tmc.Services
	specdiff.targetHepaServices = hepa.Services
	specdiff.targetSpotMonitorServices = spotMonitor.Services
	specdiff.targetFdpServices = fdp.Services
	specdiff.targetFdpUIServices = fdpUI.Services
	specdiff.targetMeshControllerServices = meshController.Services
	specdiff.targetFluentBitServices = fluentBit.Services

}

func diffDice(current, target diceyml.Object, specdiff *SpecDiff) {

	diffDiceGlobalEnv(current.Envs, target.Envs, specdiff)
	diffDiceServices(current.Services, target.Services, specdiff)
}

func diffAddonPlatform(current, target diceyml.Object, specdiff *SpecDiff) {
	diffAddonPlatformGlobalEnv(current.Envs, target.Envs, specdiff)
	diffAddonPlatformServices(current.Services, target.Services, specdiff)
}

func diffGittar(current, target diceyml.Object, specdiff *SpecDiff) {
	diffGittarGlobalEnv(current.Envs, target.Envs, specdiff)
	diffGittarServices(current.Services, target.Services, specdiff)
}

func diffPandora(current, target diceyml.Object, specdiff *SpecDiff) {
	diffPandoraGlobalEnv(current.Envs, target.Envs, specdiff)
	diffPandoraServices(current.Services, target.Services, specdiff)
}

func diffDiceUI(current, target diceyml.Object, specdiff *SpecDiff) {
	diffDiceUIGlobalEnv(current.Envs, target.Envs, specdiff)
	diffDiceUIServices(current.Services, target.Services, specdiff)
}

func diffUC(current, target diceyml.Object, specdiff *SpecDiff) {
	diffUCGlobalEnv(current.Envs, target.Envs, specdiff)
	diffUCServices(current.Services, target.Services, specdiff)
}

func diffSpotAnalyzer(current, target diceyml.Object, specdiff *SpecDiff) {
	diffSpotAnalyzerGlobalEnv(current.Envs, target.Envs, specdiff)
	diffSpotAnalyzerServices(current.Services, target.Services, specdiff)
}

func diffSpotCollector(current, target diceyml.Object, specdiff *SpecDiff) {
	diffSpotCollectorGlobalEnv(current.Envs, target.Envs, specdiff)
	diffSpotCollectorServices(current.Services, target.Services, specdiff)
}

func diffSpotDashboard(current, target diceyml.Object, specdiff *SpecDiff) {
	diffSpotDashboardGlobalEnv(current.Envs, target.Envs, specdiff)
	diffSpotDashboardServices(current.Services, target.Services, specdiff)
}
func diffSpotFilebeat(current, target diceyml.Object, specdiff *SpecDiff) {
	diffSpotFilebeatGlobalEnv(current.Envs, target.Envs, specdiff)
	diffSpotFilebeatServices(current.Services, target.Services, specdiff)
}
func diffSpotStatus(current, target diceyml.Object, specdiff *SpecDiff) {
	diffSpotStatusGlobalEnv(current.Envs, target.Envs, specdiff)
	diffSpotStatusServices(current.Services, target.Services, specdiff)
}
func diffSpotTelegraf(current, target diceyml.Object, specdiff *SpecDiff) {
	diffSpotTelegrafGlobalEnv(current.Envs, target.Envs, specdiff)
	diffSpotTelegrafServices(current.Services, target.Services, specdiff)
}
func diffTmc(current, target diceyml.Object, specdiff *SpecDiff) {
	diffTmcGlobalEnv(current.Envs, target.Envs, specdiff)
	diffTmcServices(current.Services, target.Services, specdiff)
}
func diffHepa(current, target diceyml.Object, specdiff *SpecDiff) {
	diffHepaGlobalEnv(current.Envs, target.Envs, specdiff)
	diffHepaServices(current.Services, target.Services, specdiff)
}
func diffSpotMonitor(current, target diceyml.Object, specdiff *SpecDiff) {
	diffSpotMonitorGlobalEnv(current.Envs, target.Envs, specdiff)
	diffSpotMonitorServices(current.Services, target.Services, specdiff)
}
func diffFdp(current, target diceyml.Object, specdiff *SpecDiff) {
	diffFdpGlobalEnv(current.Envs, target.Envs, specdiff)
	diffFdpServices(current.Services, target.Services, specdiff)
}
func diffFdpUI(current, target diceyml.Object, specdiff *SpecDiff) {
	diffFdpUIGlobalEnv(current.Envs, target.Envs, specdiff)
	diffFdpUIServices(current.Services, target.Services, specdiff)
}
func diffMeshController(current, target diceyml.Object, specdiff *SpecDiff) {
	diffMeshControllerGlobalEnv(current.Envs, target.Envs, specdiff)
	diffMeshControllerServices(current.Services, target.Services, specdiff)
}

func diffFluentBit(current, target diceyml.Object, specdiff *SpecDiff) {
	diffFluentBitGlobalEnv(current.Envs, target.Envs, specdiff)
	diffFluentBitServices(current.Services, target.Services, specdiff)
}

func diffDiceGlobalEnv(current, target map[string]string, specdiff *SpecDiff) {
	auxDiffGlobalEnv(current, target,
		&specdiff.currentDiceGlobalEnv, &specdiff.targetDiceGlobalEnv, &specdiff.diceGlobalEnvDiff)
}

func diffAddonPlatformGlobalEnv(current, target map[string]string, specdiff *SpecDiff) {
	auxDiffGlobalEnv(current, target,
		&specdiff.currentAddonPlatformGlobalEnv, &specdiff.targetAddonPlatformGlobalEnv,
		&specdiff.addonPlatformGlobalEnvDiff)
}

func diffGittarGlobalEnv(current, target map[string]string, specdiff *SpecDiff) {
	auxDiffGlobalEnv(current, target,
		&specdiff.currentGittarGlobalEnv, &specdiff.targetGittarGlobalEnv,
		&specdiff.gittarGlobalEnvDiff)
}

func diffPandoraGlobalEnv(current, target map[string]string, specdiff *SpecDiff) {
	auxDiffGlobalEnv(current, target,
		&specdiff.currentPandoraGlobalEnv, &specdiff.targetPandoraGlobalEnv,
		&specdiff.pandoraGlobalEnvDiff)
}

func diffDiceUIGlobalEnv(current, target map[string]string, specdiff *SpecDiff) {
	auxDiffGlobalEnv(current, target,
		&specdiff.currentDiceUIGlobalEnv, &specdiff.targetDiceUIGlobalEnv,
		&specdiff.diceUIGlobalEnvDiff)
}

func diffUCGlobalEnv(current, target map[string]string, specdiff *SpecDiff) {
	auxDiffGlobalEnv(current, target,
		&specdiff.currentUCGlobalEnv, &specdiff.targetUCGlobalEnv,
		&specdiff.ucGlobalEnvDiff)
}
func diffSpotAnalyzerGlobalEnv(current, target map[string]string, specdiff *SpecDiff) {
	auxDiffGlobalEnv(current, target,
		&specdiff.currentSpotAnalyzerGlobalEnv, &specdiff.targetSpotAnalyzerGlobalEnv,
		&specdiff.spotAnalyzerGlobalEnvDiff)
}
func diffSpotCollectorGlobalEnv(current, target map[string]string, specdiff *SpecDiff) {
	auxDiffGlobalEnv(current, target,
		&specdiff.currentSpotCollectorGlobalEnv, &specdiff.targetSpotCollectorGlobalEnv,
		&specdiff.spotCollectorGlobalEnvDiff)
}
func diffSpotDashboardGlobalEnv(current, target map[string]string, specdiff *SpecDiff) {
	auxDiffGlobalEnv(current, target,
		&specdiff.currentSpotDashboardGlobalEnv, &specdiff.targetSpotDashboardGlobalEnv,
		&specdiff.spotDashboardGlobalEnvDiff)
}
func diffSpotFilebeatGlobalEnv(current, target map[string]string, specdiff *SpecDiff) {
	auxDiffGlobalEnv(current, target,
		&specdiff.currentSpotFilebeatGlobalEnv, &specdiff.targetSpotFilebeatGlobalEnv,
		&specdiff.spotFilebeatGlobalEnvDiff)
}
func diffSpotStatusGlobalEnv(current, target map[string]string, specdiff *SpecDiff) {
	auxDiffGlobalEnv(current, target,
		&specdiff.currentSpotStatusGlobalEnv, &specdiff.targetSpotStatusGlobalEnv,
		&specdiff.spotStatusGlobalEnvDiff)
}
func diffSpotTelegrafGlobalEnv(current, target map[string]string, specdiff *SpecDiff) {
	auxDiffGlobalEnv(current, target,
		&specdiff.currentSpotTelegrafGlobalEnv, &specdiff.targetSpotTelegrafGlobalEnv,
		&specdiff.spotTelegrafGlobalEnvDiff)
}
func diffTmcGlobalEnv(current, target map[string]string, specdiff *SpecDiff) {
	auxDiffGlobalEnv(current, target,
		&specdiff.currentTmcGlobalEnv, &specdiff.targetTmcGlobalEnv,
		&specdiff.tmcGlobalEnvDiff)
}
func diffHepaGlobalEnv(current, target map[string]string, specdiff *SpecDiff) {
	auxDiffGlobalEnv(current, target,
		&specdiff.currentHepaGlobalEnv, &specdiff.targetHepaGlobalEnv,
		&specdiff.hepaGlobalEnvDiff)
}
func diffSpotMonitorGlobalEnv(current, target map[string]string, specdiff *SpecDiff) {
	auxDiffGlobalEnv(current, target,
		&specdiff.currentSpotMonitorGlobalEnv, &specdiff.targetSpotMonitorGlobalEnv,
		&specdiff.spotMonitorGlobalEnvDiff)
}
func diffFdpGlobalEnv(current, target map[string]string, specdiff *SpecDiff) {
	auxDiffGlobalEnv(current, target,
		&specdiff.currentFdpGlobalEnv, &specdiff.targetFdpGlobalEnv,
		&specdiff.fdpGlobalEnvDiff)
}
func diffFdpUIGlobalEnv(current, target map[string]string, specdiff *SpecDiff) {
	auxDiffGlobalEnv(current, target,
		&specdiff.currentFdpUIGlobalEnv, &specdiff.targetFdpUIGlobalEnv,
		&specdiff.fdpUIGlobalEnvDiff)
}
func diffMeshControllerGlobalEnv(current, target map[string]string, specdiff *SpecDiff) {
	auxDiffGlobalEnv(current, target,
		&specdiff.currentMeshControllerGlobalEnv, &specdiff.targetMeshControllerGlobalEnv,
		&specdiff.meshControllerGlobalEnvDiff)
}

func diffFluentBitGlobalEnv(current, target map[string]string, specdiff *SpecDiff) {
	auxDiffGlobalEnv(current, target,
		&specdiff.currentFluentBitGlobalEnv, &specdiff.currentFluentBitGlobalEnv,
		&specdiff.fluentBitGlobalEnvDiff)
}

func auxDiffGlobalEnv(current, target map[string]string, specCurrent, specTarget *map[string]string, diff *bool) {
	if len(current) != len(target) {
		*diff = true
	}
	for k, v := range current {
		if targetv, ok := target[k]; !ok || targetv != v {
			*diff = true
			break
		}
	}
	if current != nil {
		*specCurrent = current
	}
	if target != nil {
		*specTarget = target
	}
}

func diffDiceServices(current, target map[string]*diceyml.Service, specdiff *SpecDiff) {
	auxDiffServices(current, target,
		&specdiff.currentDiceServices, &specdiff.targetDiceServices,
		&specdiff.diceServiceDiff)
}

func diffAddonPlatformServices(current, target map[string]*diceyml.Service, specdiff *SpecDiff) {
	auxDiffServices(current, target,
		&specdiff.currentAddonPlatformServices, &specdiff.targetAddonPlatformServices,
		&specdiff.addonPlatformServiceDiff)
}

func diffGittarServices(current, target map[string]*diceyml.Service, specdiff *SpecDiff) {
	auxDiffServices(current, target,
		&specdiff.currentGittarServices, &specdiff.targetGittarServices,
		&specdiff.gittarServiceDiff)
}

func diffPandoraServices(current, target map[string]*diceyml.Service, specdiff *SpecDiff) {
	auxDiffServices(current, target,
		&specdiff.currentPandoraServices, &specdiff.targetPandoraServices,
		&specdiff.pandoraServiceDiff)
}
func diffDiceUIServices(current, target map[string]*diceyml.Service, specdiff *SpecDiff) {
	auxDiffServices(current, target,
		&specdiff.currentDiceUIServices, &specdiff.targetDiceUIServices,
		&specdiff.diceUIServiceDiff)
}
func diffUCServices(current, target map[string]*diceyml.Service, specdiff *SpecDiff) {
	auxDiffServices(current, target,
		&specdiff.currentUCServices, &specdiff.targetUCServices,
		&specdiff.ucServiceDiff)
}
func diffSpotAnalyzerServices(current, target map[string]*diceyml.Service, specdiff *SpecDiff) {
	auxDiffServices(current, target,
		&specdiff.currentSpotAnalyzerServices, &specdiff.targetSpotAnalyzerServices,
		&specdiff.spotAnalyzerServiceDiff)
}

func diffSpotCollectorServices(current, target map[string]*diceyml.Service, specdiff *SpecDiff) {
	auxDiffServices(current, target,
		&specdiff.currentSpotCollectorServices, &specdiff.targetSpotCollectorServices,
		&specdiff.spotCollectorServiceDiff)
}
func diffSpotDashboardServices(current, target map[string]*diceyml.Service, specdiff *SpecDiff) {
	auxDiffServices(current, target,
		&specdiff.currentSpotDashboardServices, &specdiff.targetSpotDashboardServices,
		&specdiff.spotDashboardServiceDiff)
}
func diffSpotFilebeatServices(current, target map[string]*diceyml.Service, specdiff *SpecDiff) {
	auxDiffServices(current, target,
		&specdiff.currentSpotFilebeatServices, &specdiff.targetSpotFilebeatServices,
		&specdiff.spotFilebeatServiceDiff)
}
func diffSpotStatusServices(current, target map[string]*diceyml.Service, specdiff *SpecDiff) {
	auxDiffServices(current, target,
		&specdiff.currentSpotStatusServices, &specdiff.targetSpotStatusServices,
		&specdiff.spotStatusServiceDiff)
}
func diffSpotTelegrafServices(current, target map[string]*diceyml.Service, specdiff *SpecDiff) {
	auxDiffServices(current, target,
		&specdiff.currentSpotTelegrafServices, &specdiff.targetSpotTelegrafServices,
		&specdiff.spotTelegrafServiceDiff)
}
func diffTmcServices(current, target map[string]*diceyml.Service, specdiff *SpecDiff) {
	auxDiffServices(current, target,
		&specdiff.currentTmcServices, &specdiff.targetTmcServices,
		&specdiff.tmcServiceDiff)
}
func diffHepaServices(current, target map[string]*diceyml.Service, specdiff *SpecDiff) {
	auxDiffServices(current, target,
		&specdiff.currentHepaServices, &specdiff.targetHepaServices,
		&specdiff.hepaServiceDiff)
}
func diffSpotMonitorServices(current, target map[string]*diceyml.Service, specdiff *SpecDiff) {
	auxDiffServices(current, target,
		&specdiff.currentSpotMonitorServices, &specdiff.targetSpotMonitorServices,
		&specdiff.spotMonitorServiceDiff)
}
func diffFdpServices(current, target map[string]*diceyml.Service, specdiff *SpecDiff) {
	auxDiffServices(current, target,
		&specdiff.currentFdpServices, &specdiff.targetFdpServices,
		&specdiff.fdpServiceDiff)
}
func diffFdpUIServices(current, target map[string]*diceyml.Service, specdiff *SpecDiff) {
	auxDiffServices(current, target,
		&specdiff.currentFdpUIServices, &specdiff.targetFdpUIServices,
		&specdiff.fdpUIServiceDiff)
}
func diffMeshControllerServices(current, target map[string]*diceyml.Service, specdiff *SpecDiff) {
	auxDiffServices(current, target,
		&specdiff.currentMeshControllerServices, &specdiff.targetMeshControllerServices,
		&specdiff.meshControllerServiceDiff)
}

func diffFluentBitServices(current, target map[string]*diceyml.Service, specdiff *SpecDiff) {
	auxDiffServices(current, target,
		&specdiff.currentFluentBitServices, &specdiff.targetFluentBitServices,
		&specdiff.fluentBitServiceDiff)
}

func auxDiffServices(current, target map[string]*diceyml.Service, specCurrent, specTarget *map[string]*diceyml.Service, diff *bool) {
	if len(current) != len(target) {
		*diff = true
	}
	for svcname, v := range current {
		if targetv, ok := target[svcname]; !ok || !isServiceEqual(v, targetv) {
			*diff = true
			break
		}
	}
	if current != nil {
		*specCurrent = current
	}
	if target != nil {
		*specTarget = target
	}
}

func isServiceEqual(svc1, svc2 *diceyml.Service) bool {
	return cmp.Equal(svc1, svc2)
}

func diffServiceset(set1, set2 map[string]*diceyml.Service) (
	missingInSet1, missingInSet2, shared map[string]*diceyml.Service) {
	missingInSet1 = make(map[string]*diceyml.Service)
	missingInSet2 = make(map[string]*diceyml.Service)
	shared = make(map[string]*diceyml.Service)

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

// getDifferentServices get the 'shared' part of set1, set2 and not equal services
func getDifferentServices(set1, set2, shared map[string]*diceyml.Service) map[string]*diceyml.Service {
	r := map[string]*diceyml.Service{}
	for k := range shared {
		if !isServiceEqual(set1[k], set2[k]) {
			r[k] = set2[k]
		}
	}
	return r
}

func mergemap(major, minor map[string]*diceyml.Service) {
	for name, svc := range minor {
		if _, ok := major[name]; !ok {
			major[name] = svc
		}
	}
}

func expandGlobalEnv(envs map[string]string, dicesvc map[string]*diceyml.Service) {
	for _, svc := range dicesvc {
		if svc.Envs == nil {
			svc.Envs = make(map[string]string)
		}
		for k, v := range envs {
			if _, ok := svc.Envs[k]; !ok {
				svc.Envs[k] = v
			}
		}
	}
}
