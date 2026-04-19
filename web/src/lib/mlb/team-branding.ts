// TODO: this could be an enum
const MLB_TEAM_PRIMARY_HEX: Record<string, string> = {
  ARI: '#A71930',
  ATL: '#CE1141',
  BAL: '#DF4601',
  BOS: '#BD3039',
  CHC: '#0E3386',
  CWS: '#27251F',
  CIN: '#C6011F',
  CLE: '#E31937',
  COL: '#33006F',
  DET: '#0C2340',
  HOU: '#002D62',
  KC: '#004687',
  LAA: '#BA0021',
  LAD: '#005A9C',
  MIA: '#000000',
  MIL: '#12284B',
  MIN: '#002B5C',
  NYM: '#002D72',
  NYY: '#132448',
  ATH: '#003831',
  PHI: '#E81828',
  PIT: '#27251F',
  SD: '#2F241D',
  SF: '#FD5A1E',
  SEA: '#0C2C56',
  STL: '#C41E3A',
  TB: '#092C5C',
  TEX: '#003278',
  TOR: '#134A8E',
  WSH: '#AB0003'
};

const MLB_CODE_ALIASES: Record<string, string> = {
  ANA: 'LAA',
  CAL: 'LAA',
  CHA: 'CWS',
  CHN: 'CHC',
  CLN: 'CLE',
  FLO: 'MIA',
  KCA: 'KC',
  KCN: 'KC',
  LAN: 'LAD',
  BRO: 'LAD',
  MON: 'WSH',
  NYA: 'NYY',
  NYN: 'NYM',
  PHA: 'ATH',
  OAK: 'ATH',
  SDN: 'SD',
  SFN: 'SF',
  SLN: 'STL',
  TBA: 'TB',
  TBD: 'TB',
  WAS: 'WSH',
  WSN: 'WSH'
};

export function normalizeMlbTeamCode(rawCode: string | undefined): string | undefined {
  if (!rawCode) return undefined;
  const upper = rawCode.toUpperCase();
  if (MLB_TEAM_PRIMARY_HEX[upper]) return upper;
  return MLB_CODE_ALIASES[upper];
}

export function teamPrimaryHexFor(code: string | undefined): string | undefined {
  const normalized = normalizeMlbTeamCode(code);
  if (!normalized) return undefined;
  return MLB_TEAM_PRIMARY_HEX[normalized];
}
