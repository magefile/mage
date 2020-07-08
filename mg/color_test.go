package mg

import (
	"os"
	"testing"
)

func TestValidTargetColor(t *testing.T) {
	os.Setenv(EnableColorEnv, "true")
	os.Setenv(TargetColorEnv, "Yellow")
	expected := "\u001b[33m"
	if actual := TargetColor(); actual != expected {
		t.Fatalf("expected %v but got %s", expected, actual)
	}
}

func TestValidTargetColorCaseInsensitive(t *testing.T) {
	os.Setenv(EnableColorEnv, "true")
	os.Setenv(TargetColorEnv, "rED")
	expected := "\u001b[31m"
	if actual := TargetColor(); actual != expected {
		t.Fatalf("expected %v but got %s", expected, actual)
	}
}

func TestInvalidTargetColor(t *testing.T) {
	os.Setenv(EnableColorEnv, "true")
	// NOTE: Brown is not a defined Color constant
	os.Setenv(TargetColorEnv, "Brown")
	expected := DefaultTargetAnsiColor
	if actual := TargetColor(); actual != expected {
		t.Fatalf("expected %v but got %s", expected, actual)
	}
}
