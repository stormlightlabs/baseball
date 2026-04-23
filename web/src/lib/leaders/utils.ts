import type { PaginatedResponse, Params } from '$lib/api';
import { toNumber as toNum, toObject, toRecordArray } from '$lib/common/converters';
import type { ExtractedRows } from './types';

export function extractRows(payload: unknown, keys: string[]): ExtractedRows {
  if (Array.isArray(payload)) {
    const rows = toRecordArray(payload);
    return { rows, page: 1, perPage: rows.length === 0 ? 20 : rows.length, total: rows.length };
  }

  const objectPayload = toObject(payload);
  let rows: Array<Record<string, unknown>> = [];

  for (const key of keys) {
    const candidates = toRecordArray(objectPayload[key]);
    if (candidates.length > 0) {
      rows = candidates;
      break;
    }
  }

  const page = toNum(objectPayload.page) ?? 1;
  const perPage = toNum(objectPayload.per_page) ?? (rows.length === 0 ? 20 : rows.length);
  const total = toNum(objectPayload.total) ?? rows.length;

  return { rows, page, perPage, total };
}

export function formatMetric(value: number | undefined, stat: string): string {
  if (value == null) return '—';

  const lowered = stat.toLowerCase();
  const rateStats = new Set([
    'avg',
    'obp',
    'slg',
    'ops',
    'era',
    'whip',
    'iso',
    'babip',
    'woba',
    'k_rate',
    'bb_rate',
    'k_per_9',
    'bb_per_9',
    'hr_per_9',
    'fip',
    'war',
    'fpct'
  ]);

  if (lowered === 'wrc_plus') return Math.round(value).toLocaleString();
  if (lowered === 'ip') return value.toFixed(1);
  if (rateStats.has(lowered)) return value.toFixed(3);

  return Math.round(value).toLocaleString();
}

export function toInnings(value: unknown): number | undefined {
  const outs = toNum(value);
  if (outs == null) return undefined;
  return outs / 3;
}

export function endpointWithQuery(path: string, params: Params): string {
  const query = new URLSearchParams();

  for (const [key, value] of Object.entries(params)) {
    if (value == null) continue;
    const text = String(value).trim();
    if (text.length === 0) continue;
    query.set(key, text);
  }

  const queryText = query.toString();
  if (!queryText) return path;
  return `${path}?${queryText}`;
}

export function toSampleJson(payload: unknown): string {
  try {
    const text = JSON.stringify(payload, null, 2);
    if (text.length <= 6000) return text;
    return `${text.slice(0, 6000)}\n...`;
  } catch {
    return '{"error":"Unable to serialize response"}';
  }
}

export function toErrorMessage(errorValue: unknown, fallback: string): string {
  if (errorValue instanceof Error) {
    const message = errorValue.message.trim();
    if (message.length > 0) return message;
  }

  return fallback;
}

export function fmtInt(value: number | undefined): string {
  if (value == null) return '—';
  return Math.round(value).toLocaleString();
}

export function pad2(value: number): string {
  return String(value).padStart(2, '0');
}

export function emptyPage<T>(): PaginatedResponse<T> {
  return { data: [], page: 1, per_page: 1, total: 0 };
}

export function fmtFloat(value: number | undefined, digits = 2): string {
  if (value == null) return '—';
  return value.toFixed(digits);
}

export function fmtSigned(value: number | undefined): string {
  if (value == null) return '—';
  if (value > 0) return `+${Math.round(value)}`;
  return String(Math.round(value));
}
