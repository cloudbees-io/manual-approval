package manual_approval

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

var debug bool

func init() {
	debug = os.Getenv("DEBUG") == "true"
}

func (k *Config) Run(ctx context.Context) error {
	k.Context = ctx

	switch k.Handler {
	case "init":
		return k.init()
	case "callback":
		return k.callback()
	case "cancel":
		return k.cancel()
	default:
		return fmt.Errorf("unsupported handler type: %s", k.Handler)
	}
}

func (k *Config) defaultConfig() (string, string, error) {
	Debugf("Read default configuration from the environment variables")

	apiUrl := os.Getenv("URL")
	if apiUrl == "" {
		return "", "", fmt.Errorf("failed to get URL environment variable")
	}

	apiToken := os.Getenv("API_TOKEN")
	if apiToken == "" {
		return "", "nil", fmt.Errorf("failed to get API_TOKEN environment variable")
	}

	return apiUrl, apiToken, nil
}

func (k *Config) init() error {
	Debugf("Inside init handler\n")

	// approvers are optional
	approvers := os.Getenv("APPROVERS")

	// instructions are optional
	instructions := os.Getenv("INSTRUCTIONS")

	// by default disallowLaunchedByUser is false
	disallowLaunchedByUserStr := os.Getenv("DISALLOW_LAUNCHED_BY_USER")
	if disallowLaunchedByUserStr == "" {
		disallowLaunchedByUserStr = "false"
	}
	disallowLaunchedByUser, err := strconv.ParseBool(disallowLaunchedByUserStr)
	if err != nil {
		return err
	}

	// by default notifyAllEligibleUsers is false
	notifyAllEligibleUsersStr := os.Getenv("NOTIFY_ALL_ELIGIBLE_USERS")
	if notifyAllEligibleUsersStr == "" {
		notifyAllEligibleUsersStr = "false"
	}
	notifyAllEligibleUsers, err := strconv.ParseBool(notifyAllEligibleUsersStr)
	if err != nil {
		return err
	}

	// Construct request body
	body := map[string]interface{}{
		"disallowLaunchedByUser": disallowLaunchedByUser,
		"notifyAllEligibleUsers": notifyAllEligibleUsers,
	}

	if approvers != "" {
		body["approvers"] = strings.Split(approvers, ",")
	}

	if instructions != "" {
		body["instructions"] = instructions
	}

	resp, err := k.post("/v1/workflows/approval", body)
	if err != nil {
		ferr := writeStatus("FAILED", fmt.Sprintf("Failed to initialize workflow manual approval with error: '%s'", err))
		if ferr != nil {
			return ferr
		}
		return err
	}

	fmt.Printf("Response: %s\n", resp)
	return writeStatus("PENDING_APPROVAL", "Waiting for approval from approvers")

	return nil
}

func (k *Config) callback() error {
	Debugf("Inside callback handler\n")

	payload := os.Getenv("PAYLOAD")
	if payload == "" {
		return fmt.Errorf("failed to get PAYLOAD environment variable")
	}

	parsedPayload := map[string]interface{}{}
	err := json.Unmarshal([]byte(payload), &parsedPayload)
	if err != nil {
		return err
	}

	approvalStatus := parsedPayload["status"].(string)
	Debugf("Approval status: %s\n", approvalStatus)

	comments := parsedPayload["comments"].(string)
	Debugf("Comments: %s\n", comments)

	respondedOn := parsedPayload["respondedOn"].(string)
	Debugf("Responded on: %s\n", respondedOn)

	approverUserName := parsedPayload["userName"].(string)
	Debugf("Approver user name: %s\n", approverUserName)

	resp, err := k.post("/v1/workflows/approval/status", parsedPayload)
	if err != nil {
		return err
	}

	Debugf("Response: %s\n", resp)

	return nil
}

func (k *Config) cancel() error {
	Debugf("Inside cancel handler\n")

	cancellationReason := os.Getenv("CANCELLATION_REASON")
	if cancellationReason == "" {
		return fmt.Errorf("failed to get CANCELLATION_REASON environment variable")
	}

	// Construct request body
	body := map[string]interface{}{}
	if cancellationReason == "CANCELLED" {
		fmt.Println("Workflow aborted by user")
		fmt.Println("Cancelling the manual approval request")
		body["status"] = "UPDATE_MANUAL_APPROVAL_STATUS_ABORTED"
	} else {
		fmt.Println("Workflow timed out")
		fmt.Println("Workflow approval response was not received within allotted time.")
		body["status"] = "UPDATE_MANUAL_APPROVAL_STATUS_TIMED_OUT"
	}

	resp, err := k.post("/v1/workflows/approval/status", body)
	if err != nil {
		return err
	}

	fmt.Printf("Response: %s\n", resp)
	return nil
}

func writeStatus(status string, message string) error {
	statusFile := os.Getenv("CLOUDBEES_STATUS")
	if statusFile == "" {
		return fmt.Errorf("CLOUDBEES_STATUS environment variable missing")
	}
	output := map[string]interface{}{
		"status":  status,
		"message": message,
	}

	outputBytes, err := json.Marshal(&output)
	if err != nil {
		return err
	}
	err = os.WriteFile(statusFile, outputBytes, 0666)
	if err != nil {
		return fmt.Errorf("failed to write to %s: %w", statusFile, err)
	}
	return nil
}

type RealHttpClient struct{}

func (c *RealHttpClient) Do(req *http.Request) (*http.Response, error) {
	return http.DefaultClient.Do(req)
}

func (k *Config) post(apiPath string, requestBody map[string]interface{}) (string, error) {
	Debugf("Post http request to the platform API endpoint: %s\n", apiPath)

	// Read default configuration from the environment variables
	apiUrl, apiToken, err := k.defaultConfig()
	if err != nil {
		return "", err
	}

	// Construct the request URL for the API call
	requestURL, err := url.JoinPath(apiUrl, apiPath)
	if err != nil {
		return "", err
	}

	// Prepare JSON request body for REST API call
	body, err := json.Marshal(&requestBody)
	if err != nil {
		return "", err
	}
	Debugf("Payload: %s\n", string(body))

	// Use default client if it is not already provided in the configuration
	if k.Client == nil {
		k.Client = &RealHttpClient{}
	}

	apiReq, err := http.NewRequest(
		"POST",
		requestURL,
		bytes.NewReader(body),
	)
	if err != nil {
		return "", err
	}

	apiReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiToken))
	apiReq.Header.Set("Content-Type", "application/json")
	apiReq.Header.Set("Accept", "application/json")

	resp, err := k.Client.Do(apiReq)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return string(responseBody), fmt.Errorf("failed to send event: \nPOST %s\nHTTP/%d %s\n", requestURL, resp.StatusCode, resp.Status)
	}

	return string(responseBody), nil
}

func Debugf(format string, a ...any) {
	if debug {
		t := time.Now()
		fmt.Printf("%s - "+format, append([]any{t.Format(time.RFC3339)}, a...))
	}
}
