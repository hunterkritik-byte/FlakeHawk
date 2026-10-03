package junit

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/hunterkritik-byte/FlakeHawk/internal/model"
)

type testSuites struct {
	XMLName xml.Name `xml:"testsuites"`
	Suites []testSuite `xml:"testsuite"`
}

type testSuite struct {
	Name string `xml:"name,attr"`
	Timestamp string `xml:"timestamp,attr"`
	Tests []testCase `xml:"testcase"`
}

type testCase struct {
	Name string `xml:"name,attr"`
	ClassName string `xml:"classname,attr"`
	File string `xml:"file,attr"`
	Time string `xml:"time,attr"`
	Failure *xmlElement `xml:"failure"`
	Error *xmlElement `xml:"error"`
	Skipped *xmlElement `xml:"skipped"`
}

type xmlElement struct {
	Message string `xml:"message,attr"`
}

func ParseFile(path string) ([]model.TestExecution, error) {
	f, err := os.Open(path)
	if err != nil { return nil, fmt.Errorf("open JUnit file: %w", err) }
	defer f.Close()
	return Parse(f, path)
}

func Parse(r io.Reader, source string) ([]model.TestExecution, error) {
	data, err := io.ReadAll(r)
	if err != nil { return nil, fmt.Errorf("read JUnit XML: %w", err) }

	var root testSuites
	if err := xml.Unmarshal(data, &root); err != nil {
		var suite testSuite
		if err2 := xml.Unmarshal(data, &suite); err2 != nil {
			return nil, fmt.Errorf("decode JUnit XML: %w", err)
		}
		root.Suites = []testSuite{suite}
	}

	var out []model.TestExecution
	for _, suite := range root.Suites {
		ts := parseTimestamp(suite.Timestamp)
		for _, tc := range suite.Tests {
			status := model.StatusPassed
			switch {
			case tc.Skipped != nil:
				status = model.StatusSkipped
			case tc.Error != nil:
				status = model.StatusError
			case tc.Failure != nil:
				status = model.StatusFailed
			}
			out = append(out, model.TestExecution{
				TestID: stableID(tc.ClassName, tc.Name, tc.File),
				Name: tc.Name, ClassName: tc.ClassName, File: tc.File,
				Status: status, DurationMS: parseSeconds(tc.Time), Timestamp: ts,
			})
		}
	}
	return out, nil
}

func ParsePath(path string) ([]model.TestExecution, error) {
	info, err := os.Stat(path)
	if err != nil { return nil, err }
	if !info.IsDir() { return ParseFile(path) }

	var all []model.TestExecution
	err = filepath.Walk(path, func(p string, info os.FileInfo, walkErr error) error {
		if walkErr != nil { return walkErr }
		if info.IsDir() || !strings.HasSuffix(strings.ToLower(info.Name()), ".xml") { return nil }
		items, err := ParseFile(p)
		if err != nil { return err }
		all = append(all, items...)
		return nil
	})
	return all, err
}

func stableID(className, name, file string) string {
	if className != "" { return className + "::" + name }
	if file != "" { return file + "::" + name }
	return name
}

func parseSeconds(s string) int64 {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil || v < 0 { return 0 }
	return int64(v * 1000)
}

func parseTimestamp(s string) time.Time {
	if s == "" { return time.Time{} }
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02 15:04:05"} {
		if t, err := time.Parse(layout, s); err == nil { return t }
	}
	return time.Time{}
}
