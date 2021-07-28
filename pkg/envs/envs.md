depends\_on
-----------

| depends\_on    | 注入的环境变量                                                               |
|----------------|------------------------------------------------------------------------------|
| cmdb           | CMDB\_ADDR                                                                   |
| status         | STATUS\_ADDR                                                                 |
| headless       | HEADLESS\_ADDR                                                               |
| addon-platform | ADDON\_PLATFORM\_ADDR                                                        |
| dicehub        | DICEHUB\_ADDR                                                                |
| eventbox       | EVENTBOX\_ADDR                                                               |
| gittar         | GITTAR\_ADDR, GITTAR\_PUBLIC\_ADDR, GITTAR\_PUBLIC\_URL                      |
| dashboard      | DASHBOARD\_ADDR                                                              |
| uc             | UC\_ADDR, UC\_PUBLIC\_ADDR, UC\_PUBLIC\_URL                                  |
| officer        | OFFICER\_ADDR                                                                |
| pipeline       | PIPELINE\_ADDR                                                               |
| collector      | COLLECTOR\_ADDR, COLLECTOR\_PUBLIC\_ADDR( \* ), COLLECTOR\_PUBLIC\_URL( \* ) |
| orchestrator   | ORCHESTRATOR\_ADDR                                                           |
| openapi        | OPENAPI\_ADDR, OPENAPI\_PUBLIC\_ADDR( \* ), OPENAPI\_PUBLIC\_URL( \* )       |
| scheduler      | SCHEDULER\_ADDR                                                              |
| sonar          | SONAR\_ADDR, SONAR\_PUBLIC\_ADDR, SONAR\_PUBLIC\_URL                         |
| ui             | UI\_PUBLIC\_ADDR, UI\_PUBLIC\_URL                                            |
| nexus          | NEXUS\_ADDR                                                                  |
| hepa           | HEPA\_ADDR                                                                   |
| netportal      | NETPORTAL\_ADDR                                                              |
| gittar-adaptor | GITTAR\_ADAPTOR\_ADDR                                                        |
| qa             | QA\_ADDR                                                                     |
| tmc            | TMC\_ADDR                                                                    |
| pandora        | PANDORA\_ADDR                                                                |

带(\*)表示当处于saas集群中时, 注入的是中心集群的地址

addons
------

所有组件注入了以下addon相关的环境变量, 具体内容请看 dice-platform-addon
configmap

-   ES\_URL
-   ES\_SECURITY\_ENABLE
-   ES\_SECURITY\_USERNAME
-   ES\_SECURITY\_PASSWORD
-   ETCD
-   MYSQL\_DATABASE
-   MYSQL\_HOST
-   MYSQL\_PASSWORD
-   MYSQL\_PORT
-   MYSQL\_USERNAME
-   REDIS\_MASTER\_NAME
-   REDIS\_PASSWORD
-   REDIS\_SENTINELS
-   KAFKA
-   CASSANDRA

clusterinfo
-----------

所有组件注入了以下clusterinfo, 具体内容请看 dice-cluster-info configmap

-   DICE\_CLUSTER\_NAME
-   DICE\_CLUSTER\_TYPE
-   DICE\_HTTP\_PORT
-   DICE\_INSIDE
-   DICE\_IS\_EDGE
-   DICE\_PROTOCOL
-   DICE\_ROOT\_DOMAIN
-   DICE\_SSH\_PASSWORD
-   DICE\_SSH\_USER
-   DICE\_STORAGE\_MOUNTPOINT
-   DICE\_VERSION
-   ETCD\_ADDR
-   ETCD\_ENDPOINTS
-   LB\_ADDR
-   LB\_URL
-   LB\_MONITOR\_ADDR
-   LB\_MONITOR\_URL
-   LB\_VIP\_ADDR
-   LB\_VIP\_URL
-   MASTER\_ADDR
-   MASTER\_URL
-   MASTER\_MONITOR\_ADDR
-   MASTER\_MONITOR\_URL
-   MASTER\_VIP\_ADDR
-   MASTER\_VIP\_URL
