package search

import (
	"regexp"
	"strconv"
	"strings"

	"stormlightlabs.org/baseball/internal/core"
)

// GameQuery represents parsed natural language query components for game search.
type GameQuery struct {
	RawQuery   string
	Season     *int
	HomeTeamID *string
	AwayTeamID *string
	GameType   *string
	GameNumber *int
}

var (
	// Regex to match 4-digit years
	yearPattern = regexp.MustCompile(`\b(19\d{2}|20\d{2})\b`)

	// Regex to match "game N" pattern
	gameNumberPattern = regexp.MustCompile(`\bgame\s+(\d+)\b`)

	// Common postseason/series keywords mapped to database game_type values
	seriesKeywords = map[string]string{
		"world series":    "worldseries",
		"ws":              "worldseries",
		"alcs":            "lcs",
		"nlcs":            "lcs",
		"alds":            "divisionseries",
		"nlds":            "divisionseries",
		"division series": "divisionseries",
		"wildcard":        "wildcard",
		"wild card":       "wildcard",
		"playoffs":        "postseason",
		"postseason":      "postseason",
		"all-star":        "allstar",
		"allstar":         "allstar",
		"all star":        "allstar",
	}
)

// ParseGameQuery extracts structured filters from a natural language query.
// It identifies years, team names, series keywords, and game numbers.
func ParseGameQuery(raw string) GameQuery {
	query := GameQuery{
		RawQuery: raw,
	}

	normalized := strings.ToLower(strings.TrimSpace(raw))

	if matches := yearPattern.FindStringSubmatch(normalized); len(matches) > 1 {
		if year, err := strconv.Atoi(matches[1]); err == nil {
			query.Season = &year
		}
	}

	if matches := gameNumberPattern.FindStringSubmatch(normalized); len(matches) > 1 {
		if gameNum, err := strconv.Atoi(matches[1]); err == nil {
			query.GameNumber = &gameNum
		}
	}

	for keyword, gameType := range seriesKeywords {
		if strings.Contains(normalized, keyword) {
			query.GameType = &gameType
			break
		}
	}

	return query
}

// TeamAliasResolver resolves team name aliases to official team IDs.
type TeamAliasResolver interface {
	ResolveTeamAlias(alias string, season *int) (core.TeamID, bool)
}

// EnrichWithTeamAliases attempts to resolve team names from the query using the provided resolver.
func (q *GameQuery) EnrichWithTeamAliases(resolver TeamAliasResolver) {
	normalized := strings.ToLower(q.RawQuery)
	tokens := strings.Fields(normalized)

	for i := range tokens {
		if teamID, ok := resolver.ResolveTeamAlias(tokens[i], q.Season); ok {
			if q.HomeTeamID == nil {
				id := string(teamID)
				q.HomeTeamID = &id
			} else if q.AwayTeamID == nil && *q.HomeTeamID != string(teamID) {
				id := string(teamID)
				q.AwayTeamID = &id
			}
		}

		if i < len(tokens)-1 {
			twoWord := tokens[i] + " " + tokens[i+1]
			if teamID, ok := resolver.ResolveTeamAlias(twoWord, q.Season); ok {
				if q.HomeTeamID == nil {
					id := string(teamID)
					q.HomeTeamID = &id
				} else if q.AwayTeamID == nil && *q.HomeTeamID != string(teamID) {
					id := string(teamID)
					q.AwayTeamID = &id
				}
			}
		}
	}
}
