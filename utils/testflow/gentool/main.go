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
	sampleFile = flag.String("sample-file", "", "sample file ")
	patchFile  = flag.String("patch-file", "", "patch file, it is actually a base file for generating patches")
	backupDir  = flag.String("backup-dir", "", "backup dir if output file exists")
	outputFile = flag.String("output-file", "", "output file ")
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

	sample, err := os.ReadFile(sampleFile)
	if err != nil {
		return errors.WithStack(err)
	}
	var rrs []RequestResponse
	if err := json.Unmarshal(sample, &rrs); err != nil {
		return errors.WithStack(err)
	}

	filenameWithExt := filepath.Base(sampleFile)
	flowName := strings.TrimSuffix(filenameWithExt, filepath.Ext(filenameWithExt))
	flowName = flect.Pascalize(flowName)

	steps := []*Step{}
	for i, rr := range rrs {
		v, err := generateStep(rr)
		if err != nil {
			return err
		}
		v.FlowName = flowName
		v.FuncName = fmt.Sprintf("flow%s_Step%02d_%s", flowName, i, v.FuncName)
		steps = append(steps, v)
	}

	data := struct {
		FlowName string
		Steps    []*Step
	}{
		FlowName: flowName,
		Steps:    steps,
	}

	// 生成并格式化结果
	output, err := executeOutput(flowTemplate, data)
	if err != nil {
		return err
	}

	original, err := os.ReadFile(patchFile)
	if err != nil && !os.IsNotExist(err) {
		return errors.WithStack(err)
	}

	// 获取原内容，即为我们可能已修改的内容
	edited, err := os.ReadFile(outputFile)
	if err != nil {
		// 若不存在则直接输出和记录patch
		if os.IsNotExist(err) {
			// 记录原始内容作为 patch 基础文件，以做下次使用
			if err := os.WriteFile(patchFile, []byte(output), 0o644); err != nil {
				return errors.WithStack(err)
			}

			// 写入目标文件
			if err := os.WriteFile(outputFile, []byte(output), 0o644); err != nil {
				return errors.WithStack(err)
			}

			return nil
		}

		return errors.WithStack(err)
	}

	// 如果目标文件存在，则应用 patch，必须存在对应的 patch 文件用作更新
	if len(original) <= 0 {
		return errors.Errorf("patch file not found: %s", patchFile)
	}

	dmp := diffmatchpatch.New()

	// 计算差异
	diffs := dmp.DiffMain(string(original), string(edited), false)
	// log.Println("Diffs:", diffs)

	// 即时生成补丁，其实 patch-file 是 patch-base 文件
	patches := dmp.PatchMake(string(original), diffs)
	// log.Println("Patches:", patches)

	// 对新生成的参考代码应用补丁，这样就能防止丢失我们自己的修改
	pacthed, bs := dmp.PatchApply(patches, output)
	patcherrs := []string{}
	for idx, b := range bs {
		if !b {
			patcherrs = append(patcherrs, fmt.Sprintf("patch failed: %v", patches[idx]))
		}
	}
	if len(patcherrs) > 0 {
		return errors.New(strings.Join(patcherrs, "\n"))
	}

	// 备份修改前文件以防止意外
	backupFile := filepath.Join(backupDir, filepath.Base(outputFile)+fmt.Sprintf(".%s.backup", time.Now().Format(time.RFC3339)))
	if err := os.WriteFile(backupFile, []byte(edited), 0o644); err != nil {
		return errors.WithStack(err)
	}

	// 记录原始参考内容作为 patch 基础文件，以做下次使用
	if err := os.WriteFile(patchFile, []byte(output), 0o644); err != nil {
		return errors.WithStack(err)
	}

	output = pacthed

	// 写入目标文件
	if err := os.WriteFile(outputFile, []byte(output), 0o644); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// 生成并格式化结果
func executeOutput(templateText string, data any) (string, error) {
	// 使用 template 生成代码到 buffer
	t, err := template.New("testflow").Parse(templateText)
	if err != nil {
		return "", errors.WithStack(err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", errors.WithStack(err)
	}

	// 使用 gofumpt 格式化代码
	formattedCode, err := format.Source(buf.Bytes(), format.Options{})
	if err != nil {
		return "", errors.Wrap(err, "gofumpt format")
	}

	return string(formattedCode), nil
}
