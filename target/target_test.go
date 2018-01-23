package target_test

import (
	"testing"

	"github.com/magefile/mage/target"
)

func TestPath(t *testing.T) {
	tests := []struct {
		name    string
		dst     string
		sources []string
		want    bool
		werr    bool
	}{
		{
			"name",
			"dst",
			[]string{"one", "two"},
			false,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := target.Path(tt.dst, tt.sources...)
			if err != nil {
				t.Errorf("Path() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("Path() = %v, want %v", got, tt.want)
			}
		})
	}
}
