package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/lib/pq"
	"stormlightlabs.org/baseball/internal/core"
)

func TestPlayerEndpoints(t *testing.T) {
	t.Run("GET /v1/players", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp PaginatedResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Total == 0 {
			t.Error("expected at least one player")
		}

		if resp.Page != 1 {
			t.Errorf("expected page 1, got %d", resp.Page)
		}
	})

	t.Run("GET /v1/players with name filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players?name=Judge", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp PaginatedResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Total == 0 {
			t.Error("expected to find Aaron Judge")
		}
	})

	t.Run("GET /v1/players with debut_year filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players?debut_year=2023", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp PaginatedResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Total == 0 {
			t.Error("expected to find players who debuted in 2023")
		}
	})

	t.Run("GET /v1/players with pagination", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players?page=1&per_page=5", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp PaginatedResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.PerPage != 5 {
			t.Errorf("expected per_page 5, got %d", resp.PerPage)
		}
	})

	t.Run("GET /v1/players/{id}", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/judgeaa01", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var player core.Player
		if err := json.NewDecoder(w.Body).Decode(&player); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if player.ID != "judgeaa01" {
			t.Errorf("expected player ID judgeaa01, got %s", player.ID)
		}

		if player.FirstName != "Aaron" {
			t.Errorf("expected first name Aaron, got %s", player.FirstName)
		}

		if player.LastName != "Judge" {
			t.Errorf("expected last name Judge, got %s", player.LastName)
		}
	})

	t.Run("GET /v1/players/{id} - not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/nonexistent", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}

func TestPlayerSeasonsEndpoint(t *testing.T) {
	t.Run("GET /v1/players/{id}/seasons", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/judgeaa01/seasons", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp PlayerSeasonsResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(resp.Batting) == 0 {
			t.Error("expected batting seasons")
		}
	})

	t.Run("GET /v1/players/{id}/seasons - player with both batting and pitching", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/donaljo02/seasons", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp PlayerSeasonsResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(resp.Batting) == 0 {
			t.Error("expected batting seasons for Josh Donaldson")
		}

		if len(resp.Pitching) == 0 {
			t.Error("expected pitching seasons for Josh Donaldson")
		}
	})

	t.Run("GET /v1/players/{id}/seasons - player with only batting", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/bettsmo01/seasons", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp PlayerSeasonsResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(resp.Batting) == 0 {
			t.Error("expected batting seasons for Mookie Betts")
		}

		if len(resp.Pitching) != 0 {
			t.Error("expected no pitching seasons for Mookie Betts")
		}
	})

	t.Run("GET /v1/players/{id}/seasons - player with only pitching", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/colege01/seasons", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp PlayerSeasonsResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(resp.Pitching) == 0 {
			t.Error("expected pitching seasons for Gerrit Cole")
		}

		if len(resp.Batting) != 0 {
			t.Error("expected no batting seasons for Gerrit Cole")
		}
	})
}

func TestPlayerStatsEndpoints(t *testing.T) {
	t.Run("GET /v1/players/{id}/stats/batting", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/judgeaa01/stats/batting", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp PlayerBattingStatsResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Career.PlayerID != "judgeaa01" {
			t.Errorf("expected player ID judgeaa01, got %s", resp.Career.PlayerID)
		}

		if len(resp.Seasons) == 0 {
			t.Error("expected at least one season")
		}
	})

	t.Run("GET /v1/players/{id}/stats/batting - verify career totals", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/judgeaa01/stats/batting", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp PlayerBattingStatsResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Career.HR != 37 {
			t.Errorf("expected 37 HR in 2023, got %d", resp.Career.HR)
		}

		if resp.Career.AB != 367 {
			t.Errorf("expected 367 AB in 2023, got %d", resp.Career.AB)
		}
	})

	t.Run("GET /v1/players/{id}/stats/batting - verify calculated rates", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/bettsmo01/stats/batting", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp PlayerBattingStatsResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Career.AVG == 0 {
			t.Error("expected AVG to be calculated")
		}

		if resp.Career.OBP == 0 {
			t.Error("expected OBP to be calculated")
		}

		if resp.Career.SLG == 0 {
			t.Error("expected SLG to be calculated")
		}

		if resp.Career.OPS == 0 {
			t.Error("expected OPS to be calculated")
		}
	})

	t.Run("GET /v1/players/{id}/stats/pitching", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/colege01/stats/pitching", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp PlayerPitchingStatsResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Career.PlayerID != "colege01" {
			t.Errorf("expected player ID colege01, got %s", resp.Career.PlayerID)
		}

		if len(resp.Seasons) == 0 {
			t.Error("expected at least one season")
		}
	})

	t.Run("GET /v1/players/{id}/stats/pitching - verify career totals", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/colege01/stats/pitching", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp PlayerPitchingStatsResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Career.W != 15 {
			t.Errorf("expected 15 wins in 2023, got %d", resp.Career.W)
		}

		if resp.Career.SO != 222 {
			t.Errorf("expected 222 strikeouts in 2023, got %d", resp.Career.SO)
		}
	})

	t.Run("GET /v1/players/{id}/stats/pitching - verify calculated rates", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/colege01/stats/pitching", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp PlayerPitchingStatsResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Career.ERA == 0 {
			t.Error("expected ERA to be calculated")
		}

		if resp.Career.WHIP == 0 {
			t.Error("expected WHIP to be calculated")
		}

		if resp.Career.KPer9 == 0 {
			t.Error("expected K/9 to be calculated")
		}
	})
}

