import { redirect } from '@sveltejs/kit';
import { apiHref } from '$lib/api';

export const load = () => {
  throw redirect(307, apiHref('/docs/'));
};
