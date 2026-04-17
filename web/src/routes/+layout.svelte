<script lang="ts">
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import favicon from '$lib/assets/favicon.svg';
  import '@fontsource-variable/google-sans';
  import '@fontsource-variable/google-sans-code';
  import '@fontsource-variable/inter';
  import './layout.css';

  let { children } = $props();

  type AppPath = (typeof LINKS)[number]['href'] | '/account';

  const LINKS = [
    { href: '/', label: 'Home' },
    { href: '/players', label: 'Players' },
    { href: '/teams', label: 'Teams' },
    { href: '/games', label: 'Games' },
    { href: '/seasons', label: 'Seasons' },
    { href: '/leaders', label: 'Leaders' },
    { href: '/compare', label: 'Compare' },
    { href: '/explorer', label: 'API' },
    { href: '/data', label: 'Data' }
  ] as const;

  const BADGES: Record<AppPath, string> = {
    '/': 'home',
    '/players': 'players',
    '/teams': 'teams',
    '/games': 'games',
    '/seasons': 'seasons',
    '/leaders': 'leaders',
    '/compare': 'compare',
    '/explorer': 'api explorer',
    '/data': 'data sources',
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
</svelte:head>

<header class="sticky top-0 z-50 flex h-14 items-center gap-4 border-b border-outline bg-crust px-8">
  <a href={resolve('/')} class="font-display text-[1.1rem] font-bold text-foreground no-underline"> Baseball API </a>
  <span class="rounded bg-outline px-2 py-0.5 font-monospace text-[0.7rem] text-muted">
    {badge}
  </span>
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
      href={resolve('/account')}
      class="ml-2 rounded px-2.5 py-1 text-[0.8rem] no-underline transition-colors duration-150 {pathname.startsWith(
        '/account'
      )
        ? 'bg-outline text-foreground'
        : 'text-muted hover:bg-outline hover:text-foreground'}">
      Account
    </a>
  </nav>
</header>

{@render children()}
