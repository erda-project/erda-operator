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

package envs

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/erda-project/dice-operator/pkg/spec"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	EnableSpecifiedNamespace = "ENABLE_SPECIFIED_NAMESPACE"
	CustomRegCredSecret      = "CUSTOM_REGCRED_SECRET"
	ErdaCustomRegCredSecret  = "aliyun-registry"
	Pipeline                 = "pipeline"
	Orchestrator             = "orchestrator"
	OpenAPI                  = "openapi"
	ErdaServer               = "erda-server"
)

func GetAllServices(cluster *spec.DiceCluster) []diceyml.Services {
	return []diceyml.Services{
		cluster.Spec.Dice.Services,
		cluster.Spec.AddonPlatform.Services,
		cluster.Spec.Gittar.Services,
		cluster.Spec.Pandora.Services,
		cluster.Spec.DiceUI.Services,
		cluster.Spec.UC.Services,
		cluster.Spec.SpotAnalyzer.Services,
		cluster.Spec.SpotCollector.Services,
		cluster.Spec.SpotDashboard.Services,
		cluster.Spec.SpotFilebeat.Services,
		cluster.Spec.FluentBit.Services,
		cluster.Spec.SpotStatus.Services,
		cluster.Spec.SpotTelegraf.Services,
		cluster.Spec.Tmc.Services,
		cluster.Spec.Hepa.Services,
		cluster.Spec.SpotMonitor.Services,
		cluster.Spec.Fdp.Services,
		cluster.Spec.FdpUI.Services,
		cluster.Spec.MeshController.Services,
	}
}

func InjectENVs(clusterInfo map[string]string, envs map[string]map[string]string, cluster *spec.DiceCluster) {
	dependMap := genDependsOnMap(clusterInfo, cluster)

	for _, services := range GetAllServices(cluster) {
		for name, svc := range services {
			if svc == nil {
				continue
			}
			injectServiceGlobalEnv(name, svc)
			// 注入depend-envs
			injectByDependsOn(name, svc, dependMap)
			// 注入 addons-info ConfigMap
			injectEnvmap(svc, envs["addons"])
			// 注入 clusterInfo ConfigMap
			injectEnvmap(svc, envs["clusterInfo"])
			// TODO: DELETE ME
			// 注入 `name` svc 特殊化 envs
			for k, v := range envs[name] {
				svc.Envs[k] = v
			}
			// 注入update.txt-envs
			promoteUpdatetxtEnvs(svc)
		}
	}
}

func injectServiceGlobalEnv(name string, svc *diceyml.Service) {
	if svc.Envs == nil {
		svc.Envs = make(map[string]string)
	}

	if name == Pipeline || name == Orchestrator {
		// registry credential secret name
		svc.Envs[CustomRegCredSecret] = ErdaCustomRegCredSecret
	}
}

func genDependsOnMap(clusterInfo map[string]string, cluster *spec.DiceCluster) map[string]map[string]string {

	// 包含暴露公网的组件的 PUBLIC_ADD PUBLIC_URL
	dependEnvMap := GenIngServiceDepEndsOnMap(clusterInfo, cluster)

	// 生成各个组件服务的 svc address、以及 ing 地址
	for _, services := range GetAllServices(cluster) {
		UpdateDependEnvMap(dependEnvMap, services)
	}

	return dependEnvMap
}

