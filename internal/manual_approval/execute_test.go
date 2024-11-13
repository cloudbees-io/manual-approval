package manual_approval

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func init() {
	debug = true
}
func Test_defaultConfig(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		err  string
	}{
		{
			name: "success",
			env:  map[string]string{"URL": "http://test.com", "API_TOKEN": "test"},
			err:  "",
		},
		{
			name: "no API_TOKEN environment variable",
			env:  map[string]string{"URL": "http://test.com"},
			err:  "failed to get API_TOKEN environment variable",
		},
		{
			name: "no URL environment variable",
			env:  map[string]string{},
			err:  "failed to get URL environment variable",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare
			for k, v := range tt.env {
				os.Setenv(k, v)
				defer func(k string) {
					os.Unsetenv(k)
				}(k)
			}

			// Run
			c := Config{}
			apiUrl, apiToken, err := c.defaultConfig()

			// Verify
			if tt.err == "" {
				require.NoError(t, err)
				require.Equal(t, tt.env["URL"], apiUrl)
				require.Equal(t, tt.env["API_TOKEN"], apiToken)
			} else {
				require.Error(t, err)
				require.Equal(t, tt.err, err.Error())
			}
		})
	}
}

type MockHttpClient struct {
	MockDo func(req *http.Request) (*http.Response, error)
}

func (c *MockHttpClient) Do(req *http.Request) (*http.Response, error) {
	return c.MockDo(req)
}

func Test_init(t *testing.T) {
	tests := []struct {
		name         string
		reqCheckFunc func(req map[string]interface{})
		respGenFunc  func() (*http.Response, error)
		env          map[string]string
		client       *MockHttpClient
		err          string
	}{
		{
			name: "success",
			reqCheckFunc: func(req map[string]interface{}) {
				//require.Equal(t, []string{"user1@mail.com", "user2@mail.com"}, req["approvers"].([]string))
				//require.Equal(t, "some instruction", req["instruction"].(string))
				require.Nil(t, req["approvers"])
				require.Nil(t, req["instructions"])
				require.Equal(t, false, req["disallowLaunchedByUser"].(bool))
				require.Equal(t, false, req["notifyAllEligibleUsers"].(bool))
			},
			respGenFunc: func() (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBufferString(`{"approvers":[{"userId": "123", "userEmail": "user@mail.com"}]}`)),
				}, nil
			},
			env: map[string]string{"URL": "http://test.com", "API_TOKEN": "test", "CLOUDBEES_STATUS": "/tmp/test-status-out"},
			err: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare
			for k, v := range tt.env {
				os.Setenv(k, v)
				defer func(k string) {
					os.Unsetenv(k)
				}(k)
			}

			// Run
			c := Config{Client: &MockHttpClient{
				MockDo: func(req *http.Request) (*http.Response, error) {
					require.NotNil(t, req)
					require.Equal(t, "POST", req.Method)
					require.Equal(t, "http://test.com/v1/workflows/approval", req.URL.String())
					require.Equal(t, "application/json", req.Header.Get("Content-Type"))
					require.Equal(t, "application/json", req.Header.Get("Accept"))
					require.Contains(t, req.Header.Get("Authorization"), "Bearer ")

					reqBody := make(map[string]interface{})
					bodyReader, err := req.GetBody()
					require.NoError(t, err)
					body, err := io.ReadAll(bodyReader)
					require.NoError(t, err)
					err = json.Unmarshal(body, &reqBody)
					require.NoError(t, err)

					// Check parsed request body
					tt.reqCheckFunc(reqBody)

					// Generate response
					return tt.respGenFunc()
				},
			}}
			err := c.init()

			// Verify
			if tt.err == "" {
				require.NoError(t, err)
				out, ferr := os.ReadFile(tt.env["CLOUDBEES_STATUS"])
				require.NoError(t, ferr)
				require.Equal(t, "{\"message\":\"Waiting for approval from approvers\",\"status\":\"PENDING_APPROVAL\"}", string(out))

			} else {
				require.Error(t, err)
				require.Equal(t, tt.err, err.Error())
			}
		})
	}
}
