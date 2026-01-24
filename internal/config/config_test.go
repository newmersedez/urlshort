package config

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	tests := []struct{
		name			string
		serverAddress	string
		baseURL			string
		logLevel		string
		fileStoragePath	string
		err				error
	} {
		{
			name: 			"Valid",
			serverAddress: "localhost:9999",
			baseURL: 		"http://localhost:9999",
			logLevel: 		"info",
			fileStoragePath: "file.json",
			err: 			nil,
		},
		{
			name: 			"Empty server address",
			serverAddress:	"", 
			baseURL: 		"http://localhost:9999",
			logLevel: 		"info",
			fileStoragePath: "file.json",
			err: 			errServerAddressNotSet,
		},
		{
			name: 			"Empty base url",
			serverAddress: 	"localhost:9999", 
			baseURL: 		"",
			logLevel: 		"info",
			fileStoragePath: "file.json",
			err: 			errBaseURLNotSet,
		},
		{
			name: 			"Invalid server address format",
			serverAddress: 	":9999",
			baseURL: 		"http://localhost:9999",
			logLevel: 		"info",
			fileStoragePath: "file.json",
			err: 			errServerAddressInvalid,
		},
		{
			name: 			"Invalid base url format",
			serverAddress: 	"localhost:9999", 
			baseURL: 		"1http://234",
			logLevel: 		"info",
			fileStoragePath: "file.json",
			err: 			errBaseURLInvalid,
		},
		{
			name: 			"Missing base url schema",
			serverAddress: 	"localhost:9999", 
			baseURL: 		"localhost:9999",
			logLevel: 		"info",
			fileStoragePath: "file.json",
			err: 			errBaseURLMissingSchema,
		},
		{
			name: 			"Empty log level",
			serverAddress: 	"localhost:9999", 
			baseURL: 		"http://localhost:9999",
			logLevel: 		"",
			fileStoragePath: "file.json",
			err: 			errLogLevelNotSet,
		},
		{
			name: 			"Empty file storage path",
			serverAddress: 	"localhost:9999", 
			baseURL: 		"http://localhost:9999",
			logLevel: 		"info",
			fileStoragePath: "",
			err: 			errFileStoragePathNotSet,
		},
		{
			name: 			"Invalid file storage path format",
			serverAddress: 	"localhost:9999", 
			baseURL: 		"http://localhost:9999",
			logLevel: 		"info",
			fileStoragePath: "..",
			err: 			errFileStoragePathInvalid,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet("", flag.ContinueOnError)
			os.Args = []string{"test", "-a", test.serverAddress, "-b", test.baseURL, "-l", test.logLevel, "-f", test.fileStoragePath}
			flag.CommandLine.Parse(os.Args[1:])

			_, err := NewConfig()
			
			require.Equal(t, test.err, err)
		})
	}
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