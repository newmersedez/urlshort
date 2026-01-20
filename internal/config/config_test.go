package config

import (
	"flag"
	"os"
	"testing"
)

func TestServerAddress(t *testing.T) {
	tests := []struct {
		name		string
		envValue	string
		flagName	string
		flagValue	string
		want		string
	} {
		{
			name:		"Environment value has priority if set",
			envValue:	"localhost:9999",
			flagName:	"",
			flagValue: 	"",
			want: 		"localhost:9999",
		},
		{
			name:		"Flag value has priority if environment value is not set",
			envValue:	"",
			flagName:	"a",
			flagValue: 	"localhost:9999",
			want: 		"localhost:9999",
		},
		{
			name:		"Default value if flag and value are not set",
			envValue:	"",
			flagName:	"a",
			flagValue: 	"",
			want: 		"localhost:8080",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.envValue != "" {
				t.Setenv("SERVER_ADDRESS", test.envValue)
			}
			if test.flagName != "" {
				flag.Set("a", test.flagValue)
			}
		})
	}
}

func TestBaseURL(t *testing.T) {

}