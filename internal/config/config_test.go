package config

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet("", flag.ContinueOnError)
	os.Args = []string{"test",
		"-a", "localhost:9999",
		"-b", "http://localhost:9999",
		"-l", "info",
		"-f", "file.json",
		"-d", "host=localhost user=postgres password=1234 dbname=urlshort sslmode=disable"}
	flag.CommandLine.Parse(os.Args[1:])

	_, err := NewConfig()

	require.NoError(t, err)
}

func TestServerAddressPriority(t *testing.T) {
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	tests := []struct {
		name           string
		envVars        map[string]string
		flagArgs       []string
		expectedResult string
	}{
		{
			name: "Environment value has priority if set",
			envVars: map[string]string{
				"SERVER_ADDRESS": "localhost:9999",
			},
			flagArgs:       []string{"-a", "localhost:8888"},
			expectedResult: "localhost:9999",
		},
		{
			name:           "Flag value has priority if environment value is not set",
			envVars:        map[string]string{},
			flagArgs:       []string{"-a", "localhost:8888"},
			expectedResult: "localhost:8888",
		},
		{
			name:           "Default value if flag and value are not set",
			envVars:        map[string]string{},
			flagArgs:       []string{},
			expectedResult: "localhost:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key := range tt.envVars {
				os.Unsetenv(key)
			}

			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			if len(tt.flagArgs) > 0 {
				os.Args = append([]string{"cmd"}, tt.flagArgs...)
			} else {
				os.Args = []string{"cmd"}
			}

			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			cfg, err := NewConfig()
			require.NoError(t, err)
			require.Equal(t, tt.expectedResult, cfg.ServerAddr)
		})
	}
}

func TestBaseURLPriority(t *testing.T) {
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	tests := []struct {
		name           string
		envVars        map[string]string
		flagArgs       []string
		expectedResult string
	}{
		{
			name: "Environment value has priority if set",
			envVars: map[string]string{
				"BASE_URL": "http://localhost:9999",
			},
			flagArgs:       []string{"-b", "http://localhost:8888"},
			expectedResult: "http://localhost:9999",
		},
		{
			name:           "Flag value has priority if environment value is not set",
			envVars:        map[string]string{},
			flagArgs:       []string{"-b", "http://localhost:8888"},
			expectedResult: "http://localhost:8888",
		},
		{
			name:           "Default value if flag and value are not set",
			envVars:        map[string]string{},
			flagArgs:       []string{},
			expectedResult: "http://localhost:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key := range tt.envVars {
				os.Unsetenv(key)
			}

			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			if len(tt.flagArgs) > 0 {
				os.Args = append([]string{"cmd"}, tt.flagArgs...)
			} else {
				os.Args = []string{"cmd"}
			}

			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			cfg, err := NewConfig()
			require.NoError(t, err)
			require.Equal(t, tt.expectedResult, cfg.BaseURL)
		})
	}
}

func TestLogLevelPriority(t *testing.T) {
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	tests := []struct {
		name           string
		envVars        map[string]string
		flagArgs       []string
		expectedResult string
	}{
		{
			name: "Environment value has priority if set",
			envVars: map[string]string{
				"LOG_LEVEL": "warn",
			},
			flagArgs:       []string{"-l", "debug"},
			expectedResult: "warn",
		},
		{
			name:           "Flag value has priority if environment value is not set",
			envVars:        map[string]string{},
			flagArgs:       []string{"-l", "warn"},
			expectedResult: "warn",
		},
		{
			name:           "Default value if flag and value are not set",
			envVars:        map[string]string{},
			flagArgs:       []string{},
			expectedResult: "info",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key := range tt.envVars {
				os.Unsetenv(key)
			}

			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			if len(tt.flagArgs) > 0 {
				os.Args = append([]string{"cmd"}, tt.flagArgs...)
			} else {
				os.Args = []string{"cmd"}
			}

			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			cfg, err := NewConfig()
			require.NoError(t, err)
			require.Equal(t, tt.expectedResult, cfg.LogLevel)
		})
	}
}

func TestFileStoragePathPriority(t *testing.T) {
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	tests := []struct {
		name           string
		envVars        map[string]string
		flagArgs       []string
		expectedResult string
	}{
		{
			name: "Environment value has priority if set",
			envVars: map[string]string{
				"FILE_STORAGE_PATH": "priority.json",
			},
			flagArgs:       []string{"-f", "secondary.json"},
			expectedResult: "priority.json",
		},
		{
			name:           "Flag value has priority if environment value is not set",
			envVars:        map[string]string{},
			flagArgs:       []string{"-f", "priority.json"},
			expectedResult: "priority.json",
		},
		{
			name:           "Default value if flag and value are not set",
			envVars:        map[string]string{},
			flagArgs:       []string{},
			expectedResult: filepath.Join(os.TempDir(), "storage.json"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key := range tt.envVars {
				os.Unsetenv(key)
			}

			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			if len(tt.flagArgs) > 0 {
				os.Args = append([]string{"cmd"}, tt.flagArgs...)
			} else {
				os.Args = []string{"cmd"}
			}

			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			cfg, err := NewConfig()
			require.NoError(t, err)
			require.Equal(t, tt.expectedResult, cfg.FileStoragePath)
		})
	}
}

func TestDatabaseDSNPriority(t *testing.T) {
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	tests := []struct {
		name           string
		envVars        map[string]string
		flagName       string
		flagValue      string
		expectedResult string
	}{
		{
			name: "Environment variable has priority",
			envVars: map[string]string{
				"DATABASE_DSN": "env-dsn-value",
			},
			flagName:       "-d",
			flagValue:      "flag-dsn-value",
			expectedResult: "env-dsn-value",
		},
		{
			name:           "Flag is used when no env var",
			envVars:        map[string]string{},
			flagName:       "-d",
			flagValue:      "flag-dsn-value",
			expectedResult: "flag-dsn-value",
		},
		{
			name:           "Empty when neither env nor flag",
			envVars:        map[string]string{},
			flagName:       "",
			flagValue:      "",
			expectedResult: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key := range tt.envVars {
				os.Unsetenv(key)
			}

			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			if tt.flagName != "" {
				os.Args = []string{"cmd", tt.flagName, tt.flagValue}
			} else {
				os.Args = []string{"cmd"}
			}

			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			cfg, err := NewConfig()
			require.NoError(t, err)
			require.Equal(t, tt.expectedResult, cfg.DatabaseDSN)
		})
	}
}
