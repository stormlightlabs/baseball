export type UnknownRecord = Record<string, unknown>;

export function toObjectNullable(value: unknown): UnknownRecord | undefined {
  if (value != null && typeof value === 'object' && !Array.isArray(value)) {
    return value as UnknownRecord;
  }
  return undefined;
}

export function toObject(value: unknown): UnknownRecord {
  return toObjectNullable(value) ?? {};
}

export function toArray(value: unknown): unknown[] {
  if (Array.isArray(value)) return value;
  return [];
}

export function toRecordArray(value: unknown): UnknownRecord[] {
  return toArray(value)
    .map((entry) => toObjectNullable(entry))
    .filter((entry): entry is UnknownRecord => entry !== undefined);
}

export function toString(value: unknown): string | undefined {
  if (value == null) return undefined;
  if (typeof value === 'string') {
    const trimmed = value.trim();
    return trimmed.length > 0 ? trimmed : undefined;
  }
  if (typeof value === 'number' || typeof value === 'boolean') {
    return String(value);
  }
  return undefined;
}

export function toNumber(value: unknown): number | undefined {
  if (typeof value === 'number' && Number.isFinite(value)) return value;
  if (typeof value === 'string' && value.trim().length > 0) {
    const parsed = Number(value);
    if (Number.isFinite(parsed)) return parsed;
  }
  return undefined;
}

export function toBoolean(value: unknown): boolean | undefined {
  if (typeof value === 'boolean') return value;
  if (typeof value === 'number') {
    if (value === 1) return true;
    if (value === 0) return false;
  }
  if (typeof value === 'string') {
    const normalized = value.trim().toLowerCase();
    if (normalized === 'true' || normalized === '1') return true;
    if (normalized === 'false' || normalized === '0') return false;
  }
  return undefined;
}

export function toStringArray(value: unknown): string[] {
  return toArray(value)
    .map((entry) => toString(entry))
    .filter((entry): entry is string => entry !== undefined);
}
