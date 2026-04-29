// # Big Fly Dev Tools
//
// Command:
//
//	baseball-dev
//
// Subcommands:
//
//	generate  Build OpenAPI/Hurl artifacts from internal/docs/swagger.{yaml,json}
//	run       Execute generated Hurl tests concurrently with colored output
//	cleanup   Remove generated artifacts/tests from test directories
//	swagger   Generate/fix/format/clean Swagger docs (alias: swag)
//
// Default output layout:
//
//	test/artifacts/
//	  openapi.3.0.yaml
//	  openapi.3.1.yaml
//	  openapi.3.1.sanitized.yaml
//	  all.generated.hurl
//
//	test/generated/
//	  *.hurl (flattened, one request per file)
//
// Generation pipeline:
//
//	Swagger 2.0 -> OpenAPI 3.0 -> OpenAPI 3.1 -> sanitize -> openapi-to-hurl -> split
//
// Required/open parameters:
//
//	generate forwards query/path parameter behavior to openapi-to-hurl:
//	  --query-params none|required|all   (default: required)
//	  --path-params default|variables    (default: variables)
//
//	This means required query params are included in generated Hurl by default.
//	Path params use named variables by default (more readable filenames).
//
// Representative datapoints (2026-04-28):
//
//	Path params:
//	  game_id: ARI202311010
//	  player_id: semim001, lowen001, schwk001
//	  team_id: TEX, ARI, PHI, HOU, MIN
//	  park_id: LOS03, BOS07, CHI11, HOU03, CLE08
//	  play_num/event_seq examples from plays: gid=WEG193709130, pn=82, event=99
//
//	Query params:
//	  season range (serving tables): 2023..2023
//	  date range (serving tables): 20230330..20231101
//	  league values (leaders): batting=AL,ML,NL; pitching=AL,ML
//	  top_bot: 0,1
//	  inning range: 1..20
//
// Notes:
//   - Generated split filenames intentionally strip host placeholders like {{host}}
//     and {{BASE_URL}} to avoid host-prefixed filenames.
//   - --cleanup removes transitional files after successful generation:
//     openapi.3.0.yaml, openapi.3.1.sanitized.yaml, all.generated.hurl
package main
