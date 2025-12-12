/*
TODO:

  - /v1/players/{player_id}/stats/batting/advanced

    ?season=2024
    &team_id=LAD
    &split=none|home_away|pitcher_handed|month|batting_order
    &provider=fangraphs|bbref|internal

  - /v1/players/{player_id}/stats/pitching/advanced

    ?season=2024
    &team_id=LAD
    &provider=fangraphs|bbref|internal

  - /v1/players/{player_id}/stats/war

    ?season=2024
    &team_id=LAD
    &provider=fangraphs|bbref|internal

  - /v1/games/{game_id}/plate-appearances/leverage

    ?min_li=1.5

  - /v1/players/{player_id}/leverage

    ?season=2024
    &role=pitcher|batter

  - /v1/players/{player_id}/leverage/plate-appearances

    ?season=2024
    &min_li=1.5

  - /v1/parks/{park_id}/factors

    ?season=2024

  - /v1/parks/{park_id}/factors/series

    ?from_season=2000
    &to_season=2024

  - /v1/seasons/{season}/park-factors

    ?type=runs|hr
*/
package api
