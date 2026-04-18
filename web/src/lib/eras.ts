export type Era = { code: string; label: string; from: number; to: number } & {
  /** Color bucket used for styling — drives component variant selection */
  color: 'warning' | 'secondary' | 'primary' | 'default';
  /** Known data-coverage caveats for this era */
  caveat?: string;
};

export const STATIC_ERAS: Era[] = [
  {
    code: 'fed',
    label: 'Federal League Era',
    from: 1914,
    to: 1915,
    color: 'warning',
    caveat: 'Third major league; Retrosheet play-event density is sparse.'
  },
  {
    code: 'nlg',
    label: 'Negro Leagues Era',
    from: 1935,
    to: 1949,
    color: 'warning',
    caveat: 'Retrosheet data available but coverage varies significantly by team and season.'
  },
  { code: 'boomer', label: 'Baby Boomer Era', from: 1950, to: 1962, color: 'secondary' },
  {
    code: 'pitcher',
    label: 'Pitcher Era',
    from: 1963,
    to: 1968,
    color: 'secondary',
    caveat: 'Dominant pitching environment; mound lowered after 1968.'
  },
  {
    code: 'turf',
    label: 'Turf Time',
    from: 1969,
    to: 1993,
    color: 'secondary',
    caveat: 'Artificial turf era; significant expansion and free agency changes.'
  },
  {
    code: 'steroid',
    label: 'Steroid Era',
    from: 1994,
    to: 2004,
    color: 'primary',
    caveat: 'Enhanced performance period; HR records should be interpreted with caution.'
  },
  { code: 'moneyball', label: 'Moneyball Era', from: 2005, to: 2012, color: 'primary' },
  { code: 'statcast', label: 'Statcast Era', from: 2013, to: 2019, color: 'primary' },
  {
    code: 'modern',
    label: 'Modern Era',
    from: 2020,
    to: 2025,
    color: 'primary',
    caveat: 'Pitch clock and pace-of-play rule changes affect comparison with prior eras.'
  }
];

/** Returns the static era for a given year, or undefined if out of range. */
export function eraForYear(year: number): Era | undefined {
  return STATIC_ERAS.find((e) => year >= e.from && year <= e.to);
}

/** Returns all eras that overlap a year range [from, to] (inclusive). */
export function erasInRange(from: number, to: number): Era[] {
  return STATIC_ERAS.filter((e) => e.from <= to && e.to >= from);
}

/** Returns a short display label for a year range spanning multiple eras. */
export function eraSpanLabel(from: number, to: number): string {
  const eras = erasInRange(from, to);
  const firstEra = eras[0];
  if (!firstEra) return '';
  if (eras.length === 1) return firstEra.label;
  const lastEra = eras.at(-1);
  return lastEra ? `${firstEra.label} → ${lastEra.label}` : firstEra.label;
}

export type WinExpEra = { era: string; label?: string; year_from: number; year_to: number };

/** Fetches dynamic win-expectancy era ranges from the API. */
export async function fetchWinExpEras(apiFetch: (path: string) => Promise<unknown>): Promise<WinExpEra[]> {
  try {
    const result = await apiFetch('/win-expectancy/eras');
    if (Array.isArray(result)) return result as WinExpEra[];
    const obj = result as Record<string, unknown>;
    if (Array.isArray(obj.eras)) return obj.eras as WinExpEra[];
    if (Array.isArray(obj.data)) return obj.data as WinExpEra[];
    return [];
  } catch {
    return [];
  }
}
