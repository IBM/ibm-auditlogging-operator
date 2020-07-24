# ibm-auditlogging-operator

The ibm-auditlogging-operator contains a Fluentd solution to forward audit data that is generated by IBM Cloud Platform Common Services to a configured SIEM. The operator deploys a Fluentd daemonset containing a systemd input plugin, remote_syslog output plugin, and fluent-plugin-splunk-hec output plugin. It also deploys the Audit logging policy controller.

Important: Do not install this operator directly. Only install this operator using the IBM Common Services Operator. For more information about installing this operator and other Common Services operators, see Installer documentation. If you are using this operator as part of an IBM Cloud Pak, see the documentation for that IBM Cloud Pak to learn more about how to install and use the operator service. For more information about IBM Cloud Paks, see IBM Cloud Paks that use Common Services.

## Supported platforms

Red Hat OpenShift Container Platform 4.2 or newer installed on one of the following platforms.

- Linux x86_64
- Linux on Power (ppc64le)
- Linux on IBM Z and LinuxONE

## Operator versions

- 3.5.0
- 3.6.0
- 3.6.1
- 3.6.2
- 3.7.0

  Technology Preview - Included in version 3.6.0, 3.6.1, and 3.6.2 support for sending audit log records over HTTP.

## Prerequisites

Before you install this operator, you need to first install the operator dependencies and prerequisites:

