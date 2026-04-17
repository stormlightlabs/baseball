<script lang="ts">
  let { tabs, active = $bindable() }: { tabs: Array<{ id: string; label: string }>; active: string } = $props();

  function handleKeydown(e: KeyboardEvent, id: string) {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      active = id;
    }

    if (e.key === 'ArrowRight') {
      const idx = tabs.findIndex((t) => t.id === active);
      const next = tabs[(idx + 1) % tabs.length];
      active = next.id;
    }

    if (e.key === 'ArrowLeft') {
      const idx = tabs.findIndex((t) => t.id === active);
      const prev = tabs[(idx - 1 + tabs.length) % tabs.length];
      active = prev.id;
    }
  }
</script>

<div class="mb-4 flex gap-1" role="tablist">
  {#each tabs as tab (`${tab.id}:${tab.label}`)}
    <button
      role="tab"
      aria-selected={active === tab.id}
      class="rounded border px-3 py-1.25 font-display text-[0.78rem] transition-all duration-150 {active === tab.id
        ? 'border-outline bg-surface text-foreground'
        : 'border-transparent text-muted hover:text-foreground'}"
      onclick={() => (active = tab.id)}
      onkeydown={(e) => handleKeydown(e, tab.id)}>
      {tab.label}
    </button>
  {/each}
</div>
