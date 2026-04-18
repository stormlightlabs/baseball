import { redirect } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const load: PageLoad = ({ params, url }) => {
  const q = url.searchParams.get('q');
  const encodedId = encodeURIComponent(params.id);
  const base = `/players/${encodedId}/batting`;

  if (!q) {
    throw redirect(307, base);
  }

  const qs = new URLSearchParams({ q }).toString();
  throw redirect(307, `${base}?${qs}`);
};
