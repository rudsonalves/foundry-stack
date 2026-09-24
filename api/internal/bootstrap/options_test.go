package bootstrap

import "testing"

func TestParseOptions(t *testing.T) {
	tests := []struct {
		name            string
		flag            string
		wantEnvironment string
		wantEnvFile     string
	}{
		{"selects development", "-dev", EnvDevelopment, ".env.dev"},
		{"selects staging", "-stag", EnvStaging, ".env.stag"},
		{"selects production", "-prod", EnvProduction, ".env.prod"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt, err := ParseOptions([]string{tt.flag})
			if err != nil {
				t.Fatalf("ParseOptions() unexpected error: %v", err)
			}

			if opt.Debug {
				t.Errorf("Debug = true, want false")
			}
			if opt.Environment != tt.wantEnvironment || opt.EnvFile != tt.wantEnvFile {
				t.Errorf(
					"profile = (%q, %q), want (%q, %q)",
					opt.Environment,
					opt.EnvFile,
					tt.wantEnvironment,
					tt.wantEnvFile,
				)
			}
		})
	}

	t.Run("enables debug when flag is present", func(t *testing.T) {
		opt, err := ParseOptions([]string{"-dev", "-debug"})
		if err != nil {
			t.Fatalf("ParseOptions() unexpected error: %v", err)
		}

		if !opt.Debug {
			t.Errorf("Debug = false, want true")
		}
	})

	t.Run("returns error for unknown flag", func(t *testing.T) {
		if _, err := ParseOptions([]string{"--unknown"}); err == nil {
			t.Error("ParseOptions() expected error for unknown flag, got nil")
		}
	})

	t.Run("requires one environment", func(t *testing.T) {
		if _, err := ParseOptions(nil); err == nil {
			t.Error("ParseOptions() expected error when environment is missing")
		}
	})

	t.Run("rejects multiple environments", func(t *testing.T) {
		combinations := [][]string{
			{"-dev", "-stag"},
			{"-dev", "-prod"},
			{"-stag", "-prod"},
		}

		for _, args := range combinations {
			if _, err := ParseOptions(args); err == nil {
				t.Errorf("ParseOptions(%v) expected error for multiple environments", args)
			}
		}
	})

	t.Run("rejects positional arguments", func(t *testing.T) {
		if _, err := ParseOptions([]string{"-dev", "unexpected"}); err == nil {
			t.Error("ParseOptions() expected error for positional argument")
		}
	})
}