func GenIngServiceDepEndsOnMap(clusterInfo map[string]string, cluster *spec.DiceCluster) map[string]map[string]string {
	customDomain := cluster.Spec.CustomDomain
	platformDomain := cluster.Spec.PlatformDomain
	mainPlatform := cluster.Spec.MainPlatform

	protocol := "http"
	if strutil.Contains(clusterInfo["DICE_PROTOCOL"], "https") {
		protocol = "https"
	}

	collectorPublicAddrCenter := fmt.Sprintf("collector.%s", platformDomain)
	collectorPublicURLCenter := fmt.Sprintf("%s://collector.%s", protocol, platformDomain)
	if customDomain["collector"] != "" {
		collectorPublicAddrCenter = strutil.Split(customDomain["collector"], ",", true)[0]
		collectorPublicURLCenter = fmt.Sprintf("%s://%s",
			protocol, strutil.Split(customDomain["collector"], ",", true)[0])
	}

	collectorPublicURL := map[bool]string{
		true:  collectorPublicURLCenter,
		false: mainPlatform["collector"],
	}[len(mainPlatform) == 0 || mainPlatform["collector"] == ""]

	collectorPublicAddr := map[bool]string{
		true:  collectorPublicAddrCenter,
		false: strutil.TrimPrefixes(mainPlatform["collector"], "http://", "https://"),
	}[len(mainPlatform) == 0 || mainPlatform["collector"] == ""]

	openapiPublicURLCenter := fmt.Sprintf("%s://openapi.%s", protocol, platformDomain)
	openapiPublicAddrCenter := fmt.Sprintf("openapi.%s", platformDomain)
	if customDomain["openapi"] != "" {
		openapiPublicURLCenter = fmt.Sprintf("%s://%s",
			protocol, strutil.Split(customDomain["openapi"], ",", true)[0])
		openapiPublicAddrCenter = strutil.Split(customDomain["openapi"], ",", true)[0]
	}

	openapiPublicURL := map[bool]string{
		true:  openapiPublicURLCenter,
		false: mainPlatform["openapi"],
	}[len(mainPlatform) == 0 || mainPlatform["openapi"] == ""]
	openapiPublicAddr := map[bool]string{
		true:  openapiPublicAddrCenter,
		false: strutil.TrimPrefixes(mainPlatform["openapi"], "http://", "https://"),
	}[len(mainPlatform) == 0 || mainPlatform["openapi"] == ""]

	clusterDialerPublicAddrCenter := fmt.Sprintf("cluster-dialer.%s", platformDomain)
	clusterDialerPublicURLCenter := fmt.Sprintf("%s://cluster-dialer.%s", protocol, platformDomain)
	if customDomain["cluster-dialer"] != "" {
		clusterDialerPublicAddrCenter = strutil.Split(customDomain["cluster-dialer"], ",", true)[0]
		clusterDialerPublicURLCenter = fmt.Sprintf("%s://%s",
			protocol, strutil.Split(customDomain["cluster-dialer"], ",", true)[0])
	}

	clusterDialerPublicURL := map[bool]string{
		true:  clusterDialerPublicURLCenter,
		false: mainPlatform["cluster-dialer"],
	}[len(mainPlatform) == 0 || mainPlatform["cluster-dialer"] == ""]

	clusterDialerPublicAddr := map[bool]string{
		true:  clusterDialerPublicAddrCenter,
		false: strutil.TrimPrefixes(mainPlatform["cluster-dialer"], "http://", "https://"),
	}[len(mainPlatform) == 0 || mainPlatform["cluster-dialer"] == ""]

	ucPublicAddr := fmt.Sprintf("uc.%s", platformDomain)
	if customDomain["uc"] != "" {
		ucPublicAddr = strutil.Split(customDomain["uc"], ",", true)[0]
	}
	ucPublicURL := fmt.Sprintf("%s://%s", protocol, ucPublicAddr)

	gittarPublicAddr := fmt.Sprintf("gittar.%s", platformDomain)
	if customDomain["gittar"] != "" {
		gittarPublicAddr = strutil.Split(customDomain["gittar"], ",", true)[0]
	}
	gittarPublicURL := fmt.Sprintf("%s://%s", protocol, gittarPublicAddr)

	uiPublicAddr := fmt.Sprintf("dice.%s", platformDomain)
	if customDomain["ui"] != "" {
		uiPublicAddr = strutil.Split(customDomain["ui"], ",", true)[0]
	}
	uiPublicURL := fmt.Sprintf("%s://%s", protocol, uiPublicAddr)

	soldierPublicAddr := fmt.Sprintf("soldier.%s", platformDomain)
	if customDomain["soldier"] != "" {
		soldierPublicAddr = strutil.Split(customDomain["soldier"], ",", true)[0]
	}
	soldierPublicURL := fmt.Sprintf("%s://%s", protocol, soldierPublicAddr)

	return map[string]map[string]string{
		"gittar": {
			"GITTAR_PUBLIC_ADDR": gittarPublicAddr,
			"GITTAR_PUBLIC_URL":  gittarPublicURL,
		},
		"uc": {
			"UC_PUBLIC_ADDR": ucPublicAddr,
			"UC_PUBLIC_URL":  ucPublicURL,
		},
		"collector": {
			"COLLECTOR_PUBLIC_ADDR": collectorPublicAddr,
			"COLLECTOR_PUBLIC_URL":  collectorPublicURL,
		},
		"erda-server": {
			"OPENAPI_PUBLIC_ADDR": openapiPublicAddr,
			"OPENAPI_PUBLIC_URL":  openapiPublicURL,
		},
		"cluster-dialer": {
			"CLUSTER_DIALER_PUBLIC_ADDR": clusterDialerPublicAddr,
			"CLUSTER_DIALER_PUBLIC_URL":  clusterDialerPublicURL,
		},
		"ui": {
			"UI_PUBLIC_ADDR": uiPublicAddr,
			"UI_PUBLIC_URL":  uiPublicURL,
		},
		"soldier": {
			"SOLDIER_PUBLIC_ADDR": soldierPublicAddr,
			"SOLDIER_PUBLIC_URL":  soldierPublicURL,
		},
		"netportal": {"NETPORTAL_ADDR": "netportal.default.svc.cluster.local"},
		"sonar":     {},
		"nexus":     {},
	}
}

func GenServiceAddr(svcName string, svc *diceyml.Service) map[string]string {
	addrKeyTmpl := "%s.%s.svc.cluster.local:%d"

	defaultPort := svc.Ports[0].Port
	for _, svcPort := range svc.Ports {
		if svcPort.Default {
			defaultPort = svcPort.Port
		}
	}
	var namespace = metav1.NamespaceDefault
	if os.Getenv(EnableSpecifiedNamespace) != "" {
		namespace = os.Getenv(EnableSpecifiedNamespace)
	}

	result := map[string]string{
		convertServiceAddrKey(svcName): fmt.Sprintf(addrKeyTmpl, svcName, namespace, defaultPort),
	}

	switch svcName {
	case ErdaServer:
		for _, p := range svc.Ports {
			if !p.Expose {
				continue
			}
			result[convertServiceAddrKey(OpenAPI)] = fmt.Sprintf(addrKeyTmpl, ErdaServer, namespace, p.Port)
		}
	}

	return result
}

