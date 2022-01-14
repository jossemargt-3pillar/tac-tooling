# RKE2 Chart updater - rup

Just like `rupee` but simpler.

```shell
$ go build -o rup .
[... build logs ...]
$ rup --help
```

```shell
$ RKE2_REPO=$(mktemp -d '/tmp/XXXXXX')
$ git clone -o upstream --depth 1 https://github.com/rancher/rke2-charts.git $RKE2_REPO
$ ./rup --charts $RKE2_REPO rke2-multus rancher/hardened-cni-plugins=0.0.7 rancher/hardened-multus-cni=0.1.7
cniplugins:
  image:
    repository: rancher/hardened-cni-plugins
    tag: 0.0.7
  skipcnis: flannel
global:
  systemDefaultRegistry: ""
multus:
  image:
    repository: rancher/hardened-multus-cni
    tag: 0.1.7
```
