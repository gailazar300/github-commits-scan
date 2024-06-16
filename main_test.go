package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestScanForAWSSecrets_NoSecrets(t *testing.T) {
	content := "This is a test content without any AWS secrets."
	accessKeys, secretKeys := scanForAWSSecrets(content)
	assert.Empty(t, accessKeys)
	assert.Empty(t, secretKeys)
}

func TestScanForAWSSecrets_WithSecrets(t *testing.T) {
	content := "This is a test content with AWS secrets AKIA1234567890123456 and DcCc9H6oCkGUSp3Rhmsx8NIfVG8kO2T/3jORxuZY."
	accessKeys, secretKeys := scanForAWSSecrets(content)
	assertContains(t, accessKeys, "AKIA1234567890123456")
	assertContains(t, secretKeys, "DcCc9H6oCkGUSp3Rhmsx8NIfVG8kO2T/3jORxuZY")
}

func assertContains(t *testing.T, slice []string, expectedValue string) {
	t.Helper()
	for _, v := range slice {
		if v == expectedValue {
			return
		}
	}
	t.Errorf("Expected value %s not found in slice %v", expectedValue, slice)
}
