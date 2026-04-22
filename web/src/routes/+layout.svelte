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
    { href: '/docs', label: 'Docs' }
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
  type AppPath = (typeof MAIN_LINKS)[number]['href'] | (typeof DATA_LINKS)[number]['href'];

  const BADGES: Record<AppPath, Uppercase<string>> = {
    '/': 'HOME',
    '/players': 'PLAYERS',
    '/teams': 'TEAMS',
    '/games': 'GAMES',
    '/seasons': 'SEASONS',
    '/leaders': 'LEADERS',
    '/data': 'SOURCES',
    '/docs': 'DOCS'
  };

  let pathname = $derived(page.url.pathname);
  let badge = $derived.by<Uppercase<string>>(() => {
    const match =
      MAIN_LINKS.find((link) => isActivePath(link.href)) ?? DATA_LINKS.find((link) => isActivePath(link.href));
    if (!match) return 'DASHBOARD';
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

<div class="grid h-dvh min-h-dvh grid-rows-[auto_auto_minmax(0,1fr)_auto] overflow-x-hidden bg-mantle">
  <header
    class="sticky top-0 z-50 flex flex-wrap items-center gap-3 border-b border-outline bg-crust px-3 py-2 sm:px-6 lg:px-8">
    <div class="flex min-w-0 items-center gap-2.5">
      <a href={resolve('/')} class="font-display text-[1.1rem] font-bold text-foreground no-underline">
        Big
        <span class="text-primary">Fly</span>
      </a>
      <span class="hidden rounded bg-outline px-2 py-0.5 font-mono text-[0.7rem] text-muted sm:inline-flex">
        {badge}
      </span>
      {#if meta.data}
        <span class="hidden font-mono text-xxs text-muted opacity-60 md:inline">v{meta.version}</span>
      {/if}
    </div>
    <nav class="mt-1 flex w-full min-w-0 flex-wrap items-center gap-1 sm:mt-0 sm:ml-auto sm:w-auto sm:flex-nowrap">
      {#each MAIN_LINKS as { href, label } (href)}
        <a
          href={resolve(href)}
          class="rounded px-2 py-1 text-[0.75rem] whitespace-nowrap no-underline transition-colors duration-150 sm:px-2.5 sm:text-[0.8rem] {isActive(
            href
          )
            ? 'bg-outline text-foreground'
            : 'text-muted hover:bg-outline hover:text-foreground'}">
          {label}
        </a>
      {/each}

      <DataNavigationMenu label="Explore" items={DATA_MENU_ITEMS} {pathname} />

      <div class="flex items-center gap-1 border-l border-outline pl-2">
        <a
          href={resolve(API_DOCS_ROUTE)}
          target="_blank"
          rel="noreferrer"
          class="rounded px-2 py-1 text-[0.75rem] whitespace-nowrap text-muted no-underline transition-colors duration-150 hover:bg-outline hover:text-foreground sm:px-2.5 sm:text-[0.8rem]">
          API
          <span class="ml-1 inline-flex items-center gap-0.5 text-xxs">
            <i class="i-tabler-external-link"></i>
          </span>
        </a>
      </div>
    </nav>
  </header>

  <div
    class="flex flex-col items-start gap-2 border-b border-white bg-rose-500 px-3 py-2 font-mono text-xs text-white sm:flex-row sm:items-center sm:justify-center sm:gap-4 sm:px-6 sm:text-sm lg:px-8">
    <span class="inline-flex items-center gap-1.5">
      <span class="flex items-center">
        <i class="i-tabler-alert-triangle"></i>
      </span>
      <strong>Preview</strong>
    </span>
    <div class="flex max-w-5xl flex-col">
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
