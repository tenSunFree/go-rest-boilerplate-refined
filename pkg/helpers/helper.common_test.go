package helpers_test

import (
	"testing"

	"github.com/tenSunFree/luma-lang-go/pkg/helpers"
)

func TestIsArrayContains(t *testing.T) {
	// test case 1
	arr := []string{"hello", "world", "golang"}
	str := "golang"
	expected := true
	result := helpers.IsArrayContains(arr, str)
	if result != expected {
		t.Errorf("Expected %t but got %t", expected, result)
	}

	// test case 2
	arr = []string{"hello", "world", "golang"}
	str = "java"
	expected = false
	result = helpers.IsArrayContains(arr, str)
	if result != expected {
		t.Errorf("Expected %t but got %t", expected, result)
	}
}

func TestMaskEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected string
	}{
		{"normal email", "patrick@example.com", "p******@example.com"},
		{"single char local part", "a@example.com", "*@example.com"},
		{"no at sign", "notanemail", "***"},
		{"empty string", "", "***"},
	}
	for _, tt := range tests {
		result := helpers.MaskEmail(tt.email)
		if result != tt.expected {
			t.Errorf("%s: expected %q but got %q", tt.name, tt.expected, result)
		}
	}
}
