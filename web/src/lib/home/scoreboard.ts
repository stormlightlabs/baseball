import { toArray, toBoolean, toNumber, toObject, toString } from '$lib/common/converters';
import { teamPrimaryHexFor } from '$lib/mlb/team-branding';

export type ScoreboardTeam = {
  id?: string;
  name?: string;
  abbreviation: string;
  score?: number;
  color?: string;
  wins?: number;
  losses?: number;
};

export type ScoreboardGame = {
  id: string;
  date?: string;
  gamePk?: number;
  retrosheetGameId?: string;
  venue?: string;
  statusText: string;
  statusShort: 'LIVE' | 'FINAL' | 'SCHEDULED';
  scheduledLabel?: string;
  isInProgress: boolean;
  isFinal: boolean;
  away: ScoreboardTeam;
  home: ScoreboardTeam;
};

export type ScoreboardSnapshot = {
  date: string;
  nextGameDate?: string;
  gamesInProgress: number;
  games: ScoreboardGame[];
};

function normalizeColor(value: unknown): string | undefined {
  const raw = toString(value);
  if (!raw) return undefined;
  const normalized = raw.startsWith('#') ? raw : `#${raw}`;
  if (/^#[\dA-Fa-f]{3}$/.test(normalized) || /^#[\dA-Fa-f]{6}$/.test(normalized)) {
    return normalized;
  }
  return undefined;
}

function toTeamAbbreviation(raw: unknown, fallbackName?: string): string {
  const direct = toString(raw)?.toUpperCase();
  if (direct && direct.length <= 4) return direct;
  if (fallbackName) {
    const cleaned = fallbackName.replaceAll(/[^A-Za-z]/g, '');
    if (cleaned.length >= 3) return cleaned.slice(0, 3).toUpperCase();
  }
  return 'TBD';
}

function toLocalDateISO(date: Date): string {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
}

function maybeRetrosheetGameID(value: unknown): string | undefined {
  const raw = toString(value);
  if (!raw) return undefined;
  const upper = raw.toUpperCase();
  if (/^[A-Z]{3}\d{8}\d?$/.test(upper)) {
    return upper;
  }
  return undefined;
}

function normalizeTeam(raw: unknown): ScoreboardTeam {
  const row = toObject(raw);
  const name = toString(row.name) ?? toString(row.team_name) ?? toString(row.teamName) ?? toString(row.clubName);
  const palette = toObject(row.palette);
  const teamID = toString(row.id) ?? toString(row.team_id) ?? toString(row.teamId);
  const abbreviation = toTeamAbbreviation(
    row.abbreviation ?? row.abbr ?? row.team_code ?? row.teamCode ?? row.fileCode ?? row.triCode,
    name
  );
  const mappedColor = teamPrimaryHexFor(abbreviation);
  return {
    id: teamID,
    name,
    abbreviation,
    score: toNumber(row.score ?? row.runs),
    color:
      normalizeColor(
        row.color ?? row.primary_color ?? row.team_color ?? row.teamColor ?? row.hex ?? row.hex_color ?? palette.primary
      ) ?? mappedColor,
    wins: toNumber(row.wins),
    losses: toNumber(row.losses)
  };
}

function extractTeamSide(game: Record<string, unknown>, side: 'away' | 'home'): ScoreboardTeam {
  const teams = toObject(game.teams);
  const nestedSide = toObject(teams[side]);
  const direct = toObject(game[side] ?? game[`${side}_team`] ?? game[`${side}Team`]);

  let team = normalizeTeam(direct);
  if (team.abbreviation === 'TBD' && Object.keys(nestedSide).length > 0) {
    const nestedTeam = toObject(nestedSide.team);
    team = normalizeTeam({ ...nestedTeam, ...nestedSide });
  }

  if (team.score == null) {
    team.score = toNumber(game[`${side}_score`] ?? game[`${side}Score`] ?? nestedSide.score);
  }

  return team;
}

