package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/lib/pq"
	"stormlightlabs.org/baseball/internal/core"
	"stormlightlabs.org/baseball/internal/seed"
)

func seedCrosswalkTestRows(t *testing.T) {
	t.Helper()

	if testDB == nil {
		t.Fatal("testDB is nil")
	}

	if _, err := testDB.Exec(`
		INSERT INTO team_mlbam_map (
			season,
			mlbam_team_id,
			mlb_abbreviation,
			mlb_team_code,
			mlb_file_code,
			mlb_team_name,
			mlb_franchise_name,
			mlb_club_name,
			local_team_id,
			local_franchise_id,
			local_team_name,
			local_league,
			match_method,
			confidence,
			source
		) VALUES (
			2026, 199901, 'NYY', 'nya', 'nya', 'New York Yankees', 'Yankees', 'Yankees',
			'NYA', 'NYY', 'New York Yankees', 'AL', 'team_id', 'high', 'test'
		)
		ON CONFLICT (season, mlbam_team_id) DO UPDATE SET
			local_team_id = EXCLUDED.local_team_id,
			local_franchise_id = EXCLUDED.local_franchise_id,
			match_method = EXCLUDED.match_method,
			confidence = EXCLUDED.confidence,
			source = EXCLUDED.source
	`); err != nil {
		t.Fatalf("failed to seed team_mlbam_map: %v", err)
	}

	if _, err := testDB.Exec(`
		INSERT INTO player_mlbam_map (
			mlbam_id,
			lahman_id,
			retro_id,
			bbref_id,
			full_name,
			source,
			confidence
		) VALUES (
			1999011, 'tstplr01', 'TSTP0001', 'tstplr01', 'Test Player', 'test', 'high'
		)
		ON CONFLICT (mlbam_id) DO UPDATE SET
			lahman_id = EXCLUDED.lahman_id,
			retro_id = EXCLUDED.retro_id,
			bbref_id = EXCLUDED.bbref_id,
			full_name = EXCLUDED.full_name,
			source = EXCLUDED.source,
			confidence = EXCLUDED.confidence
	`); err != nil {
		t.Fatalf("failed to seed player_mlbam_map: %v", err)
	}
}

