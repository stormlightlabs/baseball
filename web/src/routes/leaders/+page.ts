import { redirect } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const load: PageLoad = ({ url }) => {
  const qs = url.search;
  throw redirect(307, qs ? `/leaders/quick${qs}` : '/leaders/quick');
};
