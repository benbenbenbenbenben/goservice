package platform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type WindowsService struct {
	ExecutablePath string
	ServiceDisplay string
	ServiceDesc    string
}

func (s *WindowsService) Install() error {
	installDir := s.GetInstallDir()
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create installation directory: %w", err)
	}

	installedBinary := filepath.Join(installDir, "bin", filepath.Base(s.ExecutablePath))
	if err := os.MkdirAll(filepath.Dir(installedBinary), 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	if err := copyFile(s.ExecutablePath, installedBinary); err != nil {
		return fmt.Errorf("failed to copy binary: %w", err)
	}

	cmd := exec.Command("sc", "create", s.ServiceName(),
		"binPath=", fmt.Sprintf("\"%s\" -run", installedBinary),
		"DisplayName=", s.ServiceDisplay,
		"start=", "auto",
		"obj=", "LocalSystem")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	descCmd := exec.Command("sc", "description", s.ServiceName(), s.ServiceDesc)
	if err := descCmd.Run(); err != nil {
		return fmt.Errorf("failed to set service description: %w", err)
	}

	if err := exec.Command("sc", "start", s.ServiceName()).Run(); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}
	return nil
}

func (s *WindowsService) Uninstall() error {
	_ = exec.Command("sc", "stop", s.ServiceName()).Run()
	if err := exec.Command("sc", "delete", s.ServiceName()).Run(); err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}

	// Clean up installation directory
	installDir := s.GetInstallDir()
	if err := os.RemoveAll(installDir); err != nil {
		return fmt.Errorf("failed to remove installation directory: %w", err)
	}
	return nil
}
func (s *WindowsService) Status() (bool, error) {
	output, err := exec.Command("sc", "query", s.ServiceName()).Output()
	if err != nil {
		return false, nil
	}
	return strings.Contains(string(output), "RUNNING"), nil
}

func (s *WindowsService) Start() error {
	if err := exec.Command("sc", "start", s.ServiceName()).Run(); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}
	return nil
}

func (s *WindowsService) Stop() error {
	if err := exec.Command("sc", "stop", s.ServiceName()).Run(); err != nil {
		return fmt.Errorf("failed to stop service: %w", err)
	}
	return nil
}

func (s *WindowsService) GetInstallDir() string {
	return filepath.Join(os.Getenv("ProgramData"), s.ServiceName())
}

func (s *WindowsService) ServiceName() string {
	return s.ServiceDisplay
}

func (s *WindowsService) ServiceDisplayName() string {
	return s.ServiceDisplay
}

func (s *WindowsService) ServiceDescription() string {
	return s.ServiceDesc
}
