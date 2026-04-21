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
  <div
    class="flex rounded-md border border-outline bg-surface transition-[box-shadow,border-color] focus-within:border-primary/60 focus-within:ring-1 focus-within:ring-primary/45 focus-within:ring-inset">
    <input
      type="text"
      bind:value
      {placeholder}
      onkeydown={handleKeydown}
      class="min-w-0 flex-1 rounded-l-md bg-transparent px-3 py-2 text-[0.8rem] text-foreground outline-none placeholder:text-muted" />
    {#if onsubmit}
      <button
        onclick={() => onsubmit?.(value)}
        class="rounded-r-md border-l border-outline bg-outline px-3 text-muted transition-colors hover:text-foreground">
        <span class="flex items-center">
          <i class="i-tabler-arrow-right"></i>
        </span>
        <span class="sr-only">Submit search</span>
      </button>
    {/if}
  </div>
{:else}
  <div
    class="flex rounded-lg border border-outline bg-crust transition-[box-shadow,border-color] focus-within:border-primary/70 focus-within:ring-1 focus-within:ring-primary/45 focus-within:ring-inset">
    <input
      type="text"
      bind:value
      {placeholder}
      onkeydown={handleKeydown}
      class="min-w-0 flex-1 rounded-l-lg bg-transparent px-4 py-3 text-[0.9rem] text-foreground outline-none placeholder:text-muted" />
    {#if onsubmit}
      <button
        onclick={() => onsubmit?.(value)}
        class="rounded-r-lg bg-primary px-5 font-display text-[0.85rem] text-white transition-opacity duration-150 hover:opacity-85">
        Search
      </button>
    {/if}
  </div>
{/if}
