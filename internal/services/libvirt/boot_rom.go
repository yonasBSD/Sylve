// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package libvirt

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	vmModels "github.com/alchemillahq/sylve/internal/db/models/vm"
	libvirtServiceInterfaces "github.com/alchemillahq/sylve/internal/interfaces/services/libvirt"
)

const (
	uefiFirmwarePath  = "/usr/local/share/uefi-firmware/BHYVE_UEFI.fd"
	ubootFirmwarePath = "/usr/local/share/u-boot/u-boot-bhyve-arm64/u-boot.bin"
)

func hostLibvirtArch() string {
	switch runtime.GOARCH {
	case "amd64":
		return "x86_64"
	case "arm64":
		return "aarch64"
	default:
		return runtime.GOARCH
	}
}

func hostFirmwarePath() string {
	switch runtime.GOARCH {
	case "arm64":
		return ubootFirmwarePath
	default:
		return uefiFirmwarePath
	}
}

func hostUsesSplitFirmware() bool {
	return runtime.GOARCH != "arm64"
}

func availableBootROMs() []vmModels.VMBootROM {
	if runtime.GOARCH == "arm64" {
		return []vmModels.VMBootROM{vmModels.VMBootROMUBoot, vmModels.VMBootROMNone}
	}
	return []vmModels.VMBootROM{vmModels.VMBootROMUEFI, vmModels.VMBootROMNone}
}

func normalizeBootROMValue(value vmModels.VMBootROM) vmModels.VMBootROM {
	switch strings.TrimSpace(strings.ToLower(string(value))) {
	case "":
		if runtime.GOARCH == "arm64" {
			return vmModels.VMBootROMUBoot
		}
		return vmModels.VMBootROMUEFI
	case string(vmModels.VMBootROMUEFI):
		return vmModels.VMBootROMUEFI
	case string(vmModels.VMBootROMUBoot):
		return vmModels.VMBootROMUBoot
	case string(vmModels.VMBootROMNone):
		return vmModels.VMBootROMNone
	default:
		return vmModels.VMBootROM(strings.TrimSpace(strings.ToLower(string(value))))
	}
}

func parseBootROMValue(value string) (vmModels.VMBootROM, error) {
	trimmed := strings.TrimSpace(strings.ToLower(value))

	if trimmed == "uefi_csm" {
		return "", fmt.Errorf("invalid_boot_rom: %s", trimmed)
	}

	normalized := normalizeBootROMValue(vmModels.VMBootROM(trimmed))

	switch normalized {
	case vmModels.VMBootROMNone:
		return normalized, nil
	case vmModels.VMBootROMUEFI:
		if runtime.GOARCH == "arm64" {
			return "", fmt.Errorf("uefi_firmware_not_available_on_arm64")
		}
		return normalized, nil
	case vmModels.VMBootROMUBoot:
		if runtime.GOARCH != "arm64" {
			return "", fmt.Errorf("uboot_only_available_on_arm64")
		}
		return normalized, nil
	default:
		return "", fmt.Errorf("invalid_boot_rom: %s", trimmed)
	}
}

func buildBootROMLoader(bootROM vmModels.VMBootROM, vmPath string, rid uint) *libvirtServiceInterfaces.Loader {
	switch normalizeBootROMValue(bootROM) {
	case vmModels.VMBootROMNone:
		return nil
	case vmModels.VMBootROMUBoot:
		return &libvirtServiceInterfaces.Loader{
			ReadOnly: "yes",
			Type:     "pflash",
			Path:     ubootFirmwarePath,
		}
	default:
		return &libvirtServiceInterfaces.Loader{
			ReadOnly: "yes",
			Type:     "pflash",
			Path:     fmt.Sprintf("%s,%s/%d_vars.fd", uefiFirmwarePath, vmPath, rid),
		}
	}
}

func (s *Service) ensureVMBootROMArtifacts(rid uint, bootROM vmModels.VMBootROM, vmPath string) error {
	if rid == 0 {
		return fmt.Errorf("invalid_rid")
	}

	normalized := normalizeBootROMValue(bootROM)
	if normalized == vmModels.VMBootROMNone {
		return nil
	}

	// u-boot is a single firmware binary — no per-VM VARS file to prepare
	if normalized == vmModels.VMBootROMUBoot {
		return nil
	}

	if strings.TrimSpace(vmPath) == "" {
		return fmt.Errorf("invalid_vm_path")
	}

	if err := os.MkdirAll(vmPath, 0755); err != nil {
		return fmt.Errorf("failed_to_ensure_vm_path_for_boot_rom: %w", err)
	}

	varsPath := filepath.Join(vmPath, fmt.Sprintf("%d_vars.fd", rid))
	if _, err := os.Stat(varsPath); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed_to_stat_uefi_vars: %w", err)
		}

		if err := s.ResetUEFIVars(rid); err != nil {
			return fmt.Errorf("failed_to_prepare_uefi_vars: %w", err)
		}
	}

	return nil
}
