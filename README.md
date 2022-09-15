# erda-operator
Kubernetes operator for Erda service components

## Enable/Disable Erda service components pod autoscaling

### Enable/Disable autoscaling for all Erda service components

In this case, only need set spec.enableAutoScale in CRD object dice or erda。
* Enable:
  * set spec.enableAutoScale as true 
* Disable:
  * set spec.enableAutoScale as false


### Enable/Disable autoscaling for selected Erda service components

In this case, need set: 

* step 1: set env in dice-operator （for example， enable/disable service telegraf and telegraf-app pod autoscaling）
  ```yaml
  - name: ERDA_PA_COMPONENT_LIST
    value: telegraf,telegraf-app
  ```

* step 2: spec.enableAutoScale in CRD object dice or erda
  * Enable:
      * set spec.enableAutoScale as true
  * Disable:
      * set spec.enableAutoScale as false
 
If you have enabled autoscaling for all Erda service components before enable/disable autoscaling for selected Erda service components, you'd better set spec.enableAutoScale in CRD object dice or erda as `false` before set env in dice-operator, otherwise some enabled services' hpa/vpa objects will not be controlled by dice-operator. 
