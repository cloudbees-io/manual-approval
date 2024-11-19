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

var (
	instructionsInput  = "***instruction***\n`instruction2`\n# instruction3\n## instruction4\n### instruction5\n\n> Blockquotes can contain multiple paragraphs\n>\n> Add a > on the blank lines between the paragraps.\n\n- Rirst item\n- Second Item\n- Third item \n  - Indented item\n  - Indented item\n- Fourth item"
	instructionsOutput = "<p><em><strong>instruction</strong></em>\n<code>instruction2</code></p>\n<h1>instruction3</h1>\n<h2>instruction4</h2>\n<h3>instruction5</h3>\n<blockquote>\n<p>Blockquotes can contain multiple paragraphs</p>\n<p>Add a &gt; on the blank lines between the paragraps.</p>\n</blockquote>\n<ul>\n<li>Rirst item</li>\n<li>Second Item</li>\n<li>Third item\n<ul>\n<li>Indented item</li>\n<li>Indented item</li>\n</ul>\n</li>\n<li>Fourth item</li>\n</ul>\n"
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
			err:  "API_TOKEN environment variable missing",
		},
		{
			name: "no URL environment variable",
			env:  map[string]string{},
			err:  "URL environment variable missing",
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
				require.NotNil(t, req["approvers"])
				require.Equal(t, []interface{}{"123", "user@mail.com"}, req["approvers"])
				require.NotNil(t, req["instructions"])
				require.Equal(t, instructionsInput, req["instructions"].(string))
				require.Equal(t, false, req["disallowLaunchedByUser"].(bool))
				require.Equal(t, false, req["notifyAllEligibleUsers"].(bool))
			},
			respGenFunc: func() (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Status:     "200 OK",
					Body:       io.NopCloser(bytes.NewBufferString(`{"approvers":[{"userName": "testUserName", "userId": "123", "email": "user@mail.com"}]}`)),
				}, nil
			},
			env: map[string]string{
				"URL":              "http://test.com",
				"API_TOKEN":        "test",
				"CLOUDBEES_STATUS": "/tmp/test-status-out",
				"APPROVERS":        "123,user@mail.com",
				"INSTRUCTIONS":     instructionsInput,
			},
			err: "",
		},
		{
			name: "success with disallowLaunchedByUser",
			reqCheckFunc: func(req map[string]interface{}) {
				require.NotNil(t, req["approvers"])
				require.Equal(t, []interface{}{"123", "user@mail.com"}, req["approvers"])
				require.NotNil(t, req["instructions"])
				require.Equal(t, instructionsInput, req["instructions"].(string))
				require.Equal(t, true, req["disallowLaunchedByUser"].(bool))
				require.Equal(t, false, req["notifyAllEligibleUsers"].(bool))
			},
			respGenFunc: func() (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Status:     "200 OK",
					Body:       io.NopCloser(bytes.NewBufferString(`{"approvers":[{"userName": "testUserName", "userId": "123", "email": "user@mail.com"}]}`)),
				}, nil
			},
			env: map[string]string{
				"URL":                       "http://test.com",
				"API_TOKEN":                 "test",
				"CLOUDBEES_STATUS":          "/tmp/test-status-out",
				"APPROVERS":                 "123,user@mail.com",
				"INSTRUCTIONS":              instructionsInput,
				"DISALLOW_LAUNCHED_BY_USER": "true",
			},
			err: "",
		},
		{
			name: "success with invalid disallowLaunchedByUser",
			reqCheckFunc: func(req map[string]interface{}) {
				require.NotNil(t, req["approvers"])
				require.Equal(t, []interface{}{"123", "user@mail.com"}, req["approvers"])
				require.NotNil(t, req["instructions"])
				require.Equal(t, instructionsInput, req["instructions"].(string))
				require.Equal(t, true, req["disallowLaunchedByUser"].(bool))
				require.Equal(t, false, req["notifyAllEligibleUsers"].(bool))
			},
			respGenFunc: func() (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Status:     "200 OK",
					Body:       io.NopCloser(bytes.NewBufferString(`{"approvers":[{"userName": "testUserName", "userId": "123", "email": "user@mail.com"}]}`)),
				}, nil
			},
			env: map[string]string{
				"URL":                       "http://test.com",
				"API_TOKEN":                 "test",
				"CLOUDBEES_STATUS":          "/tmp/test-status-out",
				"APPROVERS":                 "123,user@mail.com",
				"INSTRUCTIONS":              instructionsInput,
				"DISALLOW_LAUNCHED_BY_USER": "invalid boolean",
			},
			err: "strconv.ParseBool: parsing \"invalid boolean\": invalid syntax",
		},
		{
			name: "success with notifyAllEligibleUsers",
			reqCheckFunc: func(req map[string]interface{}) {
				require.NotNil(t, req["approvers"])
				require.Equal(t, []interface{}{"123", "user@mail.com"}, req["approvers"])
				require.NotNil(t, req["instructions"])
				require.Equal(t, instructionsInput, req["instructions"].(string))
				require.Equal(t, false, req["disallowLaunchedByUser"].(bool))
				require.Equal(t, true, req["notifyAllEligibleUsers"].(bool))
			},
			respGenFunc: func() (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Status:     "200 OK",
					Body:       io.NopCloser(bytes.NewBufferString(`{"approvers":[{"userName": "testUserName", "userId": "123", "email": "user@mail.com"}]}`)),
				}, nil
			},
			env: map[string]string{
				"URL":                       "http://test.com",
				"API_TOKEN":                 "test",
				"CLOUDBEES_STATUS":          "/tmp/test-status-out",
				"APPROVERS":                 "123,user@mail.com",
				"INSTRUCTIONS":              instructionsInput,
				"NOTIFY_ALL_ELIGIBLE_USERS": "true",
			},
			err: "",
		},
		{
			name: "success with invalid notifyAllEligibleUsers",
			reqCheckFunc: func(req map[string]interface{}) {
				require.NotNil(t, req["approvers"])
				require.Equal(t, []interface{}{"123", "user@mail.com"}, req["approvers"])
				require.NotNil(t, req["instructions"])
				require.Equal(t, instructionsInput, req["instructions"].(string))
				require.Equal(t, false, req["disallowLaunchedByUser"].(bool))
				require.Equal(t, true, req["notifyAllEligibleUsers"].(bool))
			},
			respGenFunc: func() (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Status:     "200 OK",
					Body:       io.NopCloser(bytes.NewBufferString(`{"approvers":[{"userName": "testUserName", "userId": "123", "email": "user@mail.com"}]}`)),
				}, nil
			},
			env: map[string]string{
				"URL":                       "http://test.com",
				"API_TOKEN":                 "test",
				"CLOUDBEES_STATUS":          "/tmp/test-status-out",
				"APPROVERS":                 "123,user@mail.com",
				"INSTRUCTIONS":              instructionsInput,
				"NOTIFY_ALL_ELIGIBLE_USERS": "invalid boolean",
			},
			err: "strconv.ParseBool: parsing \"invalid boolean\": invalid syntax",
		},
		{
			name: "failure",
			reqCheckFunc: func(req map[string]interface{}) {
				require.NotNil(t, req["approvers"])
				require.Equal(t, []interface{}{"123", "user@mail.com"}, req["approvers"])
				require.NotNil(t, req["instructions"])
				require.Equal(t, instructionsInput, req["instructions"].(string))
				require.Equal(t, false, req["disallowLaunchedByUser"].(bool))
				require.Equal(t, false, req["notifyAllEligibleUsers"].(bool))
			},
			respGenFunc: func() (*http.Response, error) {
				return &http.Response{
					StatusCode: 500,
					Status:     "500 Internal Server Error",
					Body:       io.NopCloser(bytes.NewBufferString(`{"approvers":[{"userName": "testUserName", "userId": "123", "email": "user@mail.com"}]}`)),
				}, nil
			},
			env: map[string]string{
				"URL":              "http://test.com",
				"API_TOKEN":        "test",
				"CLOUDBEES_STATUS": "/tmp/test-status-out",
				"APPROVERS":        "123,user@mail.com",
				"INSTRUCTIONS":     instructionsInput,
			},
			err: "failed to send event: \nPOST http://test.com/v1/workflows/approval\nHTTP/500 500 Internal Server Error\n",
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

func Test_callback(t *testing.T) {
	tests := []struct {
		name         string
		reqCheckFunc func(req map[string]interface{})
		respGenFunc  func() (*http.Response, error)
		env          map[string]string
		client       *MockHttpClient
		statusInFile string
		err          string
	}{
		{
			name: "success APPROVED",
			reqCheckFunc: func(req map[string]interface{}) {
				require.Equal(t, "UPDATE_MANUAL_APPROVAL_STATUS_APPROVED", req["status"].(string))
				require.Equal(t, "test comments", req["comments"].(string))
				require.Equal(t, "123", req["userId"].(string))
				require.Equal(t, "testUserName", req["userName"].(string))
				require.Equal(t, "2009-11-10T23:00:00Z", req["respondedOn"].(string))
			},
			respGenFunc: func() (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Status:     "200 OK",
					Body:       io.NopCloser(bytes.NewBufferString(`{}`)),
				}, nil
			},
			env: map[string]string{
				"URL":              "http://test.com",
				"API_TOKEN":        "test",
				"CLOUDBEES_STATUS": "/tmp/test-status-out",
				"PAYLOAD":          "{\"status\":\"UPDATE_MANUAL_APPROVAL_STATUS_APPROVED\",\"comments\":\"test comments\",\"userId\":\"123\",\"userName\":\"testUserName\",\"respondedOn\":\"2009-11-10T23:00:00Z\"}",
			},
			statusInFile: "{\"message\":\"Successfully changed workflow manual approval status\",\"status\":\"APPROVED\"}",
			err:          "",
		},
		{
			name: "success REJECTED",
			reqCheckFunc: func(req map[string]interface{}) {
				require.Equal(t, "UPDATE_MANUAL_APPROVAL_STATUS_REJECTED", req["status"].(string))
				require.Equal(t, "test comments", req["comments"].(string))
				require.Equal(t, "123", req["userId"].(string))
				require.Equal(t, "testUserName", req["userName"].(string))
				require.Equal(t, "2009-11-10T23:00:00Z", req["respondedOn"].(string))
			},
			respGenFunc: func() (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Status:     "200 OK",
					Body:       io.NopCloser(bytes.NewBufferString(`{}`)),
				}, nil
			},
			env: map[string]string{
				"URL":              "http://test.com",
				"API_TOKEN":        "test",
				"CLOUDBEES_STATUS": "/tmp/test-status-out",
				"PAYLOAD":          "{\"status\":\"UPDATE_MANUAL_APPROVAL_STATUS_REJECTED\",\"comments\":\"test comments\",\"userId\":\"123\",\"userName\":\"testUserName\",\"respondedOn\":\"2009-11-10T23:00:00Z\"}",
			},
			statusInFile: "{\"message\":\"Successfully changed workflow manual approval status\",\"status\":\"REJECTED\"}",
			err:          "",
		},
		{
			name: "failure UNSPECIFIED",
			reqCheckFunc: func(req map[string]interface{}) {
				require.Equal(t, "UPDATE_MANUAL_APPROVAL_STATUS_UNSPECIFIED", req["status"].(string))
				require.Equal(t, "test comments", req["comments"].(string))
				require.Equal(t, "123", req["userId"].(string))
				require.Equal(t, "testUserName", req["userName"].(string))
				require.Equal(t, "2009-11-10T23:00:00Z", req["respondedOn"].(string))
			},
			respGenFunc: func() (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Status:     "200 OK",
					Body:       io.NopCloser(bytes.NewBufferString(`{}`)),
				}, nil
			},
			env: map[string]string{
				"URL":              "http://test.com",
				"API_TOKEN":        "test",
				"CLOUDBEES_STATUS": "/tmp/test-status-out",
				"PAYLOAD":          "{\"status\":\"UPDATE_MANUAL_APPROVAL_STATUS_UNSPECIFIED\",\"comments\":\"test comments\",\"userId\":\"123\",\"userName\":\"testUserName\",\"respondedOn\":\"2009-11-10T23:00:00Z\"}",
			},
			statusInFile: "{\"message\":\"Unexpected approval status 'UPDATE_MANUAL_APPROVAL_STATUS_UNSPECIFIED'\",\"status\":\"FAILED\"}",
			err:          "Unexpected approval status 'UPDATE_MANUAL_APPROVAL_STATUS_UNSPECIFIED'",
		},
		{
			name: "failure",
			reqCheckFunc: func(req map[string]interface{}) {
				require.Equal(t, "UPDATE_MANUAL_APPROVAL_STATUS_APPROVED", req["status"].(string))
				require.Equal(t, "test comments", req["comments"].(string))
				require.Equal(t, "123", req["userId"].(string))
				require.Equal(t, "testUserName", req["userName"].(string))
				require.Equal(t, "2009-11-10T23:00:00Z", req["respondedOn"].(string))
			},
			respGenFunc: func() (*http.Response, error) {
				return &http.Response{
					StatusCode: 500,
					Status:     "500 Internal Server Error",
					Body:       io.NopCloser(bytes.NewBufferString(`{}`)),
				}, nil
			},
			env: map[string]string{
				"URL":              "http://test.com",
				"API_TOKEN":        "test",
				"CLOUDBEES_STATUS": "/tmp/test-status-out",
				"PAYLOAD":          "{\"status\":\"UPDATE_MANUAL_APPROVAL_STATUS_APPROVED\",\"comments\":\"test comments\",\"userId\":\"123\",\"userName\":\"testUserName\",\"respondedOn\":\"2009-11-10T23:00:00Z\"}",
			},
			statusInFile: "{\"message\":\"Failed to change workflow manual approval status: 'failed to send event: \\nPOST http://test.com/v1/workflows/approval/status\\nHTTP/500 500 Internal Server Error\\n'\",\"status\":\"FAILED\"}",
			err:          "failed to send event: \nPOST http://test.com/v1/workflows/approval/status\nHTTP/500 500 Internal Server Error\n",
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
					require.Equal(t, "http://test.com/v1/workflows/approval/status", req.URL.String())
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
			err := c.callback()

			// Verify
			if tt.err == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Equal(t, tt.err, err.Error())
			}

			out, ferr := os.ReadFile(tt.env["CLOUDBEES_STATUS"])
			require.NoError(t, ferr)
			require.Equal(t, tt.statusInFile, string(out))
		})
	}
}

