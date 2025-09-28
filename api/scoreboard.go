package api

import (
	"context"
	"encoding/json"
	"fmt"
)

func (c *ApiClient) GetScoreboard(ctx context.Context) ([]ScoreboardEntry, error) {
	u := fmt.Sprintf("%s%s", c.baseUrl, scoreboardApiURL)

	resp, err := c.get(ctx, u)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	var scoreboard ApiResponse[[]ScoreboardEntry]
	if err := json.NewDecoder(resp.Body).Decode(&scoreboard); err != nil {
		return nil, err
	}

	if scoreboard.Success != true {
		return nil, ErrFailedFetchingBoard
	}

	return scoreboard.Data, nil
}
