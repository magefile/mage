package sh

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"
)

func TestOutCmd(t *testing.T) {
	cmd := OutCmd(os.Args[0], "-printArgs", "foo", "bar")
	out, err := cmd("baz", "bat")
	if err != nil {
		t.Fatal(err)
	}
	expected := "[foo bar baz bat]"
	if out != expected {
		t.Fatalf("expected %q but got %q", expected, out)
	}
}

func TestExitCode(t *testing.T) {
	ran, err := Exec(nil, nil, nil, os.Args[0], "-helper", "-exit", "99")
	if err == nil {
		t.Fatal("unexpected nil error from run")
	}
	if !ran {
		t.Errorf("ran returned as false, but should have been true")
	}
	code := ExitStatus(err)
	if code != 99 {
		t.Fatalf("expected exit status 99, but got %v", code)
	}
}

func TestEnv(t *testing.T) {
	env := "SOME_REALLY_LONG_MAGEFILE_SPECIFIC_THING"
	out := &bytes.Buffer{}
	ran, err := Exec(map[string]string{env: "foobar"}, out, nil, os.Args[0], "-printVar", env)
	if err != nil {
		t.Fatalf("unexpected error from runner: %#v", err)
	}
	if !ran {
		t.Errorf("expected ran to be true but was false.")
	}
	if out.String() != "foobar\n" {
		t.Errorf("expected foobar, got %q", out)
	}
}

func TestEnvCfg(t *testing.T) {
	testCases := []struct {
		name       string
		envvars    map[string]string
		envCfgMode CfgMode
		envCfgCmds []string
		cmdResults map[string]string
		wantError  error
	}{
		{
			name: "default config",
			envvars: map[string]string{
				"SOME_FOO": "bar",
				"SOME_FAZ": "baz",
			},
			cmdResults: map[string]string{
				"echo $SOME_FOO":   "bar",
				"echo $SOME_FAZ":   "baz",
				"printf $SOME_FOO": "bar",
				"printf $SOME_FAZ": "baz",
			},
		},
		{
			name:       "OnForAll with commands",
			envCfgMode: OnForAll,
			envCfgCmds: []string{"echo"},
			wantError:  errors.New("OnForAll cannot accept slice of commands"),
		},
		{
			name:       "OffForAll with commands",
			envCfgMode: OffForAll,
			envCfgCmds: []string{"echo", "ls"},
			wantError:  errors.New("OffForAll cannot accept slice of commands"),
		},
		{
			name:       "IncludeCmds without commands",
			envCfgMode: IncludeCmds,
			wantError:  errors.New("IncludeCmds expects a slice of commands to be passed"),
		},
		{
			name:       "ExcludeCmds without commands",
			envCfgMode: ExcludeCmds,
			wantError:  errors.New("ExcludeCmds expects a slice of commands to be passed"),
		},
		{
			name:       "OffForAll",
			envCfgMode: OffForAll,
			envvars: map[string]string{
				"SOME_FOO": "bar",
				"SOME_FAZ": "baz",
			},
			cmdResults: map[string]string{
				"echo $SOME_FOO":   "$SOME_FOO",
				"echo $SOME_FAZ":   "$SOME_FAZ",
				"printf $SOME_FOO": "$SOME_FOO",
				"printf $SOME_FAZ": "$SOME_FAZ",
			},
		},
		{
			name:       "IncludeCmds",
			envCfgMode: IncludeCmds,
			envCfgCmds: []string{"echo"},
			envvars: map[string]string{
				"SOME_FOO": "bar",
				"SOME_FAZ": "baz",
			},
			cmdResults: map[string]string{
				"echo $SOME_FOO":   "bar",
				"echo $SOME_FAZ":   "baz",
				"printf $SOME_FOO": "$SOME_FOO",
				"printf $SOME_FAZ": "$SOME_FAZ",
			},
		},
		{
			name:       "ExcludeCmds",
			envCfgMode: ExcludeCmds,
			envCfgCmds: []string{"echo"},
			envvars: map[string]string{
				"SOME_FOO": "bar",
				"SOME_FAZ": "baz",
			},
			cmdResults: map[string]string{
				"echo $SOME_FOO":   "$SOME_FOO",
				"echo $SOME_FAZ":   "$SOME_FAZ",
				"printf $SOME_FOO": "bar",
				"printf $SOME_FAZ": "baz",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var cfgerr error
			// Apply expand env config.
			if len(tc.envCfgCmds) > 0 {
				cfgerr = CfgExpandEnv(tc.envCfgMode, tc.envCfgCmds...)
			} else {
				cfgerr = CfgExpandEnv(tc.envCfgMode)
			}
			if tc.wantError != nil && cfgerr.Error() != tc.wantError.Error() {
				t.Errorf("unexpected error while setting expand env config: %#v", cfgerr)
			}

			// Set all the env vars.
			for k, v := range tc.envvars {
				if err := os.Setenv(k, v); err != nil {
					t.Fatal(err)
				}
			}

			// Run commands with the applied expand env config.
			for k, v := range tc.cmdResults {
				// Split the whole command string into command and args.
				cmd := strings.Split(k, " ")
				s, err := Output(cmd[0], cmd[1:]...)
				if err != nil {
					t.Fatal(err)
				}
				if s != v {
					t.Errorf("expected value of $%s to be %q, but got %q", k, v, s)
				}
			}
		})
	}

	// Empty the expandEnvConfig map to not affect other tests.
	expandEnvConfig = make(map[string]bool)
}

func TestNotRun(t *testing.T) {
	ran, err := Exec(nil, nil, nil, "thiswontwork")
	if err == nil {
		t.Fatal("unexpected nil error")
	}
	if ran {
		t.Fatal("expected ran to be false but was true")
	}
}

func TestAutoExpand(t *testing.T) {
	if err := os.Setenv("MAGE_FOOBAR", "baz"); err != nil {
		t.Fatal(err)
	}
	s, err := Output("echo", "$MAGE_FOOBAR")
	if err != nil {
		t.Fatal(err)
	}
	if s != "baz" {
		t.Fatalf(`Expected "baz" but got %q`, s)
	}

}
