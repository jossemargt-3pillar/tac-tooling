# Rupee - rke2 upstream packaging references updater

```shell
$ go build -o rupee .
[... build logs ...]
$ rupee --help
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
