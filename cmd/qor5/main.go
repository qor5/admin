package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
)

//go:embed admin-template
var adminBox embed.FS

//go:embed bare-template
var bareBox embed.FS

//go:embed website-template
var websiteBox embed.FS

const TIPS = "\nRun the following command to start your App:"

func main() {
	validateFileExists := func(input string) error {
		dir := filepath.Base(input)
		_, err := os.Stat(dir)
		if err == nil {
			return fmt.Errorf("%s already exists, remove it first to generate.\n", err)
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:     "Go Package",
		AllowEdit: true,
		Default:   "github.com/theplant/myadmin",
	}

	pkg, err := prompt.Run()
	if err != nil {
		panic(err)
	}

	if err = validateFileExists(pkg); err != nil {
		sel := promptui.Select{
			Label: "Directory exists, Overwrite?",
			Items: []string{
				"Yes",
				"No",
			},
		}
		result, _, _ := sel.Run()
		if result != 0 {
			return
		}
	}

	templateSel := promptui.Select{
		Label: "Select a template",
		Items: []string{
			"Admin: Content Management System",
			"Website: Content Management System with Website Examples",
			"Bare: Simplest Workable Web App",
		},
	}

	result, _, _ := templateSel.Run()

	dir := filepath.Base(pkg)

	err = os.MkdirAll(dir, 0o755)
	if err != nil {
		panic(err)
	}

	switch result {
	case 0:
		copyAndReplaceFiles(adminBox, dir, "admin-template", pkg)
		fmt.Println(TIPS)
		color.Magenta("cd %s && docker-compose up -d && source dev_env && go run main.go\n", dir)
	case 1:
		copyAndReplaceFiles(websiteBox, dir, "website-template", pkg)
		fmt.Println(TIPS)
		color.Magenta("cd %s && docker-compose up -d && source dev_env && go run main.go\n", dir)
	case 2:
		copyAndReplaceFiles(bareBox, dir, "bare-template", pkg)
		fmt.Println(TIPS)
		color.Magenta("cd %s && go run main.go\n", dir)
	default:
		panic(fmt.Errorf("wrong option"))
	}
}

func copyAndReplaceFiles(box embed.FS, dir string, template string, pkg string) {
	var err error
	fs.WalkDir(box, template, func(path string, d fs.DirEntry, err1 error) error {
		if d != nil && d.IsDir() {
			return nil
		}
		newPath := strings.ReplaceAll(path, template+"/", "")
		fp := filepath.Join(dir, newPath)
		err := os.MkdirAll(filepath.Dir(fp), 0o755)
		if err != nil {
			panic(err)
		}
		var f *os.File
		f, err = os.Create(fp)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		content, err := box.ReadFile(path)
		if err != nil {
			panic(err)
		}

		err = os.WriteFile(fp, []byte(content), 0o644)
		if err != nil {
			panic(err)
		}
		fmt.Println(fp, "generated")
		return err
	})

	fmt.Println("Done")

	replaceInFiles(dir, "github.com/qor5/admin/v3/docs/cmd/qor5/"+template, pkg)
	replaceInFiles(dir, "QOR5PackageName", dir)

	if _, err = os.Stat(filepath.Join(dir, "go.mod")); err != nil {
		runCmd(dir, "go", "mod", "init", pkg)
		runCmd(dir, "go", "get", "./...")
	}
}

func runCmd(dir string, name string, args ...string) {
	cmdGet := exec.Command(name, args...)
	cmdGet.Dir = dir
	cmdGet.Stdout = os.Stdout
	cmdGet.Stderr = os.Stderr

	err := cmdGet.Run()
	if err != nil {
		panic(err)
	}
}

func replaceInFiles(baseDir string, from, to string) {
	var fileList []string
	err := filepath.Walk(baseDir, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			fileList = append(fileList, path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	for _, file := range fileList {
		replaceInFile(file, from, to)
	}
}

func replaceInFile(filepath, from, to string) {
	read, err := os.ReadFile(filepath)
	if err != nil {
		panic(err)
	}

	newContents := strings.ReplaceAll(string(read), from, to)

	// fmt.Println(newContents)

	err = os.WriteFile(filepath, []byte(newContents), 0)
	if err != nil {
		panic(err)
	}
}
