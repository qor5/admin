package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/gobuffalo/flect"
	"github.com/pkg/errors"
	"github.com/sergi/go-diff/diffmatchpatch"
	"mvdan.cc/gofumpt/format"
)

var (
	sampleFile = flag.String("sample-file", "", "Path to the sample file")
	patchFile  = flag.String("patch-file", "", "Path to the patch file, actually a base file for generating patches")
	backupDir  = flag.String("backup-dir", "", "Backup directory if output file exists")
	outputFile = flag.String("output-file", "", "Path to the output file")
)

func main() {
	flag.Parse()

	if err := run(*sampleFile, *patchFile, *backupDir, *outputFile); err != nil {
		log.Panicf("%+v", err)
	}
}

func run(sampleFile, patchFile, backupDir, outputFile string) (xerr error) {
	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
			xerr = errors.Wrap(err, "Panic")
		}
	}()

	// Read and parse the sample JSON file
	sample, err := os.ReadFile(sampleFile)
	if err != nil {
		return errors.WithStack(err)
	}
	var rrs []RequestResponse
	if err := json.Unmarshal(sample, &rrs); err != nil {
		return errors.WithStack(err)
	}

	// Generate the flow name from the file name
	filenameWithExt := filepath.Base(sampleFile)
	flowName := strings.TrimSuffix(filenameWithExt, filepath.Ext(filenameWithExt))
	flowName = flect.Pascalize(flowName)

	// Generate steps for each request-response pair
	var steps []*Step
	for i, rr := range rrs {
		v, err := generateStep(rr)
		if err != nil {
			return err
		}
		v.FlowName = flowName
		v.FuncName = fmt.Sprintf("flow%s_Step%02d_%s", flowName, i, v.FuncName)
		steps = append(steps, v)
	}

	// Prepare data for output template
	data := struct {
		FlowName string
		Steps    []*Step
	}{
		FlowName: flowName,
		Steps:    steps,
	}

	// Generate and format output
	output, err := executeOutput(flowTemplate, data)
	if err != nil {
		return err
	}

	original, err := os.ReadFile(patchFile)
	if err != nil && !os.IsNotExist(err) {
		return errors.WithStack(err)
	}

	// Read existing output for possible updates
	edited, err := os.ReadFile(outputFile)
	if err != nil {
		if os.IsNotExist(err) {
			// Write original output and record as a patch base for future use
			if err := os.WriteFile(patchFile, []byte(output), 0o644); err != nil {
				return errors.WithStack(err)
			}
			if err := os.WriteFile(outputFile, []byte(output), 0o644); err != nil {
				return errors.WithStack(err)
			}
			return nil
		}
		return errors.WithStack(err)
	}

	// Apply patch if target file exists and corresponding patch file is required
	if len(original) <= 0 {
		return errors.Errorf("patch file not found: %s", patchFile)
	}

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(string(original), string(edited), false)
	patches := dmp.PatchMake(string(original), diffs)
	pacthed, bs := dmp.PatchApply(patches, output)
	var patcherrs []string
	for idx, b := range bs {
		if !b {
			patcherrs = append(patcherrs, fmt.Sprintf("patch failed: %s", patches[idx].String()))
		}
	}
	if len(patcherrs) > 0 {
		patcherrs = append(patcherrs, "pacthed: "+pacthed)
		return errors.New(strings.Join(patcherrs, "\n"))
	}

	// Backup file before making changes
	backupFile := filepath.Join(backupDir, filepath.Base(outputFile)+fmt.Sprintf(".%s.backup", time.Now().Format(time.RFC3339)))
	if err := os.WriteFile(backupFile, []byte(edited), 0o644); err != nil {
		return errors.WithStack(err)
	}

	// Record original reference content as a patch base for future use
	if err := os.WriteFile(patchFile, []byte(output), 0o644); err != nil {
		return errors.WithStack(err)
	}

	output = pacthed

	// Write to the output file
	if err := os.WriteFile(outputFile, []byte(output), 0o644); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// Generate and format output using a template
func executeOutput(templateText string, data any) (string, error) {
	t, err := template.New("testflow").Parse(templateText)
	if err != nil {
		return "", errors.WithStack(err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", errors.WithStack(err)
	}

	// Format the generated code using gofumpt
	formattedCode, err := format.Source(buf.Bytes(), format.Options{})
	if err != nil {
		return "", errors.Wrap(err, "gofumpt format")
	}

	return string(formattedCode), nil
}
