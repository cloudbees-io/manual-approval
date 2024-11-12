package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_UnknownArguments(t *testing.T) {
	prevArgs := os.Args
	defer func() {
		os.Args = prevArgs
	}()

	tests := []struct {
		name string
		args []string
		env  map[string]string
		err  string
	}{
		{
			name: "wrong handler argument",
			args: []string{"manual-approval", "--handler", "something wrong"},
			err:  "unsupported handler type: something wrong",
		},
		{
			name: "missed handler argument",
			args: []string{"manual-approval"},
			err:  "unsupported handler type: something wrong",
		},
		{
			name: "init - no URL environment variable",
			args: []string{"manual-approval", "--handler", "init"},
			env:  map[string]string{},
			err:  "failed to get URL environment variable",
		},
		{
			name: "init - wrong DISALLOW_LAUNCHED_BY_USER environment variable",
			args: []string{"manual-approval", "--handler", "init"},
			env:  map[string]string{"DISALLOW_LAUNCHED_BY_USER": "not a boolean"},
			err:  "strconv.ParseBool: parsing \"not a boolean\": invalid syntax",
		},
		{
			name: "init - wrong NOTIFY_ALL_ELIGIBLE_USERS environment variable",
			args: []string{"manual-approval", "--handler", "init"},
			env:  map[string]string{"NOTIFY_ALL_ELIGIBLE_USERS": "not a boolean"},
			err:  "strconv.ParseBool: parsing \"not a boolean\": invalid syntax",
		},
		{
			name: "init - no API_TOKEN environment variable",
			args: []string{"manual-approval", "--handler", "init"},
			env:  map[string]string{"URL": "http://test.com"},
			err:  "failed to get API_TOKEN environment variable",
		},
		/*{
			name: "callback - no PAYLOAD environment variable",
			args: []string{"manual-approval", "--handler", "callback"},
			env:  map[string]string{},
			err:  "failed to get PAYLOAD environment variable",
		},
		{
			name: "callback - no URL environment variable",
			args: []string{"manual-approval", "--handler", "callback"},
			env:  map[string]string{"PAYLOAD": "test payload"},
			err:  "failed to get URL environment variable",
		},
		{
			name: "callback - no API_TOKEN environment variable",
			args: []string{"manual-approval", "--handler", "callback"},
			env:  map[string]string{"PAYLOAD": "test payload", "URL": "http://test.com"},
			err:  "failed to get API_TOKEN environment variable",
		},
		{
			name: "cancel - no CANCELLATION_REASON environment variable",
			args: []string{"manual-approval", "--handler", "cancel"},
			env:  map[string]string{},
			err:  "failed to get CANCELLATION_REASON environment variable",
		},
		{
			name: "cancel - no URL environment variable",
			args: []string{"manual-approval", "--handler", "cancel"},
			env:  map[string]string{"CANCELLATION_REASON": "test reason"},
			err:  "failed to get URL environment variable",
		},
		{
			name: "cancel - no API_TOKEN environment variable",
			args: []string{"manual-approval", "--handler", "cancel"},
			env:  map[string]string{"CANCELLATION_REASON": "test reason", "URL": "http://test.com"},
			err:  "failed to get API_TOKEN environment variable",
		},*/
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare
			os.Args = tt.args
			for k, v := range tt.env {
				os.Setenv(k, v)
				defer func(k string) {
					os.Unsetenv(k)
				}(k)
			}

			// Run
			err := cmd.Execute()

			// Verify
			if tt.err == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Equal(t, tt.err, err.Error())
			}
		})
	}
}