func convertServiceAddrKey(name string) string {
	addrNameTmpl := "%s_ADDR"
	addrKey := strings.ReplaceAll(name, "-", "_")
	addrKey = strings.ToUpper(addrKey)
	return fmt.Sprintf(addrNameTmpl, addrKey)
}

func UpdateDependEnvMap(envs map[string]map[string]string, services diceyml.Services) {

	for name, svc := range services {

		// 无端口暴露（内网、公网）不需要注入
		if len(svc.Ports) < 1 {
			continue
		}

		// 暴露公网的组件
		if _, ok := envs[name]; ok {
			// addon 组件
			if len(envs[name]) == 0 || name == "netportal" {
				continue
			}

			// 非暴露公网组件
		} else {
			envs[name] = map[string]string{}
		}

		// 生成 dice 相关组件 addr
		addrs := GenServiceAddr(name, svc)
		for k, v := range addrs {
			envs[name][k] = v
		}
	}
}

func injectEnvmap(svc *diceyml.Service, addons map[string]string) {
	if svc.Envs == nil {
		svc.Envs = map[string]string{}
	}
	for k, v := range addons {
		svc.Envs[k] = v
	}
}

// promoteUpdatetxtEnvs 将 updatetxt 中的 env 的格式恢复(_key_ -> key)
// 为什么需要这样的key?
// 希望实现如下的优先级: updatetxt-envs > configmap > diceyml-origin-envs
// 如果没有限定特殊格式的 key, 则 updatetxt-envs 和
// diceyml-origin-envs 无法区分, 因为updatetxt中的env最终也仅仅是插入在diceyml中
func promoteUpdatetxtEnvs(svc *diceyml.Service) {
	for k, v := range svc.Envs {
		if strutil.HasPrefixes(k, "_") && strutil.HasSuffixes(k, "_") {
			newk := strutil.TrimSuffixes(strutil.TrimPrefixes(k, "_"), "_")
			svc.Envs[newk] = v
		}
	}
}

func injectByDependsOn(svcname string, svc *diceyml.Service,
	dependonMap map[string]map[string]string) {
	dependsOn := append(svc.DependsOn, "gittar", "uc", "collector", "erda-server", "ui", "soldier", "netportal", "cluster-dialer")
	for _, dependon := range dependsOn {
		r, ok := dependonMap[dependon]
		if !ok {
			logrus.Errorf("illegal depends_on: %v, svc: %v", dependon, svcname)
			continue
		}
		if svc.Envs == nil {
			svc.Envs = map[string]string{}
		}
		for k, v := range r {
			svc.Envs[k] = v
		}
	}
	if svc.Envs == nil {
		svc.Envs = map[string]string{}
	}
	for k, v := range dependonMap[svcname] {
		if strutil.HasSuffixes(k, "_PUBLIC_ADDR") {
			svc.Envs["SELF_PUBLIC_ADDR"] = v
		} else if strutil.HasSuffixes(k, "_ADDR") {
			svc.Envs["SELF_ADDR"] = v
		} else if strutil.HasSuffixes(k, "_PUBLIC_URL") {
			svc.Envs["SELF_PUBLIC_URL"] = v
		}
	}
	return
}

func GenDiceSvcENVs(cluster *spec.DiceCluster, addonConfigMap,
	clusterInfo map[string]string) map[string]map[string]string {
	cookieDomain := cluster.Spec.CookieDomain
	platformDomain := cluster.Spec.PlatformDomain

	addonConfigMap["REDIS_SENTINELS"] = addonConfigMap["REDIS_SENTINELS_ADDR"]

	openapiCustomDomain := cluster.Spec.CustomDomain["openapi"]
	if openapiCustomDomain != "" {
		openapiCustomDomain = strutil.Join(strutil.Map(strutil.Split(openapiCustomDomain, ",", true), func(s string) string {
			return "http://" + s + "/logincb,https://" + s + "/logincb"
		}), ",", true)
	}
	diceLoginCallback := fmt.Sprintf("http://openapi.%s/logincb,https://openapi.%s/logincb", platformDomain, platformDomain)
	if openapiCustomDomain != "" {
		diceLoginCallback = diceLoginCallback + "," + openapiCustomDomain
	}

	resultEnvs := map[string]map[string]string{
		"addons":      addonConfigMap,
		"clusterInfo": clusterInfo,
		"erda-server": {
			"COOKIE_DOMAIN":      fmt.Sprintf(".%s", cookieDomain),
			"CSRF_COOKIE_DOMAIN": fmt.Sprintf(".%s", cookieDomain),
		},
		"uc": {
			"DICE_LOGIN_CALLBACK": diceLoginCallback,
			"COOKIE_DOMAIN":       cookieDomain,
		},
	}
	return resultEnvs
}
