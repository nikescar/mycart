package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaintenanceEnable(t *testing.T) {
	// Ensure clean state
	os.Remove("maintenance.flag")

	// Execute enable command
	cmd := cmdMaintenanceEnable()
	cmd.SetArgs([]string{})
	err := cmd.Execute()

	assert.NoError(t, err)

	// Verify file created
	_, err = os.Stat("maintenance.flag")
	assert.NoError(t, err, "maintenance.flag file should exist")

	// Cleanup
	os.Remove("maintenance.flag")
}

func TestMaintenanceDisable(t *testing.T) {
	// Setup: create maintenance.flag file
	f, err := os.Create("maintenance.flag")
	assert.NoError(t, err)
	f.Close()

	// Execute disable command
	cmd := cmdMaintenanceDisable()
	cmd.SetArgs([]string{})
	err = cmd.Execute()

	assert.NoError(t, err)

	// Verify file removed
	_, err = os.Stat("maintenance.flag")
	assert.True(t, os.IsNotExist(err), "maintenance.flag file should not exist")
}

func TestMaintenanceStatus(t *testing.T) {
	// Test when disabled
	os.Remove("maintenance.flag")
	cmd := cmdMaintenanceStatus()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	assert.NoError(t, err)

	// Test when enabled
	f, err := os.Create("maintenance.flag")
	assert.NoError(t, err)
	f.Close()
	defer os.Remove("maintenance.flag")

	cmd = cmdMaintenanceStatus()
	cmd.SetArgs([]string{})
	err = cmd.Execute()
	assert.NoError(t, err)
}
