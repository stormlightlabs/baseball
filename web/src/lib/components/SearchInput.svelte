<script lang="ts">
  let {
    value = $bindable(''),
    placeholder = 'Search…',
    mini = false,
    onsubmit
  }: { value?: string; placeholder?: string; mini?: boolean; onsubmit?: (value: string) => void } = $props();

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter') onsubmit?.(value);
  }
</script>

{#if mini}
  <div class="flex overflow-hidden rounded-md border border-outline bg-surface">
    <input
      type="text"
      bind:value
      {placeholder}
      onkeydown={handleKeydown}
      class="flex-1 bg-transparent px-3 py-2 text-[0.8rem] text-foreground outline-none placeholder:text-muted" />
    {#if onsubmit}
      <button
        onclick={() => onsubmit?.(value)}
        class="border-l border-outline bg-outline px-3 text-muted transition-colors hover:text-foreground">
        →
      </button>
    {/if}
  </div>
{:else}
  <div class="flex overflow-hidden rounded-lg border border-outline bg-crust">
    <input
      type="text"
      bind:value
      {placeholder}
      onkeydown={handleKeydown}
      class="flex-1 bg-transparent px-4 py-3 text-[0.9rem] text-foreground outline-none placeholder:text-muted" />
    {#if onsubmit}
      <button
        onclick={() => onsubmit?.(value)}
        class="bg-primary px-5 font-display text-[0.85rem] text-white transition-opacity duration-150 hover:opacity-85">
        Search
      </button>
    {/if}
  </div>
{/if}
