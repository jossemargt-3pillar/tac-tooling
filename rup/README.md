# RKE2 Chart updater - rup

Just like `rupee` but simpler.

## Build

```sh
$ go build -o rup .
[... build logs ...]
$ rup --help
```

## Usage

```sh
./rup --help
NAME:
   rup - rke2 charts updater

USAGE:
   rup [options] [<version field>=<version value>]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --charts value, -c value  rke2 charts directory (default: "rke2-charts") [$CHARTS]
   --in-place, -i            write changes into their respective files (default: false)
   --print, -p               print resulting yaml file on STDOUT (default: false)
   --help, -h                show help (default: false)
```

### Showcase

```sh
# rup --chart flag is not needed when CHART env. var is present
$ export CHARTS=$(mktemp -d '/tmp/XXXXXX')
$ git clone -o upstream --depth 1 https://github.com/rancher/rke2-charts.git $CHARTS

## print versions
# rup <rke2 package name> [no args]
$ ./rup rke2-canal-1.19-1.20
appVersion: v3.13.3
version: v3.13.3-build20211022
rancher/hardened-calico: v3.13.3-build20210223
rancher/hardened-flannel: v0.14.1-build20211022
packageVersion: 5

## print updated version, no file change
# rup <rke2 package name> [field=version]
$ ./rup rke2-canal-1.19-1.20 appVersion=v3.13.4
appVersion: v3.13.4
version: v3.13.3-build20211022
rancher/hardened-calico: v3.13.3-build20210223
rancher/hardened-flannel: v0.14.1-build20211022
packageVersion: 5

## print resulting yaml file into STDOUT
# rup -p|--print <rke2 package name> [field=version]
$ ./rup -p rke2-canal-1.19-1.20 appVersion=v3.13.4
apiVersion: v1
appVersion: v3.13.4
[... skipped for brevity ...]

## Write changes into their respective files
# rup -i|--in-place <rke2 package name> [field=version]
$ ./rup -i rke2-canal-1.19-1.20 appVersion=v3.13.4
```
