package main

import (
	"path/filepath"
	"testing"
)

func TestProcessPathWithValidPackageJSON(t *testing.T) {
	// Test cases
	path, err := filepath.Abs("./testdata/ValidPackageJSON")
	if err != nil {
		t.Error(err)
	}
	pJson, str, err := ProcessPath(path)
	if err != nil {
		t.Error(err)
	}
	if pJson == nil {
		t.Error("pJson is nil")
	}
	if str == "" {
		t.Error("str is empty")
	}
}

func TestProcessPathWithInvalidPackageJSON(t *testing.T) {
	// Test cases
	path, err := filepath.Abs("./testdata/InvalidPackageJSON")
	if err != nil {
		t.Error(err)
	}
	pJson, str, err := ProcessPath(path)
	if err == nil {
		t.Error("This should have thrown an error")
	}
	if pJson != nil {
		t.Error("pJson is not nil")
	}
	if str != "" {
		t.Error("str is not empty")
	}
}

func TestProcessPathWithFaultyPath(t *testing.T) {
	// Test cases
	path, err := filepath.Abs("./testdata/MissingPackageJSON")
	if err != nil {
		t.Error(err)
	}
	pJson, str, err := ProcessPath(path, 1)
	if err == nil {
		t.Error("This should have thrown an error")
	}
	if pJson != nil {
		t.Error("pJson is not nil")
	}
	if str != "" {
		t.Error("str is not empty")
	}
}
