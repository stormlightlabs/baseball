<script lang="ts">
  /**
   * DatePicker — a custom calendar-based date input.
   *
   * Props:
   *   value   — bound ISO date string "YYYY-MM-DD" (or empty string)
   *   id      — forwarded to the hidden input (for label association)
   *   label   — optional visible label rendered inside the component
   *   min     — minimum selectable date "YYYY-MM-DD"
   *   max     — maximum selectable date "YYYY-MM-DD"
   *   placeholder — text shown when no date selected
   */

  type Props = { value?: string; id?: string; min?: string; max?: string; placeholder?: string };

  let { value = $bindable(''), id = '', min = '', max = '', placeholder = 'Select date' }: Props = $props();

  const DAYS = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
  const MONTHS = [
    'January',
    'February',
    'March',
    'April',
    'May',
    'June',
    'July',
    'August',
    'September',
    'October',
    'November',
    'December'
  ];

  function parseIso(iso: string): { year: number; month: number; day: number } | null {
    if (!iso) return null;
    const [y, m, d] = iso.split('-').map(Number);
    if (!y || !m || !d) return null;
    return { year: y, month: m - 1, day: d };
  }

  function toIso(year: number, month: number, day: number): string {
    return `${year}-${String(month + 1).padStart(2, '0')}-${String(day).padStart(2, '0')}`;
  }

  function formatDisplay(iso: string): string {
    const p = parseIso(iso);
    if (!p) return '';
    return `${MONTHS[p.month]} ${p.day}, ${p.year}`;
  }

  const _initial = parseIso(value);
  let viewYear = $state<number>(_initial?.year ?? new Date().getFullYear());
  let viewMonth = $state<number>(_initial?.month ?? new Date().getMonth());

  let open = $state(false);

  $effect(() => {
    const p = parseIso(value);
    if (p) {
      viewYear = p.year;
      viewMonth = p.month;
    }
  });

  const daysInMonth = (year: number, month: number): number => new Date(year, month + 1, 0).getDate();

  const firstDayOfMonth = (year: number, month: number): number => new Date(year, month, 1).getDay();

  let grid = $derived.by(() => {
    const count = daysInMonth(viewYear, viewMonth);
    const start = firstDayOfMonth(viewYear, viewMonth);
    const cells: (number | null)[] = [];

    for (let i = 0; i < start; i++) {
      cells.push(null);
    }

    for (let d = 1; d <= count; d++) {
      cells.push(d);
    }

    while (cells.length % 7 !== 0) {
      cells.push(null);
    }

    return cells;
  });

  function isDisabled(day: number): boolean {
    const iso = toIso(viewYear, viewMonth, day);
    if (min && iso < min) return true;
    if (max && iso > max) return true;
    return false;
  }

  function selectDay(day: number) {
    if (isDisabled(day)) return;
    value = toIso(viewYear, viewMonth, day);
    open = false;
  }

  function prevMonth() {
    if (viewMonth === 0) {
      viewMonth = 11;
      viewYear--;
    } else viewMonth--;
  }

  function nextMonth() {
    if (viewMonth === 11) {
      viewMonth = 0;
      viewYear++;
    } else viewMonth++;
  }

  function clear(e: MouseEvent) {
    e.stopPropagation();
    value = '';
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') open = false;
  }

  function handleOutside(node: HTMLElement) {
    function onClick(e: MouseEvent) {
      if (!node.contains(e.target as Node)) open = false;
    }
    document.addEventListener('click', onClick, true);
    return {
      destroy() {
        document.removeEventListener('click', onClick, true);
      }
    };
  }
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="relative" use:handleOutside onkeydown={handleKeydown}>
  <input type="hidden" {id} {value} />

  <button
    type="button"
    class="flex w-full items-center justify-between gap-2 rounded border border-outline bg-surface px-2.5 py-1.5 font-mono text-xs transition-colors hover:border-primary/50 focus:border-primary focus:outline-none {open
      ? 'border-primary'
      : ''}"
    onclick={() => (open = !open)}
    aria-haspopup="true"
    aria-expanded={open}>
    <span class={value ? 'text-foreground' : 'text-muted'}>
      {value ? formatDisplay(value) : placeholder}
    </span>
    <span class="flex shrink-0 items-center gap-1">
      {#if value}
        <span
          role="button"
          tabindex="0"
          class="flex items-center text-muted hover:text-foreground"
          aria-label="Clear date"
          onclick={clear}
          onkeydown={(e) => e.key === 'Enter' && clear(e as unknown as MouseEvent)}>
          <i class="i-tabler-x size-3"></i>
        </span>
      {/if}
      <span class="flex items-center text-muted">
        <i class="i-tabler-calendar size-3.5"></i>
      </span>
    </span>
  </button>

  {#if open}
    <div
      class="absolute top-full left-0 z-50 mt-1 w-64 rounded-lg border border-outline bg-crust shadow-xl shadow-black/40"
      role="dialog"
      aria-label="Date picker calendar">
      <div class="flex items-center justify-between border-b border-outline px-3 py-2">
        <button
          type="button"
          class="flex items-center rounded p-1 text-muted transition-colors hover:bg-surface hover:text-foreground"
          onclick={prevMonth}
          aria-label="Previous month">
          <span class="flex items-center"><i class="i-tabler-chevron-left size-3.5"></i></span>
        </button>

        <span class="font-display text-[0.8rem] text-foreground">
          {MONTHS[viewMonth]}
          {viewYear}
        </span>

        <button
          type="button"
          class="flex items-center rounded p-1 text-muted transition-colors hover:bg-surface hover:text-foreground"
          onclick={nextMonth}
          aria-label="Next month">
          <span class="flex items-center"><i class="i-tabler-chevron-right size-3.5"></i></span>
        </button>
      </div>

      <div class="grid grid-cols-7 px-2 pt-2">
        {#each DAYS as dayLabel (dayLabel)}
          <div class="pb-1 text-center font-mono text-[0.6rem] tracking-wider text-muted uppercase">
            {dayLabel[0]}
          </div>
        {/each}
      </div>

      <div class="grid grid-cols-7 gap-y-0.5 px-2 pb-2">
        {#each grid as cell, i (i)}
          {#if cell === null}
            <div></div>
          {:else}
            {@const iso = toIso(viewYear, viewMonth, cell)}
            {@const isSelected = value === iso}
            {@const isToday = iso === toIso(new Date().getFullYear(), new Date().getMonth(), new Date().getDate())}
            {@const disabled = isDisabled(cell)}
            <button
              type="button"
              class="mx-auto flex size-7 items-center justify-center rounded font-mono text-[0.72rem] transition-colors
                {disabled ? 'cursor-not-allowed text-muted/30' : 'cursor-pointer hover:bg-surface'}
                {isSelected ? 'bg-primary text-white hover:bg-primary/90' : ''}
                {isToday && !isSelected ? 'text-primary' : ''}
                {!isSelected && !disabled && !isToday ? 'text-foreground' : ''}"
              onclick={() => selectDay(cell)}
              {disabled}
              aria-label={formatDisplay(iso)}
              aria-pressed={isSelected}>
              {cell}
            </button>
          {/if}
        {/each}
      </div>

      {#if value}
        <div class="border-t border-outline px-3 py-1.5">
          <button
            type="button"
            class="font-mono text-xxs text-muted transition-colors hover:text-foreground"
            onclick={() => {
              value = '';
              open = false;
            }}>
            Clear date
          </button>
        </div>
      {/if}
    </div>
  {/if}
</div>
