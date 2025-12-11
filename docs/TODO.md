# Bugs/Issues/Limitations

## Current Limitations

1. Game Logs Only Show Starting Lineup
    - The game-logs endpoint only returns games where the player was in the starting lineup (positions 1-9)
    - Players who appeared as substitutes won't show up in their game logs
    - This is because the games table only stores starting lineups, not all player appearances
2. No Individual Player Stats in Game Logs
    - The game-logs endpoint returns the game metadata but not the individual player's performance in each game
    - To get a true "game log" with player stats, we'd need to query batting/pitching tables by game

### Improvements

1. Add a separate table for all player game appearances (not just starters)
2. Aggregate individual game stats from play-by-play data for true game logs