func TestPlayerAwardsEndpoints(t *testing.T) {
	t.Run("GET /v1/players/{id}/awards", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/bettsmo01/awards", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp PaginatedResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Total == 0 {
			t.Error("expected at least one award for Mookie Betts")
		}
	})

	t.Run("GET /v1/players/{id}/awards with year filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/bettsmo01/awards?year=2018", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp PaginatedResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Total == 0 {
			t.Error("expected awards for Mookie Betts in 2018")
		}
	})

	t.Run("GET /v1/players/{id}/awards with pagination", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/bettsmo01/awards?page=1&per_page=5", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp PaginatedResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.PerPage != 5 {
			t.Errorf("expected per_page 5, got %d", resp.PerPage)
		}
	})

	t.Run("GET /v1/players/{id}/hall-of-fame", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/bettsmo01/hall-of-fame", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp HallOfFameResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
	})

	t.Run("GET /v1/players/{id}/hall-of-fame - player inducted", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/ruthba01/hall-of-fame", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp HallOfFameResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(resp.Records) == 0 {
			t.Error("expected Hall of Fame records for Babe Ruth")
		}
	})

	t.Run("GET /v1/players/{id}/hall-of-fame - player not inducted", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/judgeaa01/hall-of-fame", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp HallOfFameResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(resp.Records) != 0 {
			t.Error("expected no Hall of Fame records for Aaron Judge")
		}
	})
}

func TestPlayerGameLogsEndpoints(t *testing.T) {
	t.Run("GET /v1/players/{id}/game-logs", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/judgeaa01/game-logs", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("GET /v1/players/{id}/game-logs with season filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/judgeaa01/game-logs?season=2023", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("GET /v1/players/{id}/game-logs with pagination", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/judgeaa01/game-logs?page=1&per_page=10", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("GET /v1/players/{id}/game-logs/batting", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/judgeaa01/game-logs/batting", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("GET /v1/players/{id}/game-logs/batting with season filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/judgeaa01/game-logs/batting?season=2023", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("GET /v1/players/{id}/game-logs/batting with date range", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/judgeaa01/game-logs/batting?date_from=20230401&date_to=20230430", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("GET /v1/players/{id}/game-logs/batting with min_hr filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/judgeaa01/game-logs/batting?min_hr=2", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("GET /v1/players/{id}/game-logs/batting with multiple stat filters", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/judgeaa01/game-logs/batting?min_hr=1&min_h=2", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("GET /v1/players/{id}/game-logs/pitching", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/colege01/game-logs/pitching", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("GET /v1/players/{id}/game-logs/pitching with season filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/colege01/game-logs/pitching?season=2023", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("GET /v1/players/{id}/game-logs/pitching with min_so filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/colege01/game-logs/pitching?min_so=5", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("GET /v1/players/{id}/game-logs/pitching with min_ip filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/colege01/game-logs/pitching?min_ip=5", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("GET /v1/players/{id}/game-logs/fielding", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/judgeaa01/game-logs/fielding", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("GET /v1/players/{id}/game-logs/fielding with position filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/judgeaa01/game-logs/fielding?position=9", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("GET /v1/players/{id}/game-logs/fielding with date range", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/judgeaa01/game-logs/fielding?date_from=20230401&date_to=20230430", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})
}

func TestPlayerAppearancesEndpoint(t *testing.T) {
	t.Run("GET /v1/players/{id}/appearances", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/judgeaa01/appearances", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var appearances []core.PlayerAppearance
		if err := json.NewDecoder(w.Body).Decode(&appearances); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(appearances) == 0 {
			t.Error("expected at least one appearance record")
		}
	})

	t.Run("GET /v1/players/{id}/appearances - multi-position player", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/bettsmo01/appearances", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var appearances []core.PlayerAppearance
		if err := json.NewDecoder(w.Body).Decode(&appearances); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(appearances) == 0 {
			t.Error("expected appearances for Mookie Betts who played multiple positions")
		}
	})
}

func TestPlayerTeamsEndpoint(t *testing.T) {
	t.Run("GET /v1/players/{id}/teams", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/judgeaa01/teams", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var teams []core.PlayerTeamSeason
		if err := json.NewDecoder(w.Body).Decode(&teams); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(teams) == 0 {
			t.Error("expected at least one team season")
		}
	})

	t.Run("GET /v1/players/{id}/teams - player with multiple teams", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/donaljo02/teams", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var teams []core.PlayerTeamSeason
		if err := json.NewDecoder(w.Body).Decode(&teams); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(teams) < 2 {
			t.Error("expected Josh Donaldson to have played for multiple teams in 2023")
		}
	})

	t.Run("GET /v1/players/{id}/teams - player with single team career", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/rodriju01/teams", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var teams []core.PlayerTeamSeason
		if err := json.NewDecoder(w.Body).Decode(&teams); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(teams) == 0 {
			t.Error("expected team seasons for Julio Rodriguez")
		}
	})
}

func TestPlayerSalariesEndpoint(t *testing.T) {
	t.Run("GET /v1/players/{id}/salaries", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/judgeaa01/salaries", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		if body != "null" && body != "[]" && body != "" {
			var salaries []core.PlayerSalary
			if err := json.Unmarshal([]byte(body), &salaries); err != nil {
				t.Errorf("failed to decode salaries response: %v", err)
			}
		}
	})

	t.Run("GET /v1/players/{id}/salaries - verify chronological order", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/aardsda01/salaries", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		body := w.Body.Bytes()
		if len(body) == 0 || string(body) == "null" {
			return
		}

		var salaries []core.PlayerSalary
		if err := json.Unmarshal(body, &salaries); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(salaries) > 1 {
			for i := 1; i < len(salaries); i++ {
				if salaries[i].Year < salaries[i-1].Year {
					t.Error("expected salaries to be in chronological order")
				}
			}
		}
	})
}
