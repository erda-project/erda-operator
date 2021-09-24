#  Erda Operator Structure Introduction

## CRD Introduction

Genertae the CRD by [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder)

- Name: erdas.terminus.io
- Group: terminus.io
- Version: v1beta1

```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.4.1
  creationTimestamp: null
  name: erdas.erda.terminus.io
spec:
  group: erda.terminus.io
  names:
    kind: Erda
    listKind: ErdaList
    plural: erdas
    singular: erda
  scope: Namespaced
  versions:
  - additionalPrinterColumns:
    - description: the erda status phase
      jsonPath: .status.phase
      name: Phase
      type: string
    - jsonPath: .metadata.creationTimestamp
      name: Age
      type: date
    name: v1beta1
```



## CR Structure

### Erda

```go
// Erda is the Schema for the erdas API
type Erda struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              *ErdaSpec   `json:"spec,omitempty"`
	Status            *ErdaStatus `json:"status,omitempty"`
}
```

### Erda List

```go
// ErdaList contains a list of Erda
type ErdaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Erda `json:"items"`
}
```

### Erda Spec

```go
// ErdaSpec defines the desired state of Erda
type ErdaSpec struct {
 	// Applications indicate applications can be deployed on the current namespace, 
  // every application contains a group of services
	Applications []Application `yaml:"applications" json:"applications"`
}
```

#### Application

```go
// Metdata indicate the meta data which use in application and component
// Name indicate the object‘s name
// Namespace value is same as erda namespace‘s value
type Metadata struct {
	Name        string            `yaml:"name" json:"name"`
	Namespace   string            `yaml:"-" json:"-"`
	Annotations map[string]string `yaml:"annotations,omitempty" json:"annotations,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
}

type Application struct {
	Metadata   `yaml:",inline" json:",inline"`
	// Envs indicate the application environment variable, 
	// which will append to all components environment variable 
  // if the key does not exist. if the environment variable 
  // key is the same as the key of the component, the value will use the 
  // component's environment variable value first
	Envs       []corev1.EnvVar        `yaml:"envs,omitempty" json:"envs,omitempty"`
  // EnvFrom indicate the configmap resource will be set to envs
	EnvFrom    []corev1.EnvFromSource `yaml:"envFrom,omitempty" json:"envFrom,omitempty"`
  // Components indicate the services which want to be deployed on the current application
	Components []Component            `yaml:"components,omitempty" json:"components,omitempty"`
}
```

#### Component

```go
// Component indicate the component service config which deploy on Kubernetes cluster
type Component struct {
	Metadata      `yaml:",inline" json:",inline"`
	ComponentSpec `yaml:",inline" json:",inline"`
}

// ComponentSpec indicate the description of config details
type ComponentSpec struct {
  // workload indicate the deploy type of your application
  // support Stateful,Stateless and PerNode, deafult is Stateless
  // Stateful indicate the stateful service
  // Stateless indicate the stateless service
  // PerNode indicate the daemon service
	WorkLoad       WorkLoadType                `yaml:"workload" json:"workload"`
  // ImageInfo indicate the component image info, 
  // it contains image repo URL and tag if the repo is private, 
  // the username and password or the Secret need to be filled
	ImageInfo      ImageInfo                   `yaml:"imageInfo" json:"imageInfo"`
  // Replicas indicate the count of the component you want to deploy
	Replicas       *int32                      `yaml:"replicas" json:"replicas"`
  // Resources indicate the component service usage of the cpu and memory
  // The Request means the minimal resource required when the service is starts
  // The Limit means the maximum resource that can be used when the service is running
	Resources      corev1.ResourceRequirements `yaml:"resources" json:"resources"`
  // Affinity indicates the component service will be deployed on some constraint condition
	Affinity       []Affinity                  `yaml:"affinity,omitempty" json:"affinity,omitempty"`
  // Envs indicates the component service the environment variable
	Envs           []corev1.EnvVar             `yaml:"envs,omitempty" json:"envs,omitempty"`
  // EnvFrom indicates that config map resource will be set to the environment variable
	EnvFrom        []corev1.EnvFromSource      `yaml:"envFrom,omitempty" json:"envFrom,omitempty"`
  // Command indicates the command used when the component service is started
	Command        []string                    `yaml:"command,omitempty" json:"command,omitempty"`
  // Storage indicates the storage config of the component used
  // only support volume currently
	Storage        Storage                     `yaml:"storage,omitempty" json:"storage,omitempty"`
  // Hosts indicates component alias name
	Hosts          []string                    `yaml:"hosts,omitempty" json:"hosts,omitempty"`
  // DependsOn indicates if the component depend some component,
  // the public and private address environment variable of the 
  // dependon service will be injected in component
  // it does not implement currently
	DependsOn      []string                    `yaml:"dependsOn,omitempty" json:"dependsOn,omitempty"`
  // Network indicates the config of component domain and address
	Network        *Network                    `yaml:"network,omitempty" json:"network,omitempty"`
  // HealthCheck indicates check component’s status when it starts
	HealthCheck    *HealthCheck                `yaml:"healthCheck,omitempty" json:"healthCheck,omitempty"`
  // Configurations indicate does the component need 
  // secret and config map if the data or string data 
  // isn't empty, operator will create/update configurations
  // by given type. if only fill the name, the operator will
  // to check the configuration does it exist on Kubernetes cluster
	Configurations []Configuration             `yaml:"configurations,omitempty" json:"configurations,omitempty"`
}

