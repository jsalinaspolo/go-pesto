package main

import (
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestValidateArgs(t *testing.T) {
	args := []string{
		"argument0",
		"token",
		"environment",
		"namespace",
		"app1,app2",
		"version",
	}

	pestoArgs, err := validateArgs(args)
	require.NoError(t, err, "unexpected error when reading workload manifest")

	expectedHelm := PestoArgs{
		Token:        "token",
		Environment:  "environment",
		Namespace:    "namespace",
		Applications: []string{"app1", "app2"},
		Version:      "version"}

	if diff := cmp.Diff(*pestoArgs, expectedHelm); diff != "" {
		t.Errorf("PestoArgs() mismatch (-pestoArgs +expectedHelm):\n%s", diff)
	}
}

func TestMissingAnArg(t *testing.T) {
	args := []string{
		"argument0",
		"token",
		"environment",
		"namespace",
		"app1,app2",
	}

	_, err := validateArgs(args)
	expectedMessage := "Arguments 5 different than 6"

	if err != nil && err.Error() != expectedMessage {
		t.Fatalf("Expected error message missmatch: %s", err)
	}
}

func TestHelmChartFiles(t *testing.T) {
	pestoArgs := PestoArgs{
		Token:        "token",
		Environment:  "environment",
		Namespace:    "namespace",
		Applications: []string{"app1", "app2"},
		Version:      "version"}

	expectedFiles := []string{"environment/namespace/app1", "environment/namespace/app2"}

	if diff := cmp.Diff(pestoArgs.files(), expectedFiles); diff != "" {
		t.Errorf("Failed matching files (-pestoArgs.files() +expectedFiles):\n%s", diff)
	}
}
