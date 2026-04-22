export function parseTimestamp(value: string | null | undefined): Date | null {
  if (!value) return null;
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return null;
  return parsed;
}

export function formatRelativeTimeFromNow(value: string | null | undefined): string | null {
  const timestamp = parseTimestamp(value);
  if (!timestamp) return null;

  const diffMs = Date.now() - timestamp.getTime();
  const future = diffMs < 0;
  const absMs = Math.abs(diffMs);

  const minuteMs = 60_000;
  const hourMs = 60 * minuteMs;
  const dayMs = 24 * hourMs;

  let amount: number;
  let unit: Intl.RelativeTimeFormatUnit;

  if (absMs >= dayMs) {
    amount = Math.round(absMs / dayMs);
    unit = 'day';
  } else if (absMs >= hourMs) {
    amount = Math.round(absMs / hourMs);
    unit = 'hour';
  } else {
    amount = Math.max(1, Math.round(absMs / minuteMs));
    unit = 'minute';
  }

  const formatter = new Intl.RelativeTimeFormat(undefined, { numeric: 'auto' });
  return formatter.format(future ? amount : -amount, unit);
}

export function formatDateTimeLabel(value: string | null | undefined): string | null {
  const timestamp = parseTimestamp(value);
  if (!timestamp) return null;
  return timestamp.toLocaleString(undefined, { month: 'short', day: 'numeric', hour: 'numeric', minute: '2-digit' });
}
