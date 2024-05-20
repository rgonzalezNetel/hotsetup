package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Web        Section `toml:"web"`
	WebLibrary Library `toml:"web-library"`
	None       Section `toml:"none"`
}

type Section struct {
	Files []string `toml:"files"`
}

type Library struct {
	Libs []string `toml:"libs"`
}

func main() {
	mode := flag.String("mode", "", "Setup mode: web or none")
	flag.Parse()

	if err := executeMode(*mode); err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Setup completed successfully.")
}

func executeMode(mode string) error {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	configPath := filepath.Join(basepath, "config.toml")

	var config Config
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	switch mode {
	case "web":
		if err := setupProject(config.Web); err != nil {
			return err
		}
		if err := installLibraries(config.WebLibrary.Libs); err != nil {
			return err
		}
	case "none":
		if err := setupProject(config.None); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported mode. Use 'web' or 'none'")
	}
	return nil
}

func setupProject(section Section) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %w", err)
	}

	for _, file := range section.Files {
		dir, filename := filepath.Split(file)
		fullPath := filepath.Join(cwd, dir, filename)
		dirPath := filepath.Dir(fullPath)

		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("failed to create %s: %w", dirPath, err)
		}

		if err := os.WriteFile(fullPath, []byte(fmt.Sprintf("package %s\n", filepath.Base(dir))), 0644); err != nil {
			return fmt.Errorf("failed to write to %s: %w", fullPath, err)
		}
	}
	return nil
}

func installLibraries(libs []string) error {
	for _, lib := range libs {
		cmd := exec.Command("go", "get", lib)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install library %s: %w", lib, err)
		}
	}
	return nil
}
