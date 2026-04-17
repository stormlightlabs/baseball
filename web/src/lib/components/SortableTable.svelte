<script lang="ts">
  type Column = { label: string; key: string; sortable?: boolean; format?: (value: unknown) => string; rank?: boolean };

  type Row = Record<string, unknown>;

  let {
    columns,
    rows,
    sort = $bindable<{ key: string; dir: 'asc' | 'desc' }>({ key: '', dir: 'asc' })
  }: { columns: Column[]; rows: Row[]; sort?: { key: string; dir: 'asc' | 'desc' } } = $props();

  function toggleSort(col: Column) {
    if (!col.sortable) return;
    if (sort.key === col.key) {
      sort = { key: col.key, dir: sort.dir === 'asc' ? 'desc' : 'asc' };
    } else {
      sort = { key: col.key, dir: 'asc' };
    }
  }

  let sorted = $derived(
    sort.key
      ? [...rows].sort((a, b) => {
          const av = a[sort.key];
          const bv = b[sort.key];
          const cmp = av! < bv! ? -1 : av! > bv! ? 1 : 0;
          return sort.dir === 'asc' ? cmp : -cmp;
        })
      : rows
  );

  function maxVal(key: string): number {
    return Math.max(...rows.map((r) => Number(r[key]) || 0));
  }

  function display(col: Column, row: Row): string {
    const val = row[col.key];
    return col.format ? col.format(val) : String(val ?? '');
  }
</script>

<div class="overflow-x-auto">
  <table class="w-full border-collapse text-[0.75rem]">
    <thead>
      <tr>
        {#each columns as col (col.label)}
          <th
            class="border-b border-outline px-2 py-1.5 text-left font-body text-[0.72rem] font-medium whitespace-nowrap text-muted {col.sortable
              ? 'cursor-pointer hover:text-foreground'
              : ''}"
            onclick={() => toggleSort(col)}>
            {col.label}
            {#if col.sortable}
              {#if sort.key === col.key}
                {sort.dir === 'asc' ? ' ↑' : ' ↓'}
              {:else}
                <span class="opacity-30"> ↕</span>
              {/if}
            {/if}
          </th>
        {/each}
      </tr>
    </thead>
    <tbody>
      {#each sorted as row, i (i)}
        <tr class="border-b border-outline last:border-b-0 hover:[&>td]:bg-surface">
          {#each columns as col (col.label)}
            <td class="px-2 py-1.5 font-monospace text-[0.72rem] text-foreground">
              {#if col.rank}
                {@const val = Number(row[col.key]) || 0}
                {@const max = maxVal(col.key)}
                <div class="flex items-center gap-2">
                  <div class="h-1.5 w-16 overflow-hidden rounded-full bg-outline">
                    <div class="h-full rounded-full bg-primary" style="width: {max ? (val / max) * 100 : 0}%"></div>
                  </div>
                  <span>{display(col, row)}</span>
                </div>
              {:else}
                {display(col, row)}
              {/if}
            </td>
          {/each}
        </tr>
      {/each}
    </tbody>
  </table>
</div>
