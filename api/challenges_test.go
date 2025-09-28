package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestGetChallenges_Success(t *testing.T) {
	responseBody := `{
		"success": true,
		"data": [
			{
				"id": 1,
				"type": "standard",
				"name": "warmup",
				"value": 100,
				"solves": 10,
				"solved_by_me": false,
				"category": "misc"
			}
		]
	}`

	mock := mockResponse(t, newResponse(200, responseBody))

	base, _ := url.Parse("https://ctf.example.com")
	api := &ApiClient{
		client:  mock,
		baseUrl: base,
	}

	challs, err := api.GetChallenges(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(challs) != 1 {
		t.Fatalf("expected 1 challenge, got %d", len(challs))
	}
	if challs[0].Name != "warmup" {
		t.Errorf("expected challenge name 'warmup', got %q", challs[0].Name)
	}
}

func TestGetChallenges_Failure(t *testing.T) {
	responseBody := `{"success": false, "data": []}`

	mock := mockResponse(t, newResponse(200, responseBody))

	base, _ := url.Parse("https://ctf.example.com")
	api := &ApiClient{
		client:  mock,
		baseUrl: base,
	}

	_, err := api.GetChallenges(context.Background())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, ErrFailedFetchingChals) {
		t.Errorf("expected error %q, got %q", ErrFailedFetchingChals, err.Error())
	}
}

func TestGetChallenge_Success(t *testing.T) {
	responseBody := `{
		"success": true,
		"data": {
			"id": 42,
			"type": "standard",
			"name": "reverseme",
			"description": "A reversing challenge",
			"value": 200,
			"solves": 5,
			"solved_by_me": false,
			"category": "reversing",
			"files": [],
			"connection_info": "",
			"tags": [],
			"attempts": 0,
			"max_attempts": 0,
			"hints": []
		}
	}`

	mock := mockResponse(t, newResponse(200, responseBody))

	base, _ := url.Parse("https://ctf.example.com")
	api := &ApiClient{
		client:  mock,
		baseUrl: base,
	}

	chall, err := api.GetChallenge(context.Background(), 42)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if chall.Id != 42 {
		t.Errorf("expected challenge id 42, got %d", chall.Id)
	}
	if chall.Name != "reverseme" {
		t.Errorf("expected challenge name 'reverseme', got %q", chall.Name)
	}
	if chall.Category != "reversing" {
		t.Errorf("expected category 'reversing', got %q", chall.Category)
	}
}

func TestGetChallenge_Failure(t *testing.T) {
	responseBody := `{"success": false, "data": {}}`

	mock := mockResponse(t, newResponse(200, responseBody))

	base, _ := url.Parse("https://ctf.example.com")
	api := &ApiClient{
		client:  mock,
		baseUrl: base,
	}

	_, err := api.GetChallenge(context.Background(), 99)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), errFailedFetchingChallenge) {
		t.Errorf("expected error to contain %q, got %q", errFailedFetchingChallenge, err.Error())
	}
}

func TestSubmitFlag_Success(t *testing.T) {
	htmlBody := `
	<html>
		<head></head>
		<body>
		<script>
			var data = { 'csrfNonce': "abcdef123456", };
		</script>
		</body>
	</html>`

	attemptResp := ApiResponse[AttemptResult]{
		Success: true,
		Data:    AttemptResult{Status: "correct", Message: "Well done!"},
	}
	jsonBody, _ := json.Marshal(attemptResp)

	mock := sequenceResponses(t, []*http.Response{
		newResponse(200, htmlBody),
		newResponse(200, string(jsonBody)),
	})

	base, _ := url.Parse("https://ctf.example.com")
	api := &ApiClient{
		client:  mock,
		baseUrl: base,
	}

	result, err := api.SubmitFlag(context.Background(), 123, "flag{test}")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Status != "correct" {
		t.Errorf("expected status 'correct', got %q", result.Status)
	}
	if result.Message != "Well done!" {
		t.Errorf("expected message 'Well done!', got %q", result.Message)
	}
}

func TestSubmitFlag_Failure(t *testing.T) {
	htmlBody := `<script>var data = { 'csrfNonce': "deadbeef", };</script>`

	attemptResp := ApiResponse[AttemptResult]{
		Success: false,
		Data:    AttemptResult{Status: "incorrect", Message: "Wrong flag!"},
	}
	jsonBody, _ := json.Marshal(attemptResp)

	mock := sequenceResponses(t, []*http.Response{
		newResponse(200, htmlBody),
		newResponse(200, string(jsonBody)),
	})

	base, _ := url.Parse("https://ctf.example.com")
	api := &ApiClient{
		client:  mock,
		baseUrl: base,
	}

	_, err := api.SubmitFlag(context.Background(), 321, "flag{bad}")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), errFailedSubmittingFlag) {
		t.Errorf("expected error to contain %q, got %q", errFailedSubmittingFlag, err.Error())
	}
}