type HealthCheck struct {
  // Duration means the time duration when check the component whether ready
	Duration  int32      `yaml:"duration,omitempty" json:"duration,omitempty"`
  // HTTPCheck means that the component will check whether
  // the component is ready through the path and port of http
	HTTPCheck *HTTPCheck `yaml:"httpCheck,omitempty" json:"httpCheck,omitempty"`
  // ExecCheck means that the component will check whether
  // the component is ready through specified command lines
	ExecCheck *ExecCheck `yaml:"execCheck,omitempty" json:"execCheck,omitempty"`
}

// HTTPCheck means that the component will check whether
// the component is ready through the path and port of http
type HTTPCheck struct {
	Port int    `yaml:"port,omitempty" json:"port,omitempty"`
	Path string `yaml:"path,omitempty" json:"path,omitempty"`
}

// ExecCheck means that the component will check whether
// the component is ready through specified command lines
type ExecCheck struct {
	Command []string `yaml:"command,omitempty" json:"command,omitempty"`
}

// Configurations indicate does the component need 
// secret and config map if the data or string data 
// isn't empty, operator will create/update configurations
// by given type. if only fill the name, the operator will
// to check the configuration does it exist on Kubernetes cluster
type Configuration struct {
  // Name is the Configuration Name, if the Configuration not exist, 
  // it will be created use the Name
	Name       string            `yaml:"name" json:"name"`
  // Type indicates the Configuration use which type config of the Kuberneter
  // support ConfigMap and Secret
	Type       ConfigurationType `yaml:"type" json:"type"`
  // TargetPath is set, the Configuration will be mount 
  // in the component to specified target path
	TargetPath string            `yaml:"targetPath" json:"targetPath"`
  // Data means a map that needs string as key and []byte as value.
  // it is used for secret type at mostly condition and 
  // the []byte will be encoded as base64 if use secret type
	Data       map[string][]byte `yaml:"data,omitempty" json:"data,omitempty"`
  // StringData means a map that needs string as key and string as value
  // it is used for configmap type at mostly condition
	StringData map[string]string `yaml:"stringData,omitempty" json:"stringData,omitempty"`
}

// Affinity needs to be perfected
type Affinity struct {
	Key   string       `yaml:"key" json:"key"`
	Value string       `yaml:"value" json:"value"`
	Exist bool         `yaml:"exist" json:"exist"`
	Type  AffinityType `yaml:"type" json:"type"`
}

// ImageInfo indicate the component image info
type ImageInfo struct {
  // Image indicates the image is used when the component start
	Image      string `yaml:"image" json:"image"`
  // UserName and Password means the user can fill those two 
  // params to log into repo when pulling docker image from the private repo
  // It is currently not implemented
  // Mutually exclusive with pullSecret
	UserName   string `yaml:"userName,omitempty" json:"userName,omitempty"`
	Password   string `yaml:"password,omitempty" json:"password,omitempty"`
  // PullPolicy means image pull policy, support Always and IfNotPresent
  // If the Policy be set as Always, the image will be pulling everytime when the component starts
  // If the Policy be set as IfNotPresent, the image will be pulling when the image not exist
	PullPolicy string `yaml:"pullPolicy,omitempty" json:"pullPolicy,omitempty"`
  // PullSecret means use secret to pull image when the image repo is private
  // Mutually exclusive with UserName and Password
	PullSecret string `yaml:"pullSecret,omitempty" json:"pullSecret,omitempty"`
}

// Storage indicates the storage config of the component used
type Storage struct {
	Volumes []Volume `yaml:"volumes,omitempty" json:"volumes,omitempty"`
}

// needs to be perfected
// Volume means component used volume
type Volume struct {
  // Size means the capacity of the volume usage
	Size         *resource.Quantity `yaml:"size,omitempty" json:"size,omitempty"`
  // StorageClass means the storage use which prepraed storageclass
	StorageClass string             `yaml:"storageClass,omitempty" json:"storageClass,omitempty"`
  // SourcePath means the volume from path
  // if it be set, volume will be mounted as hostpath file
	SourcePath   string             `yaml:"sourcePath,omitempty" json:"sourcePath,omitempty"`
  // Targetpath means the volume will be mount to target path in component
	TargetPath   string             `yaml:"targetPath,omitempty" json:"targetPath,omitempty"`
  // ReadOnly if is true, the data on volume can be read only
	ReadOnly     bool               `yaml:"readOnly,omitempty" json:"readOnly,omitempty"`
  // Snapshot means volume will create snapshot
	Snapshot     *VolumeSnapshot    `yaml:"snapshot,omitempty" json:"snapshot,omitempty"`
}

