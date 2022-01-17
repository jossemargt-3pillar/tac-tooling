package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
	"k8s.io/helm/pkg/chartutil"
)

var (
	ErrNoChart = errors.New("no chart provided")
)

const (
	FilePackage = "package.yaml"
	FileChart   = "Chart.yaml"
	FileValues  = "values.yaml"
)

func main() {
	app := &cli.App{
		Usage:     "rke2 charts updater",
		UsageText: "rup [options] [<version field>=<version value>]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "charts",
				Value:   "rke2-charts",
				Usage:   "rke2 charts directory",
				Aliases: []string{"c"},
				EnvVars: []string{"CHARTS"},
			},
			&cli.BoolFlag{
				Name:    "in-place",
				Usage:   "write changes into their respective files",
				Aliases: []string{"i"},
			},
			&cli.BoolFlag{
				Name:    "print",
				Usage:   "print resulting yaml file on STDOUT",
				Aliases: []string{"p"},
			},
		},
		Action: func(c *cli.Context) error {

			if c.NArg() == 0 {
				return ErrNoChart
			}

			chartName := c.Args().First()
			chartBasePath := c.String("charts")
			chartPath := filepath.Join(chartBasePath, "packages", chartName)

			overrides := make(map[string]string, c.NArg()-1)
			for _, v := range c.Args().Tail() {
				a := strings.Split(v, "=")

				if len(a) < 2 {
					continue
				}

				overrides[a[0]] = a[1]
			}

			matchers := make(map[string]YAMLUpdater)

			matchers[FileChart] = NewUpdater([]string{"appVersion", "version"})
			matchers[FilePackage] = NewUpdater([]string{"packageVersion"})
			matchers[FileValues] = NewImageUpdater("tag")

			err := filepath.Walk(chartPath, func(path string, info fs.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if info.IsDir() {
					if strings.Contains(path, "generated") {
						return filepath.SkipDir
					}
					return nil
				}

				if u, ok := matchers[filepath.Base(path)]; ok {
					f, err := os.Open(path)
					if err != nil {
						return err
					}
					defer f.Close()

					var b bytes.Buffer
					if _, err := io.Copy(&b, f); err != nil {
						return err
					}

					u.Load(b.Bytes())
					u.Update(overrides)

					if !(c.Bool("print") || c.Bool("in-place")) {
						fmt.Print(u)
					}

					if c.Bool("print") {
						fmt.Print(chartutil.ToYaml(u))
					}

					if c.Bool("in-place") && u.HasChanged() {
						if err := os.WriteFile(path, []byte(chartutil.ToYaml(u)), info.Mode()); err != nil {
							return err
						}
					}
				}
				return nil
			})

			return err
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

type YAMLUpdater interface {
	Load([]byte) error
	Update(map[string]string)
	HasChanged() bool
}

func NewUpdater(targets []string) YAMLUpdater {
	m := make(map[string]interface{})

	return &VersionTree{
		targets:  targets,
		yaml:     nil,
		Versions: m,
	}
}

func NewImageUpdater(target string) YAMLUpdater {
	v := make(map[string]interface{})
	r := make(map[string][]map[string]interface{})

	return &VersionTreeR{
		target:            target,
		yaml:              nil,
		Versions:          v,
		versionReferences: r,
	}
}

type VersionTree struct {
	yaml     chartutil.Values
	targets  []string
	modified bool
	Versions map[string]interface{}
}

func (v *VersionTree) Load(b []byte) error {
	values, err := chartutil.ReadValues(b)
	if err != nil {
		return err
	}

	v.yaml = values
	for _, t := range v.targets {
		if val, ok := values[t]; ok {
			v.Versions[t] = val
		}
	}

	return nil
}

func (v *VersionTree) Update(overrides map[string]string) {
	m := v.yaml.AsMap()
	for _, t := range v.targets {
		if val, ok := overrides[t]; ok {
			v.modified = true
			m[t] = val
			v.Versions[t] = val
		}
	}
}

func (v *VersionTree) HasChanged() bool {
	return v.modified
}

func (v *VersionTree) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.yaml)
}

func (v *VersionTree) String() string {
	return chartutil.ToYaml(v.Versions)
}

type VersionTreeR struct {
	yaml              chartutil.Values
	target            string
	Versions          map[string]interface{}
	versionReferences map[string][]map[string]interface{}
	modified          bool
}

func (vr *VersionTreeR) Load(b []byte) error {
	values, err := chartutil.ReadValues(b)
	if err != nil {
		return err
	}

	vr.yaml = values
	targetLookup(vr.target, values.AsMap(), vr.versionReferences)

	for k, v := range vr.versionReferences {
		vr.Versions[k] = v[0][vr.target]
	}

	return nil
}

func (vr *VersionTreeR) Update(overrides map[string]string) {
	for k, _ := range vr.versionReferences {
		if val, ok := overrides[k]; ok {
			vr.modified = true
			vr.Versions[k] = val
			for _, ref := range vr.versionReferences[k] {
				ref[vr.target] = val
			}
		}
	}
}

func (vr *VersionTreeR) HasChanged() bool {
	return vr.modified
}

func (vr *VersionTreeR) MarshalJSON() ([]byte, error) {
	return json.Marshal(vr.yaml)
}

func (vr *VersionTreeR) String() string {
	return chartutil.ToYaml(vr.Versions)
}

func targetLookup(target string, tree map[string]interface{}, dictionary map[string][]map[string]interface{}) {
	if _, found := tree[target]; !found {
		for _, v := range tree {
			if vv, ok := v.(map[string]interface{}); ok {
				targetLookup(target, vv, dictionary)
			}
		}
		return
	}

	var relative string
	for k, v := range tree {
		if k == target {
			continue
		}
		relative = v.(string)
	}

	l := dictionary[relative]
	dictionary[relative] = append(l, tree)
}
