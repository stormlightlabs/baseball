<script lang="ts">
  let {
    page,
    perPage,
    total,
    onPageChange,
    onPerPageChange
  }: {
    page: number;
    perPage: number;
    total: number;
    onPageChange?: (page: number) => void;
    onPerPageChange?: (perPage: number) => void;
  } = $props();

  let totalPages = $derived(Math.max(1, Math.ceil(total / perPage)));

  function setPage(next: number): void {
    const clamped = Math.min(Math.max(1, next), totalPages);
    if (clamped === page) return;
    onPageChange?.(clamped);
  }

  function handlePerPageChange(event: Event): void {
    const selected = Number((event.target as HTMLSelectElement).value);
    if (selected === perPage) return;
    onPerPageChange?.(selected);
  }
</script>

<div class="flex items-center gap-2 text-[0.8rem] text-muted">
  <span class="font-mono text-[0.68rem] text-muted">{total.toLocaleString()} results</span>

  <div class="ml-auto flex items-center gap-1">
    <button
      disabled={page <= 1}
      onclick={() => setPage(page - 1)}
      aria-label="Previous page"
      class="flex items-center rounded border border-outline p-1.5 transition-colors hover:border-primary hover:text-foreground disabled:cursor-not-allowed disabled:opacity-30">
      <span class="flex items-center"><i class="i-tabler-chevron-left size-3.5"></i></span>
    </button>
    <span class="min-w-16 px-1.5 text-center font-mono text-[0.68rem]">{page} / {totalPages}</span>
    <button
      disabled={page >= totalPages}
      onclick={() => setPage(page + 1)}
      aria-label="Next page"
      class="flex items-center rounded border border-outline p-1.5 transition-colors hover:border-primary hover:text-foreground disabled:cursor-not-allowed disabled:opacity-30">
      <span class="flex items-center"><i class="i-tabler-chevron-right size-3.5"></i></span>
    </button>
  </div>

  {#if onPerPageChange}
    <select
      value={perPage}
      onchange={handlePerPageChange}
      class="rounded border border-outline bg-crust px-2 py-1 font-mono text-[0.68rem] text-muted">
      {#each [10, 20, 50, 100] as n (n)}
        <option value={n}>{n} / page</option>
      {/each}
    </select>
  {/if}
</div>
