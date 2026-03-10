package mg

import (
	"testing"
)

func TestValidTargetColor(t *testing.T) {
	t.Setenv(EnableColorEnv, "true")
	t.Setenv(TargetColorEnv, "Yellow")
	expected := "\u001b[33m"
	if actual := TargetColor(); actual != expected {
		t.Fatalf("expected %v but got %s", expected, actual)
	}
}

func TestValidTargetColorCaseInsensitive(t *testing.T) {
	t.Setenv(EnableColorEnv, "true")
	t.Setenv(TargetColorEnv, "rED")
	expected := "\u001b[31m"
	if actual := TargetColor(); actual != expected {
		t.Fatalf("expected %v but got %s", expected, actual)
	}
}

func TestInvalidTargetColor(t *testing.T) {
	t.Setenv(EnableColorEnv, "true")
	// NOTE: Brown is not a defined Color constant
	t.Setenv(TargetColorEnv, "Brown")
	expected := DefaultTargetAnsiColor
	if actual := TargetColor(); actual != expected {
		t.Fatalf("expected %v but got %s", expected, actual)
	}
}
