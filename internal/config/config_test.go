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
	os.Args = []string{"test", "-a", "localhost:9999", "-b", "http://localhost:9999", "-l", "info", "-f", "file.json"}
	flag.CommandLine.Parse(os.Args[1:])

	_, err := NewConfig()
	
	require.NoError(t, err)
}

func TestServerAddressPriority(t *testing.T) {
	tests := []struct {
		name		string
		envName		string
		envValue 	string
		flagName	string
		flagValue	string
		want     	string
	}{
		{
			name:		"Environment value has priority if set",
			envName:	"SERVER_ADDRESS",
			envValue:	"localhost:9999",
			flagName:	"-a",
			flagValue:	"localhost:8888",
			want:    	"localhost:9999",
		},
		{
			name:		"Flag value has priority if environment value is not set",
			envName:	"",
			envValue:	"",
			flagName:	"-a",
			flagValue:	"localhost:8888",
			want:    	"localhost:8888",
		},
		{
			name:		"Default value if flag and value are not set",
			envName:	"",
			envValue:	"",
			flagName:	"",
			flagValue:	"",
			want:    	"localhost:8080",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet("", flag.ContinueOnError)
			
			if test.envValue != "" {
				t.Setenv(test.envName, test.envValue)
			} else {
				os.Unsetenv(test.envName)
			}

			if test.flagName != "" {
				os.Args = []string{"test", test.flagName, test.flagValue}
				flag.CommandLine.Parse(os.Args[1:])
			} else {
				os.Args = []string{"test"}
				flag.CommandLine.Parse(os.Args[1:])
			}

			cfg, err := NewConfig()
			require.NoError(t, err)
			require.Equal(t, test.want, cfg.ServerAddr)
		})
	}
}

func TestBaseURLPriority(t *testing.T) {
	tests := []struct {
		name		string
		envName		string
		envValue 	string
		flagName	string
		flagValue	string
		want     	string
	}{
		{
			name:		"Environment value has priority if set",
			envName:	"BASE_URL",
			envValue:	"http://localhost:9999",
			flagName:	"-b",
			flagValue:	"http://localhost:8888",
			want:    	"http://localhost:9999",
		},
		{
			name:		"Flag value has priority if environment value is not set",
			envName:	"",
			envValue:	"",
			flagName:	"-b",
			flagValue:	"http://localhost:8888",
			want:    	"http://localhost:8888",
		},
		{
			name:		"Default value if flag and value are not set",
			envName:	"",
			envValue:	"",
			flagName:	"",
			flagValue:	"",
			want:    	"http://localhost:8080",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet("", flag.ContinueOnError)

			if test.envValue != "" {
				t.Setenv(test.envName, test.envValue)
			} else {
				os.Unsetenv(test.envName)
			}

			if test.flagName != "" {
				os.Args = []string{"test", test.flagName, test.flagValue}
				flag.CommandLine.Parse(os.Args[1:])
			} else {
				os.Args = []string{"test"}
				flag.CommandLine.Parse(os.Args[1:])
			}

			cfg, err := NewConfig()
			require.NoError(t, err)
			require.Equal(t, test.want, cfg.BaseURL)
		})
	}
}

func TestLogLevelPriority(t *testing.T) {
	tests := []struct {
		name		string
		envName		string
		envValue 	string
		flagName	string
		flagValue	string
		want     	string
	}{
		{
			name:		"Environment value has priority if set",
			envName:	"LOG_LEVEL",
			envValue:	"warn",
			flagName:	"-l",
			flagValue:	"debug",
			want:    	"warn",
		},
		{
			name:		"Flag value has priority if environment value is not set",
			envName:	"",
			envValue:	"",
			flagName:	"-l",
			flagValue:	"warn",
			want:    	"warn",
		},
		{
			name:		"Default value if flag and value are not set",
			envName:	"",
			envValue:	"",
			flagName:	"",
			flagValue:	"",
			want:    	"info",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet("", flag.ContinueOnError)

			if test.envValue != "" {
				t.Setenv(test.envName, test.envValue)
			} else {
				os.Unsetenv(test.envName)
			}

			if test.flagName != "" {
				os.Args = []string{"test", test.flagName, test.flagValue}
				flag.CommandLine.Parse(os.Args[1:])
			} else {
				os.Args = []string{"test"}
				flag.CommandLine.Parse(os.Args[1:])
			}

			cfg, err := NewConfig()
			require.NoError(t, err)
			require.Equal(t, test.want, cfg.LogLevel)
		})
	}
}

func TestFileStoragePathPriority(t *testing.T) {
	tests := []struct {
		name		string
		envName		string
		envValue 	string
		flagName	string
		flagValue	string
		want     	string
	}{
		{
			name:		"Environment value has priority if set",
			envName:	"FILE_STORAGE_PATH",
			envValue:	"priority.json",
			flagName:	"-f",
			flagValue:	"secondary.json",
			want:    	"priority.json",
		},
		{
			name:		"Flag value has priority if environment value is not set",
			envName:	"",
			envValue:	"",
			flagName:	"-f",
			flagValue:	"priority.json",
			want:    	"priority.json",
		},
		{
			name:		"Default value if flag and value are not set",
			envName:	"",
			envValue:	"",
			flagName:	"",
			flagValue:	"",
			want:    	filepath.Join(os.TempDir(), "storage.json"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet("", flag.ContinueOnError)

			if test.envValue != "" {
				t.Setenv(test.envName, test.envValue)
			} else {
				os.Unsetenv(test.envName)
			}

			if test.flagName != "" {
				os.Args = []string{"test", test.flagName, test.flagValue}
				flag.CommandLine.Parse(os.Args[1:])
			} else {
				os.Args = []string{"test"}
				flag.CommandLine.Parse(os.Args[1:])
			}

			cfg, err := NewConfig()
			require.NoError(t, err)
			require.Equal(t, test.want, cfg.FileStoragePath)
		})
	}
}