// VolumeSnapshot indicates 
// It is currently not implemented
type VolumeSnapshot struct {
  // SnapShotClass means the snapshot use which prepraed snapshotclass
	SnapShotClass string `yaml:"snapshotClass,omitempty" json:"snapshotClass,omitempty"`
  // MaxHistory means the max count of the snapshot could be created
	MaxHistory    int32  `yaml:"maxHistory,omitempty" json:"maxHistory,omitempty"`
}

type Network struct {
  // Type means the component network type
  // if type equal host, the network of component
  // is same as network of host
	Type NetworkType `yaml:"type,omitempty" json:"type,omitempty"`
	// ServiceDiscovery means component the address when other components access
  // if the domain be filled, it will created ingrees for public
  // the first ServiceDiscovery is deafult, the port will be set at Kubernetes service
	ServiceDiscovery []ServiceDiscovery `yaml:"serviceDiscovery,omitempty" json:"serviceDiscovery,omitempty"`
	Microcomponents    *Microcomponents     `yaml:"microcomponent,omitempty" json:"microcomponents,omitempty"`
}

type ServiceDiscovery struct {
  // Port is the serivce access port
	Port     int32  `yaml:"port" json:"port"`
  // Protocol is the serivce access protocol
  // support TCP UDP and SCTP
	Protocol string `yaml:"protocol" json:"protocol"`
  // Domain means the component public address for accessing
	Domain   string `yaml:"domain,omitempty" json:"domain,omitempty"`
  // Path means the public address for accessing with specified path
	Path     string `yaml:"path,omitempty" json:"path,omitempty"`
}

// needs to be perfected
// it does not implement currently
type TrafficSecurity struct {
	Mode string `yaml:"mode,omitempty" json:"mode,omitempty"`
}
// needs to be perfected
// it does not implement currently
type Microcomponents struct {
	MeshEnable      *bool           `yaml:"meshEnable,omitempty" json:"meshEnable,omitempty"`
	TrafficSecurity TrafficSecurity `yaml:"trafficSecurity,omitempty" json:"trafficSecurity,omitempty"`
	Endpoints       []Endpoint      `yaml:"endpoints,omitempty" json:"endpoints,omitempty"`
}
// needs to be perfected
// it does not implement currently
type Endpoint struct {
	Domain      string           `yaml:"domain,omitempty" json:"domain,omitempty"`
	Path        string           `yaml:"path,omitempty" json:"path,omitempty"`
	BackendPath string           `yaml:"backend_path,omitempty" json:"backend_path,omitempty"`
	Policies    EndpointPolicies `yaml:"policies,omitempty" json:"policies,omitempty"`
}
// needs to be perfected
// it does not implement currently
type EndpointPolicies struct {
	Cors      *map[string]apiextensionsv1.JSON `yaml:"cors,omitempty" json:"cors,omitempty"`
	RateLimit *map[string]apiextensionsv1.JSON `yaml:"rateLimit,omitempty" json:"rateLimit,omitempty"`
}
```



### Erda Status

```go
// ErdaStatus defines the observed state of Erda
type ErdaStatus struct {
	Phase        PhaseType           `yaml:"phase,omitempty" json:"phase,omitempty"`
	Applications []ApplicationStatus `yaml:"applications,omitempty"json:"applications,omitempty"`
}

type ApplicationStatus struct {
	Name       string            `json:"name"`
	Status     StatusType        `json:"status"`
	Components []ComponentStatus `json:"components"`
}


type ComponentStatus struct {
	Name   string     `json:"name"`
	Status StatusType `json:"status"`
}
```



```go
type StatusType string

const (
	StatusReady     StatusType = "Ready"
	StatusDeploying StatusType = "Deploying"
	StatusUnReady   StatusType = "Unready"
	StatusRunning   StatusType = "Running"
	StatusFailed    StatusType = "Failed"
	StatusCompleted StatusType = "Completed"
	StatusUnKnown   StatusType = "UnKnown"
)

type PhaseType string

const (
	PhaseReady     PhaseType = "Ready"
	PhaseDeploying PhaseType = "Deploying"
)

const (
	ErdaPrefix  = "erda"
	PreJobType  = "PreJob"
	PostJobType = "PostJob"
)

const (
	ErdaJobTypeLabel  = "erda.io/job-type"
	ErdaOperatorLabel = "erda.io/erda-operator"
	ErdaOperatorApp   = "erda.io/erda-operator-app"
)

type NetworkType string

const (
	NetworkKindHost NetworkType = "host"
)

type WorkLoadType string

const (
	Stateful  WorkLoadType = "Stateful"
	Stateless WorkLoadType = "Stateless"
	PerNode   WorkLoadType = "PerNode"
)

type AffinityType string

const (
	NodePreferredAffinityType AffinityType = "NodePreferred"
	NodeRequestedAffinityType AffinityType = "NodeRequested"
)

const (
	CPUBound string = "cpu_bound"
	IOBound  string = "io_bound"
)

type ConfigurationType string

const (
	ConfigurationSecret    = "Secret"
	ConfigurationConfigMap = "ConfigMap"
)
```