function statusFromGame(
  game: Record<string, unknown>
): Pick<ScoreboardGame, 'statusText' | 'statusShort' | 'isInProgress' | 'isFinal' | 'scheduledLabel'> {
  const status = toObject(game.status);
  const linescore = toObject(game.linescore);

  const detailed = toString(status.detailedState) ?? toString(status.description) ?? toString(game.status_text);
  const abstract =
    toString(status.abstractGameState) ??
    toString(status.abstract_state) ??
    toString(status.state) ??
    toString(game.state) ??
    toString(game.status);

  const coded = toString(status.codedGameState) ?? '';
  const abstractLower = (abstract ?? '').toLowerCase();
  const detailedLower = (detailed ?? '').toLowerCase();

  const liveFlag =
    toBoolean(game.in_progress) ??
    toBoolean(game.games_in_progress) ??
    toBoolean(game.is_live) ??
    toBoolean(status.is_live) ??
    toBoolean(status.isLive);

  const finalFlag = toBoolean(game.is_final) ?? toBoolean(status.isFinal);

  const isInProgress =
    liveFlag === true ||
    abstractLower.includes('live') ||
    abstractLower.includes('in progress') ||
    detailedLower.includes('in progress');

  const isFinal =
    finalFlag === true ||
    abstractLower.includes('final') ||
    detailedLower.includes('final') ||
    detailedLower.includes('game over') ||
    coded.toUpperCase() === 'F';

  if (isInProgress && !isFinal) {
    const inningHalf = toString(linescore.inningHalf ?? game.inning_half ?? game.inningHalf);
    const inning = toString(
      linescore.currentInningOrdinal ??
        linescore.currentInning ??
        game.inning_ordinal ??
        game.inningOrdinal ??
        game.inning
    );
    const outs = toNumber(linescore.outs ?? game.outs);

    const parts: string[] = [];
    const inningText = [inningHalf, inning].filter((entry): entry is string => entry != null).join(' ');
    if (inningText) parts.push(inningText);
    if (outs != null) parts.push(`${outs} out${outs === 1 ? '' : 's'}`);

    return {
      statusText: parts.join(' · ') || detailed || 'Live',
      statusShort: 'LIVE',
      isInProgress: true,
      isFinal: false,
      scheduledLabel: undefined
    };
  }

  if (isFinal) {
    return {
      statusText: detailed ?? 'Final',
      statusShort: 'FINAL',
      isInProgress: false,
      isFinal: true,
      scheduledLabel: undefined
    };
  }

  const scheduledLabel =
    toString(game.start_time) ??
    toString(game.game_time) ??
    toString(game.gameTime) ??
    toString(status.startTime) ??
    toString(game.gameDate) ??
    toString(game.game_date);

  return {
    statusText: scheduledLabel ?? detailed ?? 'Scheduled',
    statusShort: 'SCHEDULED',
    isInProgress: false,
    isFinal: false,
    scheduledLabel
  };
}

function normalizeGame(raw: unknown, index: number): ScoreboardGame {
  const row = toObject(raw);
  const crosswalk = toObject(row.crosswalk);
  const gamePK = toNumber(row.game_pk ?? row.gamePk);
  const retrosheetGameID =
    maybeRetrosheetGameID(row.retrosheet_game_id) ??
    maybeRetrosheetGameID(row.retrosheetGameId) ??
    maybeRetrosheetGameID(crosswalk.retrosheet_game_id) ??
    maybeRetrosheetGameID(crosswalk.retrosheetGameId) ??
    maybeRetrosheetGameID(row.retro_game_id) ??
    maybeRetrosheetGameID(row.game_id);

  const venue = toString(toObject(row.venue).name) ?? toString(row.venue_name) ?? toString(row.venueName);
  const away = extractTeamSide(row, 'away');
  const home = extractTeamSide(row, 'home');
  const gameDate = toString(row.date) ?? toString(row.game_date) ?? toString(row.gameDate);
  const status = statusFromGame(row);

  const id =
    toString(row.id) ??
    retrosheetGameID ??
    (gamePK != null ? `mlb-${gamePK}` : undefined) ??
    `${away.abbreviation}-${home.abbreviation}-${gameDate ?? `idx-${index}`}`;

  return {
    id,
    date: gameDate,
    gamePk: gamePK,
    retrosheetGameId: retrosheetGameID,
    venue,
    statusText: status.statusText,
    statusShort: status.statusShort,
    scheduledLabel: status.scheduledLabel,
    isInProgress: status.isInProgress,
    isFinal: status.isFinal,
    away,
    home
  };
}

export function normalizeScoreboardResponse(raw: unknown): ScoreboardSnapshot {
  const row = toObject(raw);
  const embeddedScoreboard = toObject(row.scoreboard);
  const dates = toArray(row.dates);
  const firstDateRow = toObject(dates[0]);

  let gamesRaw = toArray(firstDateRow.games);
  const topLevelGames = toArray(row.games);
  if (topLevelGames.length > 0) {
    gamesRaw = topLevelGames;
  } else {
    const topLevelData = toArray(row.data);
    if (topLevelData.length > 0) {
      gamesRaw = topLevelData;
    } else {
      const embeddedGames = toArray(embeddedScoreboard.games);
      if (embeddedGames.length > 0) {
        gamesRaw = embeddedGames;
      }
    }
  }

  const games = gamesRaw.map((game, index) => normalizeGame(game, index));
  const fallbackLive = games.filter((game) => game.isInProgress).length;

  const date =
    toString(row.date) ??
    toString(embeddedScoreboard.date) ??
    toString(firstDateRow.date) ??
    toLocalDateISO(new Date());

  const gamesInProgress =
    toNumber(
      row.games_in_progress ??
        row.totalGamesInProgress ??
        row.total_games_in_progress ??
        embeddedScoreboard.games_in_progress ??
        firstDateRow.totalGamesInProgress
    ) ?? fallbackLive;

  const nextGameDate =
    toString(row.next_game_date) ??
    toString(row.nextGameDate) ??
    toString(embeddedScoreboard.next_game_date) ??
    toString(firstDateRow.next_game_date);

  return { date, nextGameDate, gamesInProgress, games };
}

export function todayLocalISODate(): string {
  return toLocalDateISO(new Date());
}
