<script lang="ts">
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import favicon from '$lib/assets/favicon.svg';
  import AppFooter from '$lib/components/AppFooter.svelte';
  import DataNavigationMenu from '$lib/components/DataNavigationMenu.svelte';
  import { meta } from '$lib/meta.svelte.js';
  import '@fontsource-variable/google-sans';
  import '@fontsource-variable/google-sans-code';
  import '@fontsource-variable/inter';
  import { onMount } from 'svelte';
  import './layout.css';

  let { children } = $props();

  onMount(() => meta.init());

  const MAIN_LINKS = [
    { href: '/', label: 'Home' },
    { href: '/compare', label: 'Compare' },
    { href: '/docs', label: 'About' }
  ] as const;

  const DATA_LINKS = [
    { href: '/players', label: 'Players' },
    { href: '/teams', label: 'Teams' },
    { href: '/games', label: 'Games' },
    { href: '/seasons', label: 'Seasons' },
    { href: '/leaders', label: 'Leaders' },
    { href: '/data', label: 'Sources' }
  ] as const satisfies Array<{ href: string; label: string }>;

  const DATA_MENU_ITEMS = [
    { href: '/players', label: 'Players', description: 'Player search, profiles, splits, streaks, and season logs.' },
    { href: '/teams', label: 'Teams', description: 'Team/franchise lookup, rosters, schedules, and run differential.' },
    { href: '/games', label: 'Games', description: 'Game explorer with events, plays, pitches, and win probability.' },
    { href: '/seasons', label: 'Seasons', description: 'Season schedules, awards, postseason, and park factors.' },
    { href: '/leaders', label: 'Leaders', description: 'Batting, pitching, and advanced leaderboard endpoints.' },
    { href: '/data', label: 'Sources', description: 'Dataset provenance, ingestion notes, and source metadata.' }
  ];

  const API_DOCS_ROUTE = '/explorer' as const;
  type AppPath = (typeof MAIN_LINKS)[number]['href'] | (typeof DATA_LINKS)[number]['href'] | '/account';

  const BADGES: Record<AppPath, string> = {
    '/': 'home',
    '/players': 'players',
    '/teams': 'teams',
    '/games': 'games',
    '/seasons': 'seasons',
    '/leaders': 'leaders',
    '/compare': 'compare',
    '/data': 'sources',
    '/docs': 'about',
    '/account': 'account'
  };

  let pathname = $derived(page.url.pathname);
  let badge = $derived.by(() => {
    if (pathname.startsWith('/account')) return BADGES['/account'];
    const match =
      MAIN_LINKS.find((link) => isActivePath(link.href)) ?? DATA_LINKS.find((link) => isActivePath(link.href));
    if (!match) return 'dashboard';
    return BADGES[match.href];
  });

  function isActivePath(href: string): boolean {
    if (href === '/') return pathname === '/';
    return pathname === href || pathname.startsWith(`${href}/`);
  }

  function isActive(href: string): boolean {
    return isActivePath(href);
  }
</script>

<svelte:head>
  <link rel="icon" href={favicon} />
  <title>Big Fly</title>
</svelte:head>

<div class="grid h-dvh min-h-dvh grid-rows-[3.5rem_auto_minmax(0,1fr)_auto] overflow-x-hidden bg-mantle">
  <header class="sticky top-0 z-50 flex h-14 items-center gap-4 border-b border-outline bg-crust px-4 sm:px-6 lg:px-8">
    <a href={resolve('/')} class="font-display text-[1.1rem] font-bold text-foreground no-underline">Big Fly</a>
    <span class="rounded bg-outline px-2 py-0.5 font-mono text-[0.7rem] text-muted">
      {badge}
    </span>
    {#if meta.data}
      <span class="font-mono text-xxs text-muted opacity-60">v{meta.version}</span>
    {/if}
    <nav class="ml-auto flex items-center gap-1">
      {#each MAIN_LINKS as { href, label } (href)}
        <a
          href={resolve(href)}
          class="rounded px-2.5 py-1 text-[0.8rem] no-underline transition-colors duration-150 {isActive(href)
            ? 'bg-outline text-foreground'
            : 'text-muted hover:bg-outline hover:text-foreground'}">
          {label}
        </a>
      {/each}

      <DataNavigationMenu label="Dashboard" items={DATA_MENU_ITEMS} {pathname} />

      <div class="border-l border-outline px-2">
        <a
          href={resolve(API_DOCS_ROUTE)}
          target="_blank"
          rel="noreferrer"
          class="rounded px-2.5 py-1 text-[0.8rem] text-muted no-underline transition-colors duration-150 hover:bg-outline hover:text-foreground">
          API
          <span class="ml-1 inline-flex items-center gap-0.5 text-xxs">
            <i class="i-tabler-external-link"></i>
          </span>
        </a>
        <a
          href={resolve('/account')}
          class="rounded px-2.5 py-1 text-[0.8rem] no-underline transition-colors duration-150 {pathname.startsWith(
            '/account'
          )
            ? 'bg-outline text-foreground'
            : 'text-muted hover:bg-outline hover:text-foreground'}">
          Account
          <span class="ml-1 inline-flex items-center gap-0.5 text-xxs">
            <i class="i-tabler-user"></i>
          </span>
        </a>
      </div>
    </nav>
  </header>

  <div
    class="flex items-center justify-center gap-4 border-b border-white bg-rose-500 px-4 py-2 font-mono text-sm text-white sm:px-6 lg:px-8">
    <span class="inline-flex items-center gap-1.5">
      <i class="i-tabler-alert-triangle"></i>
      <strong>Preview</strong>
    </span>
    <div class="flex flex-col">
      <span>
        Big Fly is in early alpha for the next few weeks. Expect bugs, breaking changes, and incomplete features.
      </span>
      <span>We appreciate your patience and feedback as we work towards a stable release in the coming weeks! </span>
    </div>
  </div>

  <main class="min-h-0 overflow-x-hidden overflow-y-auto">
    {@render children()}
  </main>

  <AppFooter />
</div>
