import { redirect } from '@sveltejs/kit';
import { FIRST_DOC_SLUG } from '$lib/docs/catalog';

export const prerender = true;

export function load() {
  throw redirect(307, `/docs/${FIRST_DOC_SLUG}`);
}
