package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

const (
	rke2Repository = "https://github.com/rancher/rke2-charts.git"
	rke2Local      = "./charts"
)

// pathToDiffCmd, err := exec.LookPath("diff")
// if err != nil {
// 	return false, fmt.Errorf("Cannot generate patch file if GNU diff is not available")
// }

func ensureRepository(remoteURL, localDir string) error {
	if remoteURL == "" {
		remoteURL = rke2Repository
	}

	if localDir == "" {
		localDir = rke2Local
	}

	if _, err := os.Stat(localDir); !os.IsNotExist(err) {
		return nil
	}

	args := strings.Split(`clone -o upstream --depth 1`, " ")
	args = append(args, remoteURL, localDir)

	cmd := exec.Command("git", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return errors.Wrap(err, out.String())
	}

	return nil
}

func listCharts(localDir string) ([]string, error) {
	if localDir == "" {
		localDir = rke2Local
	}

	fd, err := os.Open(filepath.Join(localDir, "packages"))
	if err != nil {
		return nil, err
	}

	files, err := fd.Readdir(0)
	if err != nil {
		return nil, err
	}

	results := []string{}
	for _, f := range files {
		if f.IsDir() {
			results = append(results, f.Name())
		}
	}

	return results, nil
}

func getVersionsFor(localDir, chartName string) (map[string]string, error) {
	if localDir == "" {
		localDir = rke2Local
	}

	chartPath := filepath.Join(localDir, "packages", chartName)
	if _, err := os.Stat(chartPath); err != nil {
		return nil, err
	}

	// find "./charts/packages/$PACKAGE" ! -name '*json*' ! -name '*.patch' -type f -exec grep -E '^\+?\s*(version|tag|appVersion|packageVersion|repository)\s*:\s*.\w+' {} \;
	type foo struct {
		file, targetRegexp string
		behavior           func(string, string, map[string]string)
	}

	pkg := foo{
		file:         "package.yaml",
		targetRegexp: "url|packageVersion",
	}

	chart := foo{
		file:         "Chart.yaml",
		targetRegexp: "^\\s*(version|appVersion)",
	}

	values := foo{
		file:         "values.yaml",
		targetRegexp: "tag",
		behavior: func(tag, repository string, versions map[string]string) {
			c := strings.Index(tag, ":") + 1
			l := strings.Index(repository, ":") + 1
			versions[strings.TrimSpace(repository[l:])] = strings.TrimSpace(tag[c:])
		},
	}

	alist := []foo{pkg, chart, values}
	versions := make(map[string]string)

	filepath.Walk(chartPath, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() && strings.Contains(path, "generated") {
			return filepath.SkipDir
		}

		if info.IsDir() {
			return nil
		}

		for _, v := range alist {
			if strings.Contains(path, v.file) {
				scanVersions(path, v.targetRegexp, v.behavior, versions)
			}
		}

		return nil
	})

	return versions, nil
}

// var versionNotFound = errors.New("version metadata not found")

func scanVersions(path, versionPattern string, b func(string, string, map[string]string), versions map[string]string) (map[string]string, error) {

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	matcher, err := regexp.Compile(versionPattern)
	if err != nil {
		return nil, errors.Wrap(err, "compiling regex")
	}

	scanner := bufio.NewScanner(file)

	var lineNumber int
	var last string
	for scanner.Scan() {
		lineNumber++

		line := scanner.Text()

		if !matcher.MatchString(line) {
			last = line
			continue
		}

		if b != nil {
			b(line, last, versions)
			last = line
			continue
		}

		l := strings.Split(line, ":")
		if len(l) > 1 {
			kind := strings.TrimSpace(l[0])
			value := strings.TrimSpace(l[1])
			versions[kind] = value
		}

		last = line
	}

	return versions, nil
}

func prettyPrint(w io.Writer, v interface{}) (err error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err == nil {
		fmt.Fprintln(w, string(b))
	}
	return
}
