import { redirect } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const load: PageLoad = ({ params, url }) => {
  const encodedId = encodeURIComponent(params.id);
  const base = `/games/${encodedId}/overview`;
  const qs = url.search;
  throw redirect(307, qs ? `${base}${qs}` : base);
};
