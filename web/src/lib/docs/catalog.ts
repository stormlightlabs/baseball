import type { Component } from 'svelte';
import GithubSlugger from 'github-slugger';

export type DocGroup = 'Get Started' | 'API Reference' | 'Data & Architecture' | 'Project';

export type TocEntry = { depth: 2 | 3; text: string; id: string };

export type DocEntry = {
  slug: string;
  sourceFile: string;
  title: string;
  navTitle: string;
  group: DocGroup;
  toc: TocEntry[];
  component: Component;
};

type MarkdownModule = { default: Component; metadata?: { title?: string; updated?: string } };

const GROUP_ORDER: DocGroup[] = ['Get Started', 'API Reference', 'Data & Architecture', 'Project'];
const GET_STARTED_SLUGS = new Set(['introduction', 'about', 'apps']);
const GET_STARTED_ORDER = ['introduction', 'about', 'apps'] as const;
const GET_STARTED_INDEX: Record<string, number> = Object.fromEntries(
  GET_STARTED_ORDER.map((slug, index) => [slug, index])
);

const DATA_ARCHITECTURE_SLUGS = new Set(['id-crosswalk', 'pitches', 'statistical-methodology']);

const NAV_TITLE_OVERRIDES: Record<string, string> = {
  'api-achievements': 'Achievements',
  'api-auth': 'Auth',
  'api-awards-postseason': 'Awards & Postseason',
  'api-computed': 'Computed & Advanced',
  'api-derived-advanced': 'Derived & Advanced',
  'api-game-context': 'Game Context',
  'api-games': 'Games',
  'api-league-coverage': 'League Coverage',
  'api-meta-utility': 'Meta & Utility',
  'api-mlb-proxy': 'MLB Proxy',
  'api-parks-umpires-managers': 'Parks, Umpires & Managers',
  'api-per-game-aggregations': 'Per-Game Aggregations',
  'api-pitches': 'Pitches',
  'api-play-by-play': 'Play-by-Play & Events',
  'api-players': 'Players',
  'api-search': 'Search',
  'api-stats': 'Stats & Leaders',
  'api-teams': 'Teams & Franchises',
  'id-crosswalk': 'ID Crosswalk',
  pitches: 'Pitch Sequencing',
  'statistical-methodology': 'Statistical Methodology'
};

const markdownModules = import.meta.glob('../../routes/docs/*.md', { eager: true }) as Record<string, MarkdownModule>;
const markdownSources = import.meta.glob('../../routes/docs/*.md', {
  eager: true,
  query: '?raw',
  import: 'default'
}) as Record<string, string>;

function inferGroup(slug: string): DocGroup {
  if (GET_STARTED_SLUGS.has(slug)) return 'Get Started';
  if (slug.startsWith('api-')) return 'API Reference';
  if (DATA_ARCHITECTURE_SLUGS.has(slug)) return 'Data & Architecture';
  return 'Project';
}

function inferSlug(path: string): string {
  const filename = path.split('/').pop();
  if (!filename) return path;
  return filename.replace(/\.md$/u, '');
}

function stripInlineMarkdown(text: string): string {
  return text
    .replaceAll(/!\[([^\]]*)\]\([^)]+\)/g, '$1')
    .replaceAll(/\[([^\]]+)\]\([^)]+\)/g, '$1')
    .replaceAll(/`([^`]+)`/g, '$1')
    .replaceAll(/[*_~]/g, '')
    .replaceAll(/<\/?[^>]+>/g, '')
    .trim();
}

function inferTitle(source: string, metadataTitle: string | undefined, slug: string): string {
  if (metadataTitle) return metadataTitle;
  const heading = source.match(/^#\s+(.+)$/mu)?.[1];
  if (heading) return stripInlineMarkdown(heading);
  return slug
    .split('-')
    .map((part) => `${part[0]?.toUpperCase() ?? ''}${part.slice(1)}`)
    .join(' ');
}

function inferNavTitle(slug: string, title: string): string {
  if (NAV_TITLE_OVERRIDES[slug]) return NAV_TITLE_OVERRIDES[slug];
  return title.replace(/\s+api\s+overview$/iu, '').trim();
}

function extractToc(source: string): TocEntry[] {
  const toc: TocEntry[] = [];
  const slugger = new GithubSlugger();
  let insideFence = false;

  for (const line of source.split(/\r?\n/u)) {
    const trimmed = line.trim();
    if (trimmed.startsWith('```')) {
      insideFence = !insideFence;
      continue;
    }
    if (insideFence) continue;

    const match = /^(##|###)\s+(.+)$/u.exec(trimmed);
    if (!match) continue;

    const depth = match[1].length as 2 | 3;
    const text = stripInlineMarkdown(match[2]);
    if (!text) continue;

    const id = slugger.slug(text);

    toc.push({ depth, text, id });
  }

  return toc;
}

const builtDocs: DocEntry[] = Object.entries(markdownModules).map(([path, module]) => {
  const slug = inferSlug(path);
  const source = markdownSources[path] ?? '';
  const title = inferTitle(source, module.metadata?.title, slug);

  return {
    slug,
    sourceFile: `docs/${slug}.md`,
    title,
    navTitle: inferNavTitle(slug, title),
    group: inferGroup(slug),
    toc: extractToc(source),
    component: module.default
  };
});

builtDocs.sort((a, b) => {
  const groupDelta = GROUP_ORDER.indexOf(a.group) - GROUP_ORDER.indexOf(b.group);
  if (groupDelta !== 0) return groupDelta;
  if (a.group === 'Get Started') {
    const aIndex = GET_STARTED_INDEX[a.slug] ?? Number.POSITIVE_INFINITY;
    const bIndex = GET_STARTED_INDEX[b.slug] ?? Number.POSITIVE_INFINITY;
    return aIndex - bIndex;
  }
  return a.navTitle.localeCompare(b.navTitle);
});

export const DOCS = builtDocs;
export const DOC_SLUGS = DOCS.map((doc) => doc.slug);
export const DOCS_BY_SLUG: Record<string, DocEntry> = Object.fromEntries(DOCS.map((doc) => [doc.slug, doc]));
export const FIRST_DOC_SLUG = (function () {
  if (DOCS_BY_SLUG.introduction) {
    return 'introduction';
  } else if (DOCS_BY_SLUG['api-players']) {
    return 'api-players';
  } else if (DOCS.length > 0) {
    return DOCS[0].slug;
  } else {
    return 'api-players';
  }
})();

export const DOC_GROUPS: Array<{ group: DocGroup; docs: DocEntry[] }> = GROUP_ORDER.map((group) => ({
  group,
  docs: DOCS.filter((doc) => doc.group === group)
}));
