import { redirect } from '@sveltejs/kit';

export const load = ({ url }: { url: URL }) => {
  const year = new Date().getFullYear();
  const qs = url.searchParams.toString();
  const suffix = qs ? `?${qs}` : '';
  throw redirect(307, `/seasons/${year}${suffix}`);
};
