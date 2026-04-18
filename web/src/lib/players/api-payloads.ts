export type PlayerStatsPayload<T> = { career?: T; seasons?: T[]; data?: T[] } | null | undefined;

export type ApiPlayerPayload = {
  id: string;
  name?: string;
  first_name?: string;
  last_name?: string;
  position?: string;
  primary_position?: string;
  positions?: string | string[];
  birth_date?: string;
  birth_year?: number;
  birth_month?: number;
  birth_day?: number;
  birth_city?: string;
  birth_state?: string;
  bats?: string;
  throws?: string;
  debut_year?: number;
  final_year?: number;
  latest_season?: number;
};

export type HallOfFamePayload<T> = { records?: T[] | null; data?: T[] };
