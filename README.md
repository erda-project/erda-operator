# Erda-Operator

## Introduction

Erda-Operator is a Kubernetes operator developed by [Erda Team](https://github.com/erda-project/erda). It helps users who know the Kubernetes less easily to deploy an application and service on the Kubernetes cluster with Erda YAML.



## Quickly Started

- Install the erda-operator on your Kubernetes cluster, please refer to the [Installation](#installation)

- Write the following sample to the `sample-test.yaml` file on your Kubernetes cluster

  ```yaml
  # an nginx sample
  apiVersion: erda.terminus.io/v1beta1
  kind: Erda
  metadata:
    name: sample-test
    namespace: erda-system
  spec:
    applications:
    - name: erda
      components:
      - name: erda-nginx-sample
        labels:
          app: "nginx"
        annotations:
          nginx-app: "this is an nginx app"
        envs:
        # if you have the env value which starts and ends with the '_' symbol,
        # it will be rewritten the key's value between the '_', 
        # e.g.
        # the test's value '1234' will be rewritten to '5678'
        - name: _test_
          value: "5678"
        - name: test
          value: "1234"
        # workload indicated the deploy type of your application
        # support Stateful,Stateless and PerNode, deafult is Stateless
        # Stateful indicated the stateful service
        # Stateless indicated the stateless service
        # PerNode indicated the daemon service
        workload: Stateless
        healthCheck:
          duration: 100
          execCheck:
            command:
              - "ls"
        replicas: 1
        imageInfo:
          image: nginx:1.14.2
          pullPolicy: IfNotPresent
        resources:
          limits:
            cpu: 200m
            memory: 256Mi
          requests:
            cpu: 100m
            memory: 128Mi
        # if network not null, the Serivce of Kubertnetes will be created
        # the first serviceDiscovery will be set the service default port and it will be written in env
        # e.g.
        # SLEF_ADDR: nginx-sample.default.svc.cluster.local:80
        # if the domain is set, the ingress will be created,
        # and the ingress will be written in environment variable
        # e.g.
        # NGINX_SAMPLE_PUBLIC_ADDR: nginx-sample.erda.cloud
        network:
          serviceDiscovery:
            - port: 80
              protocol: TCP
              domain: erda-nginx-sample.erda.cloud
  ```

- Create the erda application on Kubernetes with command line:

  > kubectl apply -f ./sample-test.yaml

- Watch the erda status with command line:

  > kubectl get erda -n erda-system

- Watch the pod status with command line:

  > kubectl get pods -n Erda-system



## Installation

**TBD**



# Documentation

- Introduce [structure of the erda](./docs/structure_v1beta1.md)
- Introduce the feature of erda-operator
