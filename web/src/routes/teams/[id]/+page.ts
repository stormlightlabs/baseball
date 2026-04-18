import { redirect } from '@sveltejs/kit';

export const load = ({ params, url }: { params: { id: string }; url: URL }) => {
  const encodedId = encodeURIComponent(params.id);
  const base = `/teams/${encodedId}/overview`;
  const qs = url.search;
  throw redirect(307, qs ? `${base}${qs}` : base);
};
