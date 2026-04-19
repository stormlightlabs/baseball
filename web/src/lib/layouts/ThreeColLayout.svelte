<script lang="ts">
  import { browser } from '$app/environment';
  import type { Snippet } from 'svelte';
  import { onMount } from 'svelte';

  type LayoutPrefs = { sidebarCollapsed?: boolean; panelCollapsed?: boolean };

  const LAYOUT_PREFS_KEY = 'bigfly:three-col-layout:v1';

  let { sidebar, center, panel }: { sidebar: Snippet; center: Snippet; panel: Snippet } = $props();
  let sidebarCollapsed = $state(false);
  let panelCollapsed = $state(false);

  let desktopGridTemplate = $derived.by(() => {
    const sidebarWidth = sidebarCollapsed ? '2.75rem' : '280px';
    const panelWidth = panelCollapsed ? '2.75rem' : '320px';
    return `${sidebarWidth} minmax(0,1fr) ${panelWidth}`;
  });

  onMount(() => {
    if (!browser) return;

    const raw = localStorage.getItem(LAYOUT_PREFS_KEY);
    if (!raw) return;

    try {
      const parsed = JSON.parse(raw) as LayoutPrefs;
      sidebarCollapsed = parsed.sidebarCollapsed === true;
      panelCollapsed = parsed.panelCollapsed === true;
    } catch {
      console.debug('Failed to parse saved layout preferences, using defaults.');
    }
  });

  $effect(() => {
    if (!browser) return;
    const prefs: LayoutPrefs = { sidebarCollapsed, panelCollapsed };
    localStorage.setItem(LAYOUT_PREFS_KEY, JSON.stringify(prefs));
  });

  function toggleSidebar(): void {
    sidebarCollapsed = !sidebarCollapsed;
  }

  function togglePanel(): void {
    panelCollapsed = !panelCollapsed;
  }
</script>

<div
  style={`--three-col-template: ${desktopGridTemplate};`}
  class="grid min-h-full w-full grid-cols-1 bg-mantle md:grid-cols-2 xl:h-full xl:max-h-full xl:min-h-0 xl:grid-cols-(--three-col-template) xl:overflow-hidden xl:transition-[grid-template-columns] xl:duration-300 xl:ease-out">
  <aside
    id="app-layout-sidebar"
    class="order-1 min-w-0 border-t border-outline md:border-r xl:order-1 xl:flex xl:h-full xl:min-h-0 xl:flex-col xl:border-t-0">
    <div
      class="relative flex items-center justify-between border-b border-outline px-2.5 py-2 sm:px-3 {sidebarCollapsed
        ? 'xl:justify-center'
        : ''}">
      <span
        class="font-mono text-[0.62rem] tracking-[0.08em] text-muted uppercase {sidebarCollapsed ? 'xl:hidden' : ''}">
        Filters
      </span>
      <button
        type="button"
        aria-controls="app-layout-sidebar-content"
        aria-expanded={!sidebarCollapsed}
        aria-label={sidebarCollapsed ? 'Show filters panel' : 'Hide filters panel'}
        onclick={toggleSidebar}
        class="inline-flex items-center gap-1 rounded border border-outline px-2 py-1 text-muted transition-colors hover:border-primary/40 hover:text-foreground">
        <span class="flex items-center">
          <i class={sidebarCollapsed ? 'i-tabler-chevron-right' : 'i-tabler-chevron-left'}></i>
        </span>
        <span class="font-mono text-[0.62rem] tracking-[0.08em] uppercase xl:hidden">
          {sidebarCollapsed ? 'Show' : 'Hide'}
        </span>
      </button>
    </div>
    <div
      id="app-layout-sidebar-content"
      class="transition-[max-height,opacity,padding] duration-300 ease-out {sidebarCollapsed
        ? 'max-h-0 overflow-hidden p-0 opacity-0 xl:pointer-events-none xl:max-h-none xl:p-0 xl:opacity-0'
        : 'max-h-800 p-4 opacity-100 sm:p-5 xl:min-h-0 xl:flex-1 xl:overflow-y-auto xl:overscroll-y-contain'}">
      {@render sidebar()}
    </div>
  </aside>
  <aside
    id="app-layout-panel"
    class="order-2 min-w-0 border-t border-outline xl:order-3 xl:flex xl:h-full xl:min-h-0 xl:flex-col xl:border-t-0 xl:border-l">
    <div
      class="relative flex items-center justify-between border-b border-outline px-2.5 py-2 sm:px-3 {panelCollapsed
        ? 'xl:justify-center'
        : ''}">
      <span class="font-mono text-[0.62rem] tracking-[0.08em] text-muted uppercase {panelCollapsed ? 'xl:hidden' : ''}">
        API
      </span>
      <button
        type="button"
        aria-controls="app-layout-panel-content"
        aria-expanded={!panelCollapsed}
        aria-label={panelCollapsed ? 'Show API panel' : 'Hide API panel'}
        onclick={togglePanel}
        class="inline-flex items-center gap-1 rounded border border-outline px-2 py-1 text-muted transition-colors hover:border-primary/40 hover:text-foreground">
        <span class="flex items-center">
          <i class={panelCollapsed ? 'i-tabler-chevron-left' : 'i-tabler-chevron-right'}></i>
        </span>
        <span class="font-mono text-[0.62rem] tracking-[0.08em] uppercase xl:hidden">
          {panelCollapsed ? 'Show' : 'Hide'}
        </span>
      </button>
    </div>
    <div
      id="app-layout-panel-content"
      class="transition-[max-height,opacity,padding] duration-300 ease-out {panelCollapsed
        ? 'max-h-0 overflow-hidden p-0 opacity-0 xl:pointer-events-none xl:max-h-none xl:p-0 xl:opacity-0'
        : 'max-h-800 p-4 opacity-100 sm:p-5 xl:min-h-0 xl:flex-1 xl:overflow-y-auto xl:overscroll-y-contain'}">
      {@render panel()}
    </div>
  </aside>
  <main
    class="order-3 min-w-0 p-4 sm:p-6 md:col-span-2 xl:order-2 xl:col-span-1 xl:h-full xl:min-h-0 xl:overflow-y-auto xl:overscroll-y-contain">
    {@render center()}
  </main>
</div>
