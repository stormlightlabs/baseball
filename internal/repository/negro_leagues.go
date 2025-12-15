package repository

import (
	"context"
	"database/sql"

	"stormlightlabs.org/baseball/internal/cache"
	"stormlightlabs.org/baseball/internal/core"
)

// NegroLeaguesRepository provides access to Negro Leagues historical data.
// It wraps the game, play, and team repositories with convenience methods for Negro Leagues data.
// Negro Leagues data is stored in the main games and plays tables with specific game_type markers.
type NegroLeaguesRepository struct {
	gameRepo *GameRepository
	playRepo *PlayRepository
	teamRepo *TeamRepository
}

// NewNegroLeaguesRepository creates a new Negro Leagues repository instance.
func NewNegroLeaguesRepository(db *sql.DB, cacheClient *cache.Client) *NegroLeaguesRepository {
	return &NegroLeaguesRepository{
		gameRepo: NewGameRepository(db, cacheClient),
		playRepo: NewPlayRepository(db),
		teamRepo: NewTeamRepository(db, cacheClient),
	}
}

// negroLeagues holds the league codes for Negro Leagues
var negroLeagues = []core.LeagueID{"NAL", "NNL", "NN2", "ECL", "ANL", "EWL", "NSL", "IND"}

// ListGames returns Negro Leagues games with filtering and pagination.
func (r *NegroLeaguesRepository) ListGames(ctx context.Context, filter core.GameFilter) ([]core.Game, error) {
	if filter.League == nil {
		allGames := []core.Game{}
		for _, league := range negroLeagues {
			leagueFilter := filter
			l := league
			leagueFilter.League = &l
			games, err := r.gameRepo.List(ctx, leagueFilter)
			if err != nil {
				return nil, err
			}
			allGames = append(allGames, games...)
		}
		return allGames, nil
	}
	return r.gameRepo.List(ctx, filter)
}

// CountGames returns the total count of Negro Leagues games matching the filter.
func (r *NegroLeaguesRepository) CountGames(ctx context.Context, filter core.GameFilter) (int, error) {
	if filter.League == nil {
		total := 0
		for _, league := range negroLeagues {
			leagueFilter := filter
			l := league
			leagueFilter.League = &l
			count, err := r.gameRepo.Count(ctx, leagueFilter)
			if err != nil {
				return 0, err
			}
			total += count
		}
		return total, nil
	}
	return r.gameRepo.Count(ctx, filter)
}

// ListTeamSeasons returns teams that played in the Negro Leagues.
func (r *NegroLeaguesRepository) ListTeamSeasons(ctx context.Context, filter core.TeamFilter) ([]core.TeamSeason, error) {
	if filter.League == nil {
		allTeams := []core.TeamSeason{}
		for _, league := range negroLeagues {
			leagueFilter := filter
			l := league
			leagueFilter.League = &l
			teams, err := r.teamRepo.ListTeamSeasons(ctx, leagueFilter)
			if err != nil {
				return nil, err
			}
			allTeams = append(allTeams, teams...)
		}
		return allTeams, nil
	}
	return r.teamRepo.ListTeamSeasons(ctx, filter)
}

// CountTeamSeasons returns the count of unique team-season combinations.
func (r *NegroLeaguesRepository) CountTeamSeasons(ctx context.Context, filter core.TeamFilter) (int, error) {
	if filter.League == nil {
		total := 0
		for _, league := range negroLeagues {
			leagueFilter := filter
			l := league
			leagueFilter.League = &l
			count, err := r.teamRepo.CountTeamSeasons(ctx, leagueFilter)
			if err != nil {
				return 0, err
			}
			total += count
		}
		return total, nil
	}
	return r.teamRepo.CountTeamSeasons(ctx, filter)
}

// ListPlays returns play-by-play data from Negro Leagues games.
func (r *NegroLeaguesRepository) ListPlays(ctx context.Context, filter core.PlayFilter) ([]core.Play, error) {
	if filter.League == nil {
		leagueFilter := filter
		leagueFilter.Leagues = negroLeagues
		return r.playRepo.List(ctx, leagueFilter)
	}
	return r.playRepo.List(ctx, filter)
}

// CountPlays returns the count of plays in Negro Leagues games.
func (r *NegroLeaguesRepository) CountPlays(ctx context.Context, filter core.PlayFilter) (int, error) {
	if filter.League == nil {
		leagueFilter := filter
		leagueFilter.Leagues = negroLeagues
		return r.playRepo.Count(ctx, leagueFilter)
	}
	return r.playRepo.Count(ctx, filter)
}

// GetSeasonSchedule returns all Negro Leagues games for a specific season.
func (r *NegroLeaguesRepository) GetSeasonSchedule(ctx context.Context, year core.SeasonYear, league *core.LeagueID, p core.Pagination) ([]core.Game, error) {
	filter := core.GameFilter{
		Season:     &year,
		League:     league,
		Pagination: p,
	}
	return r.ListGames(ctx, filter)
}

// GetTeamGames returns all games for a specific team in a season.
func (r *NegroLeaguesRepository) GetTeamGames(ctx context.Context, teamID core.TeamID, year core.SeasonYear, p core.Pagination) ([]core.Game, error) {
	filter := core.GameFilter{
		Season:     &year,
		Pagination: p,
	}

	homeFilter := filter
	homeFilter.HomeTeam = &teamID
	homeGames, err := r.gameRepo.List(ctx, homeFilter)
	if err != nil {
		return nil, err
	}

	awayFilter := filter
	awayFilter.AwayTeam = &teamID
	awayGames, err := r.gameRepo.List(ctx, awayFilter)
	if err != nil {
		return nil, err
	}

	gameMap := make(map[core.GameID]core.Game)
	for _, g := range homeGames {
		gameMap[g.ID] = g
	}
	for _, g := range awayGames {
		gameMap[g.ID] = g
	}

	games := make([]core.Game, 0, len(gameMap))
	for _, g := range gameMap {
		games = append(games, g)
	}
	return games, nil
}
