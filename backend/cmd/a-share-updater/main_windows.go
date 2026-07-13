//go:build windows

package main

import (
	"archive/zip"
	"encoding/json"
	"flag"
	"fmt"
	"go-stock/backend/provider/update"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/sys/windows"
)

func main() {
	packagePath := flag.String("package", "", "verified update ZIP")
	installDir := flag.String("install-dir", "", "application directory")
	executable := flag.String("executable", "", "application executable name")
	statePath := flag.String("state", "", "pending update state path")
	pid := flag.Int("pid", 0, "application process ID")
	flag.Parse()
	if *packagePath == "" || *installDir == "" || *executable == "" || *statePath == "" || *pid <= 0 {
		fail("missing updater arguments")
	}
	if err := waitForProcess(*pid); err != nil {
		fail(err.Error())
	}
	backups, err := installPackage(*packagePath, *installDir)
	if err != nil {
		fail(err.Error())
	}
	executablePath := filepath.Join(*installDir, *executable)
	command := exec.Command(executablePath)
	if err := command.Start(); err != nil {
		restoreBackups(backups)
		fail("start updated application: " + err.Error())
	}
	if waitForHealthy(*statePath, 60*time.Second) {
		removeBackups(backups)
		_ = os.Remove(*statePath)
		return
	}
	_ = command.Process.Kill()
	restoreBackups(backups)
	_ = exec.Command(executablePath).Start()
}

func waitForProcess(pid int) error {
	handle, err := windows.OpenProcess(windows.SYNCHRONIZE, false, uint32(pid))
	if err != nil {
		return fmt.Errorf("open application process: %w", err)
	}
	defer windows.CloseHandle(handle)
	_, err = windows.WaitForSingleObject(handle, windows.INFINITE)
	return err
}

func installPackage(packagePath, installDir string) ([]string, error) {
	archive, err := zip.OpenReader(packagePath)
	if err != nil {
		return nil, err
	}
	defer archive.Close()
	stageDir, err := os.MkdirTemp(installDir, ".a-share-update-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(stageDir)
	for _, entry := range archive.File {
		name := filepath.Clean(entry.Name)
		if name == "." || strings.HasPrefix(name, "..") || filepath.IsAbs(name) {
			return nil, fmt.Errorf("unsafe update archive path: %s", entry.Name)
		}
		target := filepath.Join(stageDir, name)
		if !strings.HasPrefix(target, stageDir+string(os.PathSeparator)) && target != stageDir {
			return nil, fmt.Errorf("unsafe update archive path: %s", entry.Name)
		}
		if entry.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0755); err != nil {
				return nil, err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return nil, err
		}
		input, err := entry.Open()
		if err != nil {
			return nil, err
		}
		output, err := os.Create(target)
		if err == nil {
			_, err = io.Copy(output, input)
			closeErr := output.Close()
			if err == nil {
				err = closeErr
			}
		}
		_ = input.Close()
		if err != nil {
			return nil, err
		}
	}

	backups := make([]string, 0)
	err = filepath.Walk(stageDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil || info.IsDir() {
			return walkErr
		}
		relativePath, err := filepath.Rel(stageDir, path)
		if err != nil {
			return err
		}
		target := filepath.Join(installDir, relativePath)
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		backup := target + ".rollback"
		if _, err := os.Stat(target); err == nil {
			_ = os.Remove(backup)
			if err := os.Rename(target, backup); err != nil {
				return err
			}
			backups = append(backups, backup)
		}
		return os.Rename(path, target)
	})
	if err != nil {
		restoreBackups(backups)
		return nil, err
	}
	return backups, nil
}

func waitForHealthy(statePath string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		encoded, err := os.ReadFile(statePath)
		if err == nil {
			var state update.PendingUpdate
			if json.Unmarshal(encoded, &state) == nil && state.Healthy {
				return true
			}
		}
		time.Sleep(time.Second)
	}
	return false
}

func restoreBackups(backups []string) {
	for _, backup := range backups {
		target := strings.TrimSuffix(backup, ".rollback")
		_ = os.Remove(target)
		_ = os.Rename(backup, target)
	}
}

func removeBackups(backups []string) {
	for _, backup := range backups {
		_ = os.Remove(backup)
	}
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
