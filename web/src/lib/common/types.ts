export type LeagueFilter = 'both' | 'AL' | 'NL';

export function parseLeague(value: string | null): LeagueFilter {
  if (!value) return 'both';
  const normalized = value.toLowerCase();
  if (normalized === 'al') return 'AL';
  if (normalized === 'nl') return 'NL';
  return 'both';
}
