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

<div class="flex items-center gap-3 text-[0.8rem] text-muted">
  <span class="font-mono text-[0.72rem]">{total.toLocaleString()} records</span>

  <div class="ml-auto flex items-center gap-1">
    <button
      disabled={page <= 1}
      onclick={() => setPage(page - 1)}
      class="rounded border border-outline px-2.5 py-1 transition-colors hover:border-primary hover:text-foreground disabled:cursor-not-allowed disabled:opacity-30">
      ←
    </button>
    <span class="min-w-20 px-2 text-center font-mono text-[0.72rem]">{page} / {totalPages}</span>
    <button
      disabled={page >= totalPages}
      onclick={() => setPage(page + 1)}
      class="rounded border border-outline px-2.5 py-1 transition-colors hover:border-primary hover:text-foreground disabled:cursor-not-allowed disabled:opacity-30">
      →
    </button>
  </div>

  <select
    value={perPage}
    onchange={handlePerPageChange}
    class="rounded border border-outline bg-crust px-2 py-1 font-mono text-[0.72rem] text-muted">
    {#each [10, 20, 50, 100] as n (n)}
      <option value={n}>{n} / page</option>
    {/each}
  </select>
</div>
