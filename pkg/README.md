koperator
=========

组件启动的顺序?
---------------

在 github.com/erda-project/dice-operator/pkg/cluster/launch/order.go:
bootOrder 中定义了启动顺序, 如果服务没在列表中, 则最后启动

如何 check 服务启动完成的?
--------------------------

-   github.com/erda-project/dice-operator/pkg/cluster/check/deployment.go:
    checkDeploymentAvailable
-   github.com/erda-project/dice-operator/pkg/cluster/check/daemonset.go:
    checkDaemonsetAvailable

如果 check 服务一直不过, 会怎样?
--------------------------------

check 超时 50s, 超时后, dicecluster 状态为 Pending, (kubectl get dice
中可看到), 后续组件(组件启动有顺序)均不会启动.
同时会有周期性的同步状态(spec.PeriodicSync=true), 此时如果之前 check
不过组件通过了 check, 则会继续启动后续的组件

组件中会注入哪些环境变量?
-------------------------

[./envs/envs.md](./envs/envs.md)
