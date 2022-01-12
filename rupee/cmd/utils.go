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

// TODO: Look up for required binaries
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

type source struct {
	file, match string
	parse       func(*Token, *Token, map[string]*Token)
}

func newLine(source, content string, line, col int) *Token {
	return &Token{
		Value:  content,
		source: source,
		line:   line,
		col:    col,
	}
}

type Token struct {
	Value  string
	source string
	line   int
	col    int
}

func (l Token) String() string {
	return l.Value
}

func (l Token) MarshalJSON() ([]byte, error) {
	return json.Marshal(l.Value)
}

// find "./charts/packages/$PACKAGE" ! -name '*json*' ! -name '*.patch' -type f -exec grep -E '^\+?\s*(version|tag|appVersion|packageVersion|repository)\s*:\s*.\w+' {} \;
func getVersionsFor(localDir, chartName string) (map[string]*Token, error) {
	if localDir == "" {
		localDir = rke2Local
	}

	chartPath := filepath.Join(localDir, "packages", chartName)
	if _, err := os.Stat(chartPath); err != nil {
		return nil, err
	}

	pkg := source{
		file:  "package.yaml",
		match: "^\\s*(url|packageVersion)\\s*:",
		parse: func(current, _ *Token, versions map[string]*Token) {
			line := current.Value

			i := strings.Index(line, ":")

			if strings.Contains(line[i+1:], "local") {
				return
			}

			r := regexp.MustCompile(`v?\d+(\.\d+)*`)
			current.Value = r.FindString(line)

			versions[strings.TrimSpace(line[0:i])] = current
		},
	}

	chart := source{
		file:  "Chart.yaml",
		match: "^\\s*(version|appVersion)\\s*:",
	}

	values := source{
		file:  "values.yaml",
		match: "^\\s*tag\\s*:",
		parse: func(tag, repository *Token, versions map[string]*Token) {

			// FIXME: As it is, it does not support multiple images with the same repo. name
			c := strings.Index(tag.Value, ":") + 1
			l := strings.Index(repository.Value, ":") + 1
			tag.Value = strings.TrimSpace(tag.Value[c:])

			versions[strings.TrimSpace(repository.Value[l:])] = tag
		},
	}

	alist := []source{pkg, chart, values}
	versions := make(map[string]*Token)

	filepath.Walk(chartPath, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() && strings.Contains(path, "generated") {
			return filepath.SkipDir
		}

		if info.IsDir() {
			return nil
		}

		for _, v := range alist {
			if strings.Contains(path, v.file) {
				scanVersions(path, v.match, v.parse, versions)
			}
		}

		return nil
	})

	return versions, nil
}

func scanVersions(path, versionPattern string, p func(*Token, *Token, map[string]*Token), versions map[string]*Token) (map[string]*Token, error) {

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
	var last *Token
	for scanner.Scan() {
		lineNumber++

		t := scanner.Text()
		l := newLine(path, t, lineNumber, 0)

		if !matcher.MatchString(t) {
			last = l
			continue
		}

		if p != nil {
			p(l, last, versions)
			last = l
			continue
		}

		tt := strings.Index(t, ":")
		if tt > 0 {
			kind := t[:tt]
			value := strings.TrimSpace(t[tt+1:])
			versions[kind] = newLine(path, value, lineNumber, tt+1)
		}

		last = l
	}

	return versions, nil
}

func setVersion(target *Token, value string) ([]byte, error) {
	file, err := os.Open(target.source)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	var lineNumber int

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lineNumber++

		if lineNumber != target.line {
			fmt.Fprintln(&buf, scanner.Text())
			continue
		}

		t := scanner.Text()
		fmt.Fprintln(&buf, strings.ReplaceAll(t, target.Value, value))
	}

	return buf.Bytes(), nil
}

func prettyPrint(w io.Writer, v interface{}) (err error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err == nil {
		fmt.Fprintln(w, string(b))
	}
	return
}
