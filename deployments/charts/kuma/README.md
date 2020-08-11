[![][kuma-logo]][kuma-url]

# kuma

The kuma chart supports all the features and options provided by `kumactl install control-plane`.
The chart supports Helm v3+.

## Values

| Parameter                                   | Description                                                                       | Default                              |
|---------------------------------------------|-----------------------------------------------------------------------------------|--------------------------------------|
| `global.image.registry`                     | Default registry for all Kuma images                                              | `kong-docker-kuma-docker.bintray.io` |
| `global.image.tag`                          | Default tag for all Kuma images                                                   | nil, defaults to Chart.AppVersion    |
| `patchSystemNamespace`                      | Patch the release namespace with the Kuma system label                            | `true`                               |
| `controlPlane.logLevel`                     | Kuma CP log level: one of off\|info\|debug                                        | `info`                               |
| `controlPlane.mode`                         | Kuma CP modes: one of standalone\|remote\|global                                  | `standalone`                         |
| `controlPlane.zone`                         | Kuma zone name                                                                    | nil                                  |
| `controlPlane.kdsGlobalAddress`             | URL of Global Kuma CP                                                             |                                      |
| `controlPlane.useNodePort`                  | Use NodePort instead of LoadBalancer                                              | `false`                              |
| `controlPlane.injectorFailurePolicy`        | Failure policy of the mutating webhook implemented by the Kuma Injector component | `Ignore`                             |
| `controlPlane.service.name`                 | Service name of the Kuma Control Plane                                            | nil                                  |
| `controlPlane.service.type`                 | Service type of the Kuma Control Plane                                            | ClusterIP                            |
| `controlPlane.defaults.skipMeshCreation`    | Whether or not to skip creating the default Mesh                                  | `true`                               |
| `controlPlane.resources`                    | The K8s resources spec for Kuma CP                                                | nil, differs based on mode           |
| `controlPlane.tls.{admission,sds,kds}.cert` | TLS certificate for the Admission, SDS, and KDS servers, respectively             | nil, generated and self-signed       |
| `controlPlane.tls.{admission,sds,kds}.key`  | TLS key for the Admission, SDS, and KDS servers, respectively                     | nil, generated and self-signed       |
| `controlPlane.imagePullPolicy`              | Kuma CP ImagePullPolicy                                                           | `IfNotPresent`                       |
| `controlPlane.image.registry`               | Kuma CP image registry                                                            | nil, uses global                     |
| `controlPlane.image.repository`             | Kuma CP image repository                                                          | `kuma-cp`                            |
| `controlPlane.image.tag`                    | Kuma CP image tag                                                                 | nil, uses global                     |
| `cni.enabled`                               | Install Kuma with CNI instead of proxy init container                             | `false`                              |
| `cni.logLevel`                              | CNI log level: one of off\|info\|debug                                            | `info`                               |
| `cni.image.registry`                        | CNI image registry                                                                | `docker.io`                          |
| `cni.image.repository`                      | CNI image repository                                                              | `lobkovilya/install-cni`             |
| `cni.image.tag`                             | The CNI image tag                                                                 | `0.0.1`                              |
| `dataPlane.image.registry`                  | The Kuma DP image registry                                                        | nil, uses global                     |
| `dataPlane.image.repository`                | The Kuma DP image repository                                                      | `kuma-cp`                            |
| `dataPlane.image.tag`                       | The Kuma DP image tag                                                             | nil, uses global                     |
| `dataPlane.initImage.registry`              | The Kuma DP init image registry                                                   | nil, uses global                     |
| `dataPlane.initImage.repository`            | The Kuma DP init image repository                                                 | `kuma-init`                          |
| `dataPlane.initImage.tag`                   | The Kuma DP init image tag                                                        | nil, uses global                     |

## Custom Resource Definitions

All Kuma CRDs are loaded via the [`crds`](crds) directory. For more detailed information on CRDs and Helm,
please refer to [the Helm documentation][helm-crd].


[kuma-url]: https://kuma.io/
[kuma-logo]: https://kuma-public-assets.s3.amazonaws.com/kuma-logo-v2.png
[helm-crd]: https://helm.sh/docs/chart_best_practices/custom_resource_definitions/