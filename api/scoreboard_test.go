package api

import (
	"net/url"
	"strings"
	"testing"
)

func TestGetScoreboard_Success(t *testing.T) {
	responseBody := `{
		"success": true,
		"data": [
			{
				"pos": 1,
				"account_id": 101,
				"account_url": "/teams/101",
				"account_type": "team",
				"oauth_id": 0,
				"name": "LegendaryTeam",
				"score": 5000,
				"bracket_id": 1,
				"bracket_name": "Overall",
				"members": [
					{
						"id": 201,
						"oauth_id": 0,
						"name": "player1",
						"score": 5000,
						"bracket_id": 1,
						"bracket_name": "Overall"
					}
				]
			}
		]
	}`

	mock := mockResponse(t, newResponse(200, responseBody))

	base, _ := url.Parse("https://ctf.example.com")
	api := &ApiClient{client: mock, baseUrl: base}

	scoreboard, err := api.GetScoreboard()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(scoreboard) != 1 {
		t.Fatalf("expected 1 scoreboard entry, got %d", len(scoreboard))
	}

	entry := scoreboard[0]
	if entry.Name != "LegendaryTeam" {
		t.Errorf("expected team name 'LegendaryTeam', got %q", entry.Name)
	}
	if entry.Score != 5000 {
		t.Errorf("expected score 5000, got %d", entry.Score)
	}
	if entry.Position != 1 {
		t.Errorf("expected position 1, got %d", entry.Position)
	}
}

func TestGetScoreboard_Failure(t *testing.T) {
	responseBody := `{"success": false, "data": []}`

	mock := mockResponse(t, newResponse(200, responseBody))

	base, _ := url.Parse("https://ctf.example.com")
	api := &ApiClient{client: mock, baseUrl: base}

	_, err := api.GetScoreboard()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), errFailedFetchingScoreboard) {
		t.Errorf("expected error to contain %q, got %q", errFailedFetchingScoreboard, err.Error())
	}
}
