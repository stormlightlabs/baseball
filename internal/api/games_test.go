package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/lib/pq"
)

func TestGameEndpoints(t *testing.T) {
	t.Run("GET /v1/games", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/games", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp struct {
			Data []struct {
				ID string `json:"id"`
			} `json:"data"`
			Page    int `json:"page"`
			PerPage int `json:"per_page"`
			Total   int `json:"total"`
		}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Total == 0 {
			t.Error("expected at least one game")
		}

		if resp.Page != 1 {
			t.Errorf("expected page 1, got %d", resp.Page)
		}
		if len(resp.Data) == 0 {
			t.Fatal("expected at least one game in response data")
		}
		if resp.Data[0].ID == "" {
			t.Fatal("expected first listed game to include an id")
		}
	})

	t.Run("GET /v1/games list IDs resolve in detail endpoints", func(t *testing.T) {
		listReq := httptest.NewRequest(http.MethodGet, "/v1/games?date_from=2023-09-10&date_to=2023-09-10&home_team=NYA&per_page=1", nil)
		listW := httptest.NewRecorder()
		testServer.ServeHTTP(listW, listReq)
		if listW.Code != http.StatusOK {
			t.Fatalf("expected list status 200, got %d: %s", listW.Code, listW.Body.String())
		}

		var listResp struct {
			Data []struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		if err := json.NewDecoder(listW.Body).Decode(&listResp); err != nil {
			t.Fatalf("failed to decode filtered list response: %v", err)
		}
		if len(listResp.Data) == 0 {
			t.Fatal("expected filtered list to return at least one game")
		}
		listedGameID := listResp.Data[0].ID
		if listedGameID == "" {
			t.Fatal("expected filtered list game to include an id")
		}

		detailReq := httptest.NewRequest(http.MethodGet, "/v1/games/"+listedGameID, nil)
		detailReq.SetPathValue("id", listedGameID)
		detailW := httptest.NewRecorder()
		testServer.ServeHTTP(detailW, detailReq)
		if detailW.Code != http.StatusOK {
			t.Fatalf("expected detail status 200 for listed game id %s, got %d: %s", listedGameID, detailW.Code, detailW.Body.String())
		}

		boxscoreReq := httptest.NewRequest(http.MethodGet, "/v1/games/"+listedGameID+"/boxscore", nil)
		boxscoreReq.SetPathValue("id", listedGameID)
		boxscoreW := httptest.NewRecorder()
		testServer.ServeHTTP(boxscoreW, boxscoreReq)
		if boxscoreW.Code != http.StatusOK {
			t.Fatalf("expected boxscore status 200 for listed game id %s, got %d: %s", listedGameID, boxscoreW.Code, boxscoreW.Body.String())
		}

		summaryReq := httptest.NewRequest(http.MethodGet, "/v1/games/"+listedGameID+"/summary", nil)
		summaryReq.SetPathValue("id", listedGameID)
		summaryW := httptest.NewRecorder()
		testServer.ServeHTTP(summaryW, summaryReq)
		if summaryW.Code != http.StatusOK {
			t.Fatalf("expected summary status 200 for listed game id %s, got %d: %s", listedGameID, summaryW.Code, summaryW.Body.String())
		}
	})

	t.Run("GET /v1/games with date filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/games?date_from=2023-04-01&date_to=2023-04-01", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp PaginatedResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
	})

	t.Run("GET /v1/games with team filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/games?home_team=NYA", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp PaginatedResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
	})

	t.Run("GET /v1/games/{id}", func(t *testing.T) {
		gameID := "NYA202309100"
		req := httptest.NewRequest(http.MethodGet, "/v1/games/"+gameID, nil)
		req.SetPathValue("id", gameID)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("GET /v1/games/{id} - not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/games/INVALID999", nil)
		req.SetPathValue("id", "INVALID999")
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})

	t.Run("GET /v1/games/{id}/boxscore", func(t *testing.T) {
		gameID := "NYA202309100"
		req := httptest.NewRequest(http.MethodGet, "/v1/games/"+gameID+"/boxscore", nil)
		req.SetPathValue("id", gameID)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("GET /v1/seasons/{year}/schedule", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/seasons/2023/schedule", nil)
		req.SetPathValue("year", "2023")
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp PaginatedResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
	})

	t.Run("GET /v1/seasons/{year}/dates/{date}/games", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/seasons/2023/dates/2023-09-10/games", nil)
		req.SetPathValue("year", "2023")
		req.SetPathValue("date", "2023-09-10")
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("GET /v1/seasons/{year}/teams/{team_id}/games", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/seasons/2023/teams/NYA/games", nil)
		req.SetPathValue("team_id", "NYA")
		req.SetPathValue("year", "2023")
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})
}