func TestMetaEndpoints(t *testing.T) {
	seedCrosswalkTestRows(t)

	t.Run("GET /v1/meta", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/meta", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
		if got := w.Header().Get("X-Count-Mode"); got != "lightweight" {
			t.Fatalf("expected X-Count-Mode=lightweight, got %q", got)
		}

		var resp metaResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Version == "" {
			t.Error("expected version to be set")
		}

		if resp.GeneratedAt.IsZero() {
			t.Error("expected generated_at to be set")
		}

		if resp.Coverage == nil {
			t.Error("expected coverage to be set")
		}

		if resp.SchemaHashes == nil {
			t.Error("expected schema_hashes to be set")
		}

		if resp.EraLabels == nil {
			t.Error("expected era_labels to be set")
		}

		if got := resp.EraLabels["fed"]; got != "Federal League Era" {
			t.Errorf("expected fed era label to be %q, got %q", "Federal League Era", got)
		}

		if got := resp.EraLabels["nlg"]; got != "Negro Leagues Era" {
			t.Errorf("expected nlg era label to be %q, got %q", "Negro Leagues Era", got)
		}

		if len(resp.EraLabels) != len(seed.GetAllEras()) {
			t.Errorf("expected %d era labels, got %d", len(seed.GetAllEras()), len(resp.EraLabels))
		}

		if resp.Datasets == nil {
			t.Error("expected datasets to be set")
		}
	})

	t.Run("GET /v1/meta?strict=true", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/meta?strict=true", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
		got := w.Header().Get("X-Count-Mode")
		if got != "strict" && got != "fallback" {
			t.Fatalf("expected X-Count-Mode to be strict or fallback, got %q", got)
		}
	})

	t.Run("GET /v1/meta/datasets", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/meta/datasets", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
		if got := w.Header().Get("X-Count-Mode"); got != "lightweight" {
			t.Fatalf("expected X-Count-Mode=lightweight, got %q", got)
		}

		var datasets []core.DatasetStatus
		if err := json.NewDecoder(w.Body).Decode(&datasets); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if datasets == nil {
			t.Error("expected datasets array, got nil")
		}

		if len(datasets) < 3 {
			t.Errorf("expected supplemental dataset statuses, got %d", len(datasets))
		}
	})

	t.Run("GET /v1/meta/datasets?strict=true", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/meta/datasets?strict=true", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
		got := w.Header().Get("X-Count-Mode")
		if got != "strict" && got != "fallback" {
			t.Fatalf("expected X-Count-Mode to be strict or fallback, got %q", got)
		}
	})

	t.Run("GET /v1/meta/readiness", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/meta/readiness", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
		if got := w.Header().Get("X-Count-Mode"); got != "lightweight" {
			t.Fatalf("expected X-Count-Mode=lightweight, got %q", got)
		}

		var resp core.ReadinessStatus
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !resp.Ready {
			t.Error("expected readiness to be true with test fixtures")
		}

		if len(resp.Datasets) == 0 {
			t.Error("expected required datasets in readiness response")
		}
	})

	t.Run("GET /v1/meta/constants/woba", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/meta/constants/woba", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var constants []core.WOBAConstant
		if err := json.NewDecoder(w.Body).Decode(&constants); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(constants) == 0 {
			t.Error("expected at least one wOBA constant")
		}

		for _, c := range constants {
			if c.Season == 0 {
				t.Error("expected season to be set")
			}
			if c.WOBA == 0 {
				t.Error("expected woba to be set")
			}
		}
	})

	t.Run("GET /v1/meta/constants/woba with season filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/meta/constants/woba?season=2023", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var constants []core.WOBAConstant
		if err := json.NewDecoder(w.Body).Decode(&constants); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(constants) != 1 {
			t.Errorf("expected 1 constant, got %d", len(constants))
		}

		if len(constants) > 0 && constants[0].Season != 2023 {
			t.Errorf("expected season 2023, got %d", constants[0].Season)
		}
	})

	t.Run("GET /v1/meta/constants/league", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/meta/constants/league", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var constants []core.LeagueConstant
		if err := json.NewDecoder(w.Body).Decode(&constants); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(constants) == 0 {
			t.Error("expected at least one league constant")
		}

		for _, c := range constants {
			if c.Season == 0 {
				t.Error("expected season to be set")
			}
			if c.League == "" {
				t.Error("expected league to be set")
			}
		}
	})

	t.Run("GET /v1/meta/constants/league with filters", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/meta/constants/league?season=2024&league=AL", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var constants []core.LeagueConstant
		if err := json.NewDecoder(w.Body).Decode(&constants); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(constants) != 1 {
			t.Errorf("expected 1 constant, got %d", len(constants))
		}

		if len(constants) > 0 {
			if constants[0].Season != 2024 {
				t.Errorf("expected season 2024, got %d", constants[0].Season)
			}
			if constants[0].League != "AL" {
				t.Errorf("expected league AL, got %s", constants[0].League)
			}
		}
	})

	t.Run("GET /v1/meta/constants/park-factors", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/meta/constants/park-factors", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var factors []core.ParkFactorRow
		if err := json.NewDecoder(w.Body).Decode(&factors); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(factors) == 0 {
			t.Error("expected at least one park factor")
		}

		for _, f := range factors {
			if f.Season == 0 {
				t.Error("expected season to be set")
			}
			if f.ParkID == "" {
				t.Error("expected park_id to be set")
			}
		}
	})

	t.Run("GET /v1/meta/constants/park-factors with filters", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/meta/constants/park-factors?season=2023&team=BOS", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var factors []core.ParkFactorRow
		if err := json.NewDecoder(w.Body).Decode(&factors); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(factors) == 0 {
			t.Error("expected at least one park factor for BOS in 2023")
		}

		for _, f := range factors {
			if f.Season != 2023 {
				t.Errorf("expected season 2023, got %d", f.Season)
			}
			if f.TeamID != nil && *f.TeamID != "BOS" {
				t.Errorf("expected team BOS, got %s", *f.TeamID)
			}
		}
	})

	t.Run("GET /v1/meta/crosswalk/teams", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/meta/crosswalk/teams?season=2026&mlbam_team_id=199901", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp core.TeamCrosswalkResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode team crosswalk response: %v", err)
		}
		if resp.Total == 0 || len(resp.Rows) == 0 {
			t.Fatalf("expected at least one crosswalk row, got total=%d", resp.Total)
		}
		if resp.Rows[0].MLBAMTeamID == nil || *resp.Rows[0].MLBAMTeamID != 199901 {
			t.Fatalf("expected mlbam_team_id 199901, got %+v", resp.Rows[0].MLBAMTeamID)
		}
	})

	t.Run("GET /v1/meta/crosswalk/teams/by-mlbam/{mlbam_team_id}", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/meta/crosswalk/teams/by-mlbam/199901?season=2026", nil)
		req.SetPathValue("mlbam_team_id", "199901")
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp core.TeamCrosswalkResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode team crosswalk by-mlbam response: %v", err)
		}
		if resp.Total == 0 {
			t.Fatal("expected at least one team crosswalk row for mlbam 199901")
		}
	})

	t.Run("GET /v1/meta/crosswalk/players", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/meta/crosswalk/players?mlbam_id=1999011", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp core.PlayerCrosswalkResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode player crosswalk response: %v", err)
		}
		if resp.Total == 0 || len(resp.Rows) == 0 {
			t.Fatalf("expected at least one player crosswalk row, got total=%d", resp.Total)
		}
		if resp.Rows[0].MLBAMID == nil || *resp.Rows[0].MLBAMID != 1999011 {
			t.Fatalf("expected mlbam_id 1999011, got %+v", resp.Rows[0].MLBAMID)
		}
	})

	t.Run("GET /v1/meta/crosswalk/players/by-mlbam/{mlbam_id}", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/meta/crosswalk/players/by-mlbam/1999011", nil)
		req.SetPathValue("mlbam_id", "1999011")
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp core.PlayerCrosswalkResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode player crosswalk by-mlbam response: %v", err)
		}
		if resp.Total == 0 {
			t.Fatal("expected at least one player crosswalk row for mlbam 1999011")
		}
	})

	t.Run("GET /v1/mlb/crosswalk/teams returns not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/mlb/crosswalk/teams?season=2026", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status 404 for removed route, got %d", w.Code)
		}
	})
}

func TestHealthEndpoint(t *testing.T) {
	t.Run("GET /v1/health", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/health", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp HealthResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Status != "ok" {
			t.Errorf("expected status 'ok', got '%s'", resp.Status)
		}
	})

	t.Run("GET /v1/ready", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/ready", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
		if got := w.Header().Get("X-Count-Mode"); got != "lightweight" {
			t.Fatalf("expected X-Count-Mode=lightweight, got %q", got)
		}

		var resp core.ReadinessStatus
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !resp.Ready {
			t.Error("expected ready=true")
		}
	})
}
