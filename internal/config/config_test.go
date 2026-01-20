package config

import (
	"errors"
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	tests := []struct{
		name			string
		serverAddress	string
		baseURL			string
		ok				bool
		err				error
	} {
		{
			name: 			"Valid",
			serverAddress: "localhost:9999",
			baseURL: 		"http://localhost:9999",
			err: 			nil,
		},
		{
			name: 			"Empty server address",
			serverAddress:	"", 
			baseURL: 		"http://localhost:9999",
			err: 			errors.New(errServerAddressNotSet),
		},
		{
			name: 			"Empty base url",
			serverAddress: 	"localhost:9999", 
			baseURL: 		"",
			err: 			errors.New(errBaseURLNotSet),
		},
		{
			name: 			"Invalid server address format",
			serverAddress: 	":9999",
			baseURL: 		"http://localhost:9999",
			err: 			errors.New(errServerAddressInvalid),
		},
		{
			name: 			"Invalid base url format",
			serverAddress: 	"localhost:9999", 
			baseURL: 		"1http://234",
			err: 			errors.New(errBaseURLInvalid),
		},
		{
			name: 			"Missing base url schema",
			serverAddress: 	"localhost:9999", 
			baseURL: 		"localhost:9999",
			ok:				false,
			err: 			errors.New(errBaseURLMissingSchema),
		},
		{
			name: 			"Missing base url host",
			serverAddress: 	"localhost:9999", 
			baseURL: 		"http://",
			ok:				false,
			err: 			errors.New(errBaseURLMissingHost),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet("", flag.ContinueOnError)
			os.Args = []string{"test", "-a", test.serverAddress, "-b", test.baseURL}
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