func Test_cancel(t *testing.T) {
	tests := []struct {
		name         string
		reqCheckFunc func(req map[string]interface{})
		respGenFunc  func() (*http.Response, error)
		env          map[string]string
		client       *MockHttpClient
		err          string
	}{
		{
			name: "success CANCELLED",
			reqCheckFunc: func(req map[string]interface{}) {
				require.NotNil(t, req["status"])
				require.Equal(t, "UPDATE_MANUAL_APPROVAL_STATUS_ABORTED", req["status"].(string))
			},
			respGenFunc: func() (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Status:     "200 OK",
					Body:       io.NopCloser(bytes.NewBufferString(`{}`)),
				}, nil
			},
			env: map[string]string{
				"URL":                 "http://test.com",
				"API_TOKEN":           "test",
				"CANCELLATION_REASON": "CANCELLED",
			},
			err: "",
		},
		{
			name: "success TIMED_OUT",
			reqCheckFunc: func(req map[string]interface{}) {
				require.NotNil(t, req["status"])
				require.Equal(t, "UPDATE_MANUAL_APPROVAL_STATUS_TIMED_OUT", req["status"].(string))
			},
			respGenFunc: func() (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Status:     "200 OK",
					Body:       io.NopCloser(bytes.NewBufferString(`{}`)),
				}, nil
			},
			env: map[string]string{
				"URL":                 "http://test.com",
				"API_TOKEN":           "test",
				"CANCELLATION_REASON": "TIMED_OUT",
			},
			err: "",
		},
		{
			name: "failure",
			reqCheckFunc: func(req map[string]interface{}) {
				require.NotNil(t, req["status"])
				require.Equal(t, "UPDATE_MANUAL_APPROVAL_STATUS_TIMED_OUT", req["status"].(string))
			},
			respGenFunc: func() (*http.Response, error) {
				return &http.Response{
					StatusCode: 500,
					Status:     "500 Internal Server Error",
					Body:       io.NopCloser(bytes.NewBufferString(`{}`)),
				}, nil
			},
			env: map[string]string{
				"URL":                 "http://test.com",
				"API_TOKEN":           "test",
				"CANCELLATION_REASON": "TIMED_OUT",
			},
			err: "failed to send event: \nPOST http://test.com/v1/workflows/approval/status\nHTTP/500 500 Internal Server Error\n",
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
					require.Equal(t, "http://test.com/v1/workflows/approval/status", req.URL.String())
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
			err := c.cancel()

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

func Test_markdown(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output string
	}{
		{
			name:   "Markdown",
			input:  instructionsInput,
			output: instructionsOutput,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run
			result := markdown(tt.input)

			// Verify
			require.Equal(t, tt.output, result)
		})
	}
}
