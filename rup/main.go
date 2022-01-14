package main

import (
	"bytes"
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
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "charts",
				Value:   "rke2-charts",
				Usage:   "rke2 charts directory",
				EnvVars: []string{"CHARTS"},
			},
			// TODO:
			&cli.BoolFlag{
				Name:    "debug",
				Value:   false,
				EnvVars: []string{"DEBUG"},
			},
			// TODO:
			&cli.BoolFlag{
				Name:    "in-place",
				Aliases: []string{"i"},
				Value:   false,
				EnvVars: []string{"DEBUG"},
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

			matchers := make(map[string]func([]byte, map[string]string) ([]byte, error))

			matchers[FileChart] = NewGenericUpdater([]string{"appVersion", "version"})
			matchers[FilePackage] = NewGenericUpdater([]string{"packageVersion"})
			matchers[FileValues] = updateImageValues

			filepath.Walk(chartPath, func(path string, info fs.FileInfo, err error) error {
				if info.IsDir() && strings.Contains(path, "generated") {
					return filepath.SkipDir
				}

				if info.IsDir() {
					return nil
				}

				var content []byte

				if fn, ok := matchers[filepath.Base(path)]; ok {
					f, err := os.Open(path)
					if err != nil {
						return err
					}
					defer f.Close()

					var b bytes.Buffer
					if _, err := io.Copy(&b, f); err != nil {
						return err
					}
					content, _ = fn(b.Bytes(), overrides)

					fmt.Print(string(content))
				}
				return nil
			})

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func NewGenericUpdater(targets []string) func([]byte, map[string]string) ([]byte, error) {
	return func(raw []byte, overrides map[string]string) ([]byte, error) {
		values, err := chartutil.ReadValues(raw)
		if err != nil {
			return nil, err
		}

		m := values.AsMap()
		for _, t := range targets {
			if v, ok := overrides[t]; ok {
				m[t] = v
			}
		}

		return []byte(chartutil.ToYaml(values)), nil
	}
}

func updateImageValues(raw []byte, overrides map[string]string) ([]byte, error) {
	values, err := chartutil.ReadValues(raw)
	if err != nil {
		return nil, err
	}

	m := values.AsMap()
	imageLookup(m, overrides)

	return []byte(chartutil.ToYaml(values)), nil
}

func imageLookup(tree map[string]interface{}, overrides map[string]string) {
	if _, found := tree["tag"]; !found {
		for _, v := range tree {
			if vv, ok := v.(map[string]interface{}); ok {
				imageLookup(vv, overrides)
			}
		}
		return
	}

	var imageName string
	for k, v := range tree {
		if k == "tag" {
			continue
		}
		imageName = v.(string)
	}

	if v, ok := overrides[imageName]; ok {
		tree["tag"] = v
	}
}
