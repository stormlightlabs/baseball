<script lang="ts">
  import { resolve } from '$app/paths';

  let { tabs, active }: { tabs: Array<{ id: string; label: string; href: string }>; active: string } = $props();
</script>

<div class="mb-4 flex gap-1" role="tablist" aria-label="Player sections">
  {#each tabs as tab (`${tab.id}:${tab.label}`)}
    <a
      role="tab"
      aria-selected={active === tab.id}
      href={resolve(tab.href as '/')}
      class="rounded border px-3 py-1.25 font-display text-[0.78rem] no-underline transition-all duration-150 {active ===
      tab.id
        ? 'border-outline bg-surface text-foreground'
        : 'border-transparent text-muted hover:text-foreground'}">
      {tab.label}
    </a>
  {/each}
</div>
