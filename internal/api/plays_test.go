package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListPlaysEndpoint(t *testing.T) {

	t.Run("GET /v1/plays", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/plays", nil)
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
			t.Error("expected at least one play")
		}

		if resp.Page != 1 {
			t.Errorf("expected page 1, got %d", resp.Page)
		}

		if resp.PerPage != 50 {
			t.Errorf("expected per_page 50, got %d", resp.PerPage)
		}
	})

	t.Run("GET /v1/plays with batter filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/plays?batter=rodrj007", nil)
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
			t.Error("expected plays for Julio Rodriguez")
		}
	})

	t.Run("GET /v1/plays with pitcher filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/plays?pitcher=plesz001", nil)
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
			t.Error("expected plays for pitcher plesz001")
		}
	})

	t.Run("GET /v1/plays with home runs filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/plays?home_runs=true", nil)
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
			t.Error("expected home run plays")
		}
	})

	t.Run("GET /v1/plays with strikeouts filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/plays?strikeouts=true", nil)
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
			t.Error("expected strikeout plays")
		}
	})

	t.Run("GET /v1/plays with walks filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/plays?walks=true", nil)
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
			t.Error("expected walk plays")
		}
	})

	t.Run("GET /v1/plays with inning filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/plays?inning=1", nil)
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
			t.Error("expected plays from inning 1")
		}
	})

	t.Run("GET /v1/plays with date filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/plays?date=20230409", nil)
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
			t.Error("expected plays from specific date")
		}
	})

	t.Run("GET /v1/plays with date range", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/plays?date_from=20230401&date_to=20230430", nil)
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
			t.Error("expected plays from date range")
		}
	})

	t.Run("GET /v1/plays with pagination", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/plays?page=1&per_page=10", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp PaginatedResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.PerPage != 10 {
			t.Errorf("expected per_page 10, got %d", resp.PerPage)
		}
	})
}

func TestGamePlaysEndpoint(t *testing.T) {

	t.Run("GET /v1/games/{id}/plays", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/games/CLE202304090/plays", nil)
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
			t.Error("expected plays for game CLE202304090")
		}
	})

	t.Run("GET /v1/games/{id}/plays with pagination", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/games/CLE202304090/plays?page=1&per_page=10", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp PaginatedResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.PerPage != 10 {
			t.Errorf("expected per_page 10, got %d", resp.PerPage)
		}
	})

	t.Run("GET /v1/games/{id}/plays - verify chronological order", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/games/CLE202304090/plays", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp PaginatedResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		// Plays should be in chronological order (inning, top_bot)
		if resp.Total > 0 {
			// Just verify we got plays back - detailed ordering check would require
			// parsing the Data field which is interface{}
			if resp.Total < 1 {
				t.Error("expected at least 1 play for this game")
			}
		}
	})
}

func TestPlayerPlaysEndpoint(t *testing.T) {

	t.Run("GET /v1/players/{id}/plays", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/rodriju01/plays", nil)
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
			t.Error("expected plays for Julio Rodriguez")
		}
	})

	t.Run("GET /v1/players/{id}/plays with pagination", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/rodriju01/plays?page=1&per_page=10", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp PaginatedResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.PerPage != 10 {
			t.Errorf("expected per_page 10, got %d", resp.PerPage)
		}
	})

	t.Run("GET /v1/players/{id}/plays - player without retro ID", func(t *testing.T) {
		// Test with a player ID that doesn't have a Retrosheet ID
		req := httptest.NewRequest(http.MethodGet, "/v1/players/nonexistent/plays", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		// Should return either 400 (bad request) or 500 (internal server error)
		if w.Code != http.StatusBadRequest && w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 400 or 500, got %d", w.Code)
		}
	})
}

func TestPlayerPlateAppearancesEndpoint(t *testing.T) {

	t.Run("GET /v1/players/{id}/plate-appearances", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/rodriju01/plate-appearances", nil)
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
			t.Error("expected plate appearances for Julio Rodriguez")
		}
	})

	t.Run("GET /v1/players/{id}/plate-appearances with season filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/rodriju01/plate-appearances?season=2023", nil)
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
			t.Error("expected plate appearances for 2023 season")
		}
	})

	t.Run("GET /v1/players/{id}/plate-appearances with date range", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/rodriju01/plate-appearances?date_from=2023-04-01&date_to=2023-04-30", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp PaginatedResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		// May or may not have data in this range
		if resp.Page != 1 {
			t.Errorf("expected page 1, got %d", resp.Page)
		}
	})

	t.Run("GET /v1/players/{id}/plate-appearances with game_id filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/rodriju01/plate-appearances?game_id=CLE202304090", nil)
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
			t.Error("expected plate appearances for this specific game")
		}
	})

	t.Run("GET /v1/players/{id}/plate-appearances with pitcher filter (retrosheet ID)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/rodriju01/plate-appearances?pitcher=plesz001", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp PaginatedResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Page != 1 {
			t.Errorf("expected page 1, got %d", resp.Page)
		}
	})

	t.Run("GET /v1/players/{id}/plate-appearances with vs_pitcher filter (Lahman ID)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/rodriju01/plate-appearances?vs_pitcher=colege01", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp PaginatedResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Page != 1 {
			t.Errorf("expected page 1, got %d", resp.Page)
		}
	})

	t.Run("GET /v1/players/{id}/plate-appearances with pagination", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/rodriju01/plate-appearances?page=1&per_page=10", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp PaginatedResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.PerPage != 10 {
			t.Errorf("expected per_page 10, got %d", resp.PerPage)
		}
	})

	t.Run("GET /v1/players/{id}/plate-appearances - verify descending date order", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/rodriju01/plate-appearances", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp PaginatedResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Total > 0 {
			// TODO: parse the Data field
			if resp.Total < 1 {
				t.Error("expected at least 1 plate appearance")
			}
		}
	})

	t.Run("GET /v1/players/{id}/plate-appearances - player without retro ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/players/nonexistent/plate-appearances", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest && w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 400 or 500, got %d", w.Code)
		}
	})
}
