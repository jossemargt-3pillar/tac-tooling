# Rupee - rke2 upstream packaging references updater

```shell
$ go build -o rupee .
[... build logs ...]
$ rupee --help
```

```shell
$ rupee listCharts
[
  "rke2-canal-1.19-1.20",
  "rke2-kube-proxy-1.20",
  "rke2-kube-proxy-1.18",
  "rke2-kube-proxy-1.19",
  "rke2-kube-proxy-1.21",
  "harvester-csi-driver",
  "rke2-coredns",
  "rke2-metrics-server",
  "rke2-cilium",
  "rke2-calico",
  "rke2-ingress-nginx",
  "cilium",
  "rke2-canal",
  "rke2-multus",
  "harvester-cloud-provider"
]
```

```shell
$ rupee getVersions  rke2-canal-1.19-1.20
{
  "appVersion": "v3.13.3",
  "packageVersion": "05",
  "rancher/hardened-calico": "v3.13.3-build20210223",
  "rancher/hardened-flannel": "v0.14.1-build20211022",
  "version": "v3.13.3-build20211022"
}
```

```shell
# TODO: Add -i to add change in-place just like sed
$ rupee bump rke2-canal-1.19-1.20 appVersion=0.0.7
apiVersion: v1
name: rke2-canal
description: Install Canal Network Plugin.
version: v3.13.3-build20211022
appVersion: 0.0.7
home: https://www.projectcalico.org/
keywords:
  - canal
sources:
  - https://github.com/rancher/rke2-charts
maintainers:
  - name: Rancher Labs
    email: charts@rancher.com
```