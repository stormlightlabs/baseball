<script lang="ts">
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import favicon from '$lib/assets/favicon.svg';
  import { meta } from '$lib/meta.svelte.js';
  import '@fontsource-variable/google-sans';
  import '@fontsource-variable/google-sans-code';
  import '@fontsource-variable/inter';
  import { onMount } from 'svelte';
  import './layout.css';

  let { children } = $props();

  onMount(() => meta.init());

  type AppPath = (typeof LINKS)[number]['href'] | '/account';

  const LINKS = [
    { href: '/', label: 'Home' },
    { href: '/players', label: 'Players' },
    { href: '/teams', label: 'Teams' },
    { href: '/games', label: 'Games' },
    { href: '/seasons', label: 'Seasons' },
    { href: '/leaders', label: 'Leaders' },
    { href: '/compare', label: 'Compare' },
    { href: '/data', label: 'Data' },
    { href: '/docs', label: 'About' }
  ] as const;

  const API_DOCS_ROUTE = '/explorer' as const;

  const BADGES: Record<AppPath, string> = {
    '/': 'home',
    '/players': 'players',
    '/teams': 'teams',
    '/games': 'games',
    '/seasons': 'seasons',
    '/leaders': 'leaders',
    '/compare': 'compare',
    '/data': 'data sources',
    '/docs': 'about',
    '/account': 'account'
  };

  let pathname = $derived(page.url.pathname as AppPath);
  let badge = $derived(BADGES[pathname] ?? 'dashboard');

  function isActive(href: string): boolean {
    if (href === '/') return pathname === '/';
    return pathname.startsWith(href);
  }
</script>

<svelte:head>
  <link rel="icon" href={favicon} />
  <title>Big Fly</title>
</svelte:head>

<div class="grid h-dvh min-h-dvh grid-rows-[3.5rem_auto_minmax(0,1fr)] overflow-x-hidden bg-mantle">
  <header class="sticky top-0 z-50 flex h-14 items-center gap-4 border-b border-outline bg-crust px-4 sm:px-6 lg:px-8">
    <a href={resolve('/')} class="font-display text-[1.1rem] font-bold text-foreground no-underline">Big Fly</a>
    <span class="rounded bg-outline px-2 py-0.5 font-mono text-[0.7rem] text-muted">
      {badge}
    </span>
    {#if meta.data}
      <!-- <span class="font-mono text-[0.65rem] text-muted opacity-60">v{meta.version}</span> -->
      <span class="font-mono text-[0.65rem] text-muted opacity-60">ALPHA</span>
    {/if}
    <nav class="ml-auto flex items-center gap-1">
      {#each LINKS as { href, label } (href)}
        <a
          href={resolve(href)}
          class="rounded px-2.5 py-1 text-[0.8rem] no-underline transition-colors duration-150 {isActive(href)
            ? 'bg-outline text-foreground'
            : 'text-muted hover:bg-outline hover:text-foreground'}">
          {label}
        </a>
      {/each}
      <a
        href={resolve(API_DOCS_ROUTE)}
        target="_blank"
        rel="noreferrer"
        class="ml-2 rounded px-2.5 py-1 text-[0.8rem] text-muted no-underline transition-colors duration-150 hover:bg-outline hover:text-foreground">
        API docs
      </a>
      <a
        href={resolve('/account')}
        class="rounded px-2.5 py-1 text-[0.8rem] no-underline transition-colors duration-150 {pathname.startsWith(
          '/account'
        )
          ? 'bg-outline text-foreground'
          : 'text-muted hover:bg-outline hover:text-foreground'}">
        Account
      </a>
    </nav>
  </header>

  <div class="border-b border-white bg-rose-500 px-4 py-2 font-mono text-sm text-white sm:px-6 lg:px-8">
    Preview notice: Big Fly is in early alpha for the next few weeks. Expect bugs, breaking changes, and incomplete
    features. We appreciate your patience and feedback as we work towards a stable release in the coming weeks!
  </div>

  <main class="min-h-0">
    {@render children()}
  </main>
</div>