- For the list of operator dependencies, see the IBM Knowledge Center [Common Services dependencies documentation](http://ibm.biz/cpcs_opdependencies).

- For the list of prerequisites for installing the operator, see the IBM Knowledge Center [Preparing to install services documentation](http://ibm.biz/cpcs_opinstprereq).

- ibm-auditlogging-operator must run in the `ibm-common-services` namespace

## SecurityContextConstraints Requirements

The ibm-auditlogging-operator supports running with the OpenShift Container Platform 4.3 default restricted Security Context Constraints (SCCs).

For more information about the OpenShift Container Platform Security Context Constraints, see [Managing Security Context Constraints](https://docs.openshift.com/container-platform/4.3/authentication/managing-security-context-constraints.html).

OCP 4.3 restricted SCC:

```yaml
allowHostDirVolumePlugin: false
allowHostIPC: false
allowHostNetwork: false
allowHostPID: false
allowHostPorts: false
allowPrivilegeEscalation: true
allowPrivilegedContainer: false
allowedCapabilities: null
apiVersion: security.openshift.io/v1
defaultAddCapabilities: null
fsGroup:
  type: MustRunAs
groups:
- system:authenticated
kind: SecurityContextConstraints
metadata:
  annotations:
    kubernetes.io/description: restricted denies access to all host features and requires
      pods to be run with a UID, and SELinux context that are allocated to the namespace.  This
      is the most restrictive SCC and it is used by default for authenticated users.
  creationTimestamp: "2020-03-27T15:01:00Z"
  generation: 1
  name: restricted
  resourceVersion: "6365"
  selfLink: /apis/security.openshift.io/v1/securitycontextconstraints/restricted
  uid: 6a77775c-a6d8-4341-b04c-bd826a67f67e
priority: null
readOnlyRootFilesystem: false
requiredDropCapabilities:
- KILL
- MKNOD
- SETUID
- SETGID
runAsUser:
  type: MustRunAsRange
seLinuxContext:
  type: MustRunAs
supplementalGroups:
  type: RunAsAny
users: []
volumes:
- configMap
- downwardAPI
- emptyDir
- persistentVolumeClaim
- projected
- secret
```

## Documentation

To install the operator with the IBM Common Services Operator, follow the installation and configuration instructions within the IBM Knowledge Center.

- If you are using the operator as part of an IBM Cloud Pak, see the documentation for that IBM Cloud Pak. For a list of IBM Cloud Paks, see [IBM Cloud Paks that use Common Services](http://ibm.biz/cpcs_cloudpaks).

- If you are using the operator with an IBM Containerized Software, see the IBM Cloud Platform Common Services Knowledge Center [Installer documentation](http://ibm.biz/cpcs_opinstall).

- For more information, see the [IBM Cloud Platform Common Services documentation](http://ibm.biz/cpcsdocs).

## Developer guide

As a developer, if you want to build and test this operator to try out and learn more about the operator and its capabilities, you can use the following developer guide. The guide provides commands for a quick installation and initial validation for running the operator.

    Important: The following developer guide is provided as-is and only for trial and education purposes. IBM and IBM Support does not provide any support for the usage of the operator with this developer guide. For the official supported install and usage guide for the operator, see the the IBM Knowledge Center documentation for your IBM Cloud Pak or for IBM Cloud Platform Common Services.

### Overview

- To learn more about how the ibm-auditlogging-operator was implemented, see [Operator Guidelines](https://github.com/operator-framework/getting-started#getting-started) and [Operator SDK](https://sdk.operatorframework.io/docs/)

- An operator can manage one or more controllers. The controller watches the resources for a particular CR (Custom Resource).

- All of the resources that were created by a Helm chart are now created by a controller.

- Determine how many CRDs (Custom Resource Definition) are needed. Audit logging has two CRDs:
  1. `AuditLogging`
  1. `AuditPolicy` (generated by `audit-policy-controller` repo)

## Configuration

- In order for audit logs to be forwarded to an SIEM, the `fluentd.enabled` field in the `AuditLogging` spec must be set to `true`.
- If a field in the `AuditLogging` spec is omitted, the default value will be used.
- The list of all ibm-auditlogging-operator settings can be found in the IBM Cloud Platform Common Services Knowledge Center [Configuration Documentation](http://ibm.biz/cpcs_opinstall).

### Developer Guide

- These steps are based on [Operator Framework: Getting Started](https://github.com/operator-framework/getting-started#getting-started)
  and [Creating an App Operator](https://github.com/operator-framework/operator-sdk#create-and-deploy-an-app-operator).

- Repositories: [ibm-auditlogging-operator](https://github.com/IBM/ibm-auditlogging-operator)
- Set the Go environment variables.

  `export GOPATH=/home/<username>/go`
  `export GO111MODULE=on`
  `export GOPRIVATE="github.ibm.com"`

- Create the operator skeleton.

  ```bash
  cd /home/ibmadmin/workspace/cs-operators
  operator-sdk new auditlogging-operator --repo github.com/ibm/ibm-auditlogging-operator
  ```

  1. The main program for the operator, `cmd/manager/main.go`, initializes, and runs the manager.
  1. The manager automatically registers the scheme for all custom resources defined under `pkg/apis/...` and runs all controllers under `pkg/controller/...`.
  1. The manager can restrict the namespace that all controllers watch for resources.

- Create the API definition, `kind` that is used to create the CRD.

  ```bash
  cd /home/ibmadmin/workspace/cs-operators/auditlogging-operator
  ```

  1. Create `hack/boilerplate.go.txt`.
  1. Contains copyright for generated code.

  ```bash
  operator-sdk add api --api-version=operator.ibm.com/v1alpha1 --kind=auditlogging
  ```

  1. Generates `pkg/apis/operator/v1alpha1/auditlogging_types.go`.
  1. Generates `deploy/crds/operator.ibm.com_auditloggings_crd.yaml`.
  1. Generates `deploy/crds/operator.ibm.com_v1alpha1_auditlogging_cr.yaml`.
  1. The operator can manage more than one `kind`.

- Edit `<kind>_types.go` and add the fields that will be exposed to the user. Then, regenerate the CRD.
  1. Edit `<kind>_types.go` and add fields to the `<kind>Spec` structure.

  ```bash
  operator-sdk generate k8s
  ```

  1. Updates `zz_generated.deepcopy.go`.
  1. *"Operator Framework: Getting Started" says to run `operator-sdk generate openapi`. That command is deprecated. Instead, run the nest two commands.*

  ```bash
  operator-sdk generate crds
  ```

  1. Updates `operator.ibm.com_auditloggings_crd.yaml`.
  1. `openapi-gen --logtostderr=true -o "" -i ./pkg/apis/operator/v1alpha1 -O zz_generated.openapi -p ./pkg/apis/operator/v1alpha1 -h hack/boilerplate.go.txt -r "-"`
  1. Creates `zz_generated.openapi.go`.
  1. If you need to build `openapi-gen`, follow these steps. The binary is built in `$GOPATH/bin`.

  ```bash
  git clone https://github.com/kubernetes/kube-openapi.git
  cd kube-openapi
  go mod tidy
  go build -o ./bin/openapi-gen k8s.io/kube-openapi/cmd/openapi-gen
  ```

**IMPORTANT**: Anytime you modify `<kind>_types.go`, you must run `generate k8s`, `generate crds`, and `openapi-gen` again to update the CRD and the generated code.

- Create the controller. It creates resources like Deployments, and DaemonSets.

  ```bash
  operator-sdk add gicontroller --api-version=operator.ibm.com/v1alpha1 --kind=auditlogging
  ```

  1. There is one controller for each `kind` or CRD. The controller watches and reconciles the resources that are owned by the CR.
  1. For information about the Go types that implement Deployments, DaemonSets, and others, go to <https://godoc.org/k8s.io/api/apps/v1>.
  1. For information about the Go types that implement Pods, VolumeMounts, and others, go to <https://godoc.org/k8s.io/api/core/v1>.
  1. For information about the Go types that implement Ingress, go to <https://godoc.org/k8s.io/api/networking/v1beta1>.

### Testing

#### Installing by using the OCP Console

1. Create the `ibm-common-services` namespace.
1. Create an [OperatorSource](https://github.com/IBM/operand-deployment-lifecycle-manager/blob/master/docs/install/common-service-integration.md#1-create-an-operatorsource-in-the-openshift-cluster) in your cluster.
1. Select the `Operators` tab and in the drop-down select `OperatorHub`.
1. Search for the `ibm-auditlogging-operator`.
1. Install the operator in the `ibm-common-services` namespace.
1. Create an `AuditLogging` instance.

#### Prerequisites for building the operator locally

- [Install linters](https://github.com/IBM/go-repo-template/blob/master/docs/development.md)

#### Run the operator on a cluster

- `make install`
- Run tests on the cluster.
- `make uninstall`

#### Run the operator locally

- `make install-local`
- Run tests on the cluster.
- `make uninstall-local`

#### Operator SDK's Test Framework

- [Unit Testing](https://sdk.operatorframework.io/docs/golang/legacy/unit-testing)
- [E2E Testing](https://sdk.operatorframework.io/docs/golang/legacy/e2e-tests/)
- To run unit tests use, `make test`
- To run e2e tests use, `make test-e2e`

#### Debugging the Operator

Run these commands to collect information about the audit logging deployment.

1. `kubectl get pods -n ibm-common-services | grep audit`
1. `kubectl get serviceaccount -n ibm-common-services | grep audit`
1. `kubectl get secrets -n ibm-common-servces | grep audit`
1. `kubectl get services -n ibm-common-services | grep common-audit`

These steps verify:

- The ibm-auditlogging-operator is running.
- The audit logging operands (audit-policy-controller and audit-logging daemonset) are running.
- The secrets (audit-server-certs, audit-certs) are created.
- The ibm-auditlogging-operator and ibm-auditlogging-operand service accounts are created.
- The common-audit-logging service is created.

Run these commands to collect logs:

1. `kubectl logs -n ibm-common-services  ibm-auditlogging-operator pod`
1. `kubectl logs -n ibm-common-services  audit-policy-controller pod`
1. `kubectl logs -n ibm-common-services  audit-logging-fluentd daemonset pods`

#### End-to-End testing

For more instructions on how to run end-to-end testing with the Operand Deployment Lifecycle Manager, see ODLM guide.
