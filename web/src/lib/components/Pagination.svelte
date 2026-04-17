<script lang="ts">
  let {
    page = $bindable(1),
    perPage = $bindable(20),
    total
  }: { page?: number; perPage?: number; total: number } = $props();

  let totalPages = $derived(Math.max(1, Math.ceil(total / perPage)));
</script>

<div class="flex items-center gap-3 text-[0.8rem] text-muted">
  <span class="font-monospace text-[0.72rem]">{total.toLocaleString()} records</span>

  <div class="ml-auto flex items-center gap-1">
    <button
      disabled={page <= 1}
      onclick={() => page--}
      class="rounded border border-outline px-2.5 py-1 transition-colors hover:border-primary hover:text-foreground disabled:cursor-not-allowed disabled:opacity-30">
      ←
    </button>
    <span class="min-w-20 px-2 text-center font-monospace text-[0.72rem]">{page} / {totalPages}</span>
    <button
      disabled={page >= totalPages}
      onclick={() => page++}
      class="rounded border border-outline px-2.5 py-1 transition-colors hover:border-primary hover:text-foreground disabled:cursor-not-allowed disabled:opacity-30">
      →
    </button>
  </div>

  <select
    bind:value={perPage}
    class="rounded border border-outline bg-crust px-2 py-1 font-monospace text-[0.72rem] text-muted">
    {#each [10, 20, 50, 100] as n (n)}
      <option value={n}>{n} / page</option>
    {/each}
  </select>
</div>
