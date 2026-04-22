<script lang="ts">
  import { resolve } from '$app/paths';
  import Tooltip from './Tooltip.svelte';

  type Row = Record<string, unknown>;

  type Column = {
    label: string;
    key: string;
    sortable?: boolean;
    format?: (value: unknown) => string;
    rank?: boolean;
    href?: (value: unknown, row: Row) => string | undefined;
    tooltip?: (value: unknown, row: Row) => string | undefined;
  };

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
      ? [...rows].toSorted((a, b) => {
          const activeCol = columns.find((col) => col.key === sort.key);
          const av = activeCol ? a[activeCol.key] : a[sort.key];
          const bv = activeCol ? b[activeCol.key] : b[sort.key];
          let cmp = 0;
          if (av == null && bv == null) cmp = 0;
          else if (av == null) cmp = -1;
          else if (bv == null) cmp = 1;
          else if (av < bv) cmp = -1;
          else if (av > bv) cmp = 1;
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

  function cellHref(col: Column, row: Row): string | undefined {
    const value = row[col.key];
    return col.href?.(value, row);
  }

  function cellTooltip(col: Column, row: Row): string | undefined {
    const value = row[col.key];
    return col.tooltip?.(value, row);
  }
</script>

<div class="overflow-x-auto">
  <table class="w-full border-collapse text-xs">
    <thead>
      <tr>
        {#each columns as col (col.label)}
          <th
            class="border-b border-outline px-2 py-1.5 text-left font-sans text-[0.72rem] font-medium whitespace-nowrap text-muted {col.sortable
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
            <td class="px-2 py-1.5 font-mono text-[0.72rem] text-foreground">
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
                {@const href = cellHref(col, row)}
                {@const tooltip = cellTooltip(col, row)}
                {#if href && tooltip}
                  <Tooltip text={tooltip}>
                    <a href={resolve(href as '/')} class="text-primary hover:underline">{display(col, row)}</a>
                  </Tooltip>
                {:else if href}
                  <a href={resolve(href as '/')} class="text-primary hover:underline">{display(col, row)}</a>
                {:else if tooltip}
                  <Tooltip text={tooltip}>
                    <span class="cursor-help">{display(col, row)}</span>
                  </Tooltip>
                {:else}
                  {display(col, row)}
                {/if}
              {/if}
            </td>
          {/each}
        </tr>
      {/each}
    </tbody>
  </table>
</div>
