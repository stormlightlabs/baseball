import { error } from '@sveltejs/kit';
import { DOCS_BY_SLUG, DOC_SLUGS } from '$lib/docs/catalog';

export const prerender = true;

export function entries() {
  return DOC_SLUGS.map((slug) => ({ slug }));
}

export function load({ params }) {
  const doc = DOCS_BY_SLUG[params.slug];
  if (!doc) {
    error(404, `Unknown document slug: ${params.slug}`);
  }

  return {
    slug: doc.slug,
    title: doc.title,
    navTitle: doc.navTitle,
    group: doc.group,
    toc: doc.toc,
    sourceFile: doc.sourceFile
  };
}
