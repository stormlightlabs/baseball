<script lang="ts">
  import { resolve } from '$app/paths';

  type NavItem = { href: string; label: string; description: string };

  let { label = 'Data', items, pathname }: { label?: string; items: NavItem[]; pathname: string } = $props();

  let expanded = $state(false);

  let active = $derived(
    items.some((item) => {
      if (pathname === item.href) return true;
      return pathname.startsWith(`${item.href}/`);
    })
  );

  function closeOnOutsideClick(node: HTMLElement): { destroy: () => void } {
    const onDocumentClick = (event: MouseEvent): void => {
      const target = event.target as Node | null;
      if (!target) return;
      if (node.contains(target)) return;
      expanded = false;
    };

    document.addEventListener('click', onDocumentClick);
    return {
      destroy() {
        document.removeEventListener('click', onDocumentClick);
      }
    };
  }

  function onKeyDown(event: KeyboardEvent): void {
    if (event.key !== 'Escape') return;
    expanded = false;
  }

  function toggleMenu(): void {
    expanded = !expanded;
  }

  function closeMenu(): void {
    expanded = false;
  }
</script>

<svelte:window onkeydown={onKeyDown} />

<div class="relative" use:closeOnOutsideClick>
  <button
    type="button"
    onclick={toggleMenu}
    aria-expanded={expanded}
    class="inline-flex items-center gap-1 rounded px-2 py-1 text-[0.75rem] transition-colors duration-150 sm:px-2.5 sm:text-[0.8rem] {active
      ? 'bg-outline text-foreground'
      : 'text-muted hover:bg-outline hover:text-foreground'}">
    <span>{label}</span>
    <span class="flex items-center">
      <i class="i-tabler-chevron-down text-[0.78rem] transition-transform duration-150 {expanded ? 'rotate-180' : ''}"
      ></i>
    </span>
  </button>

  {#if expanded}
    <div
      class="absolute right-0 z-30 mt-2 w-[min(18rem,calc(100vw-1rem))] rounded-lg border border-outline bg-crust p-2 shadow-xl">
      <div class="px-2 pb-1 font-mono text-[0.62rem] tracking-wide text-muted uppercase">Sections</div>
      <div class="space-y-1">
        {#each items as item (item.href)}
          {@const itemActive = pathname === item.href || pathname.startsWith(`${item.href}/`)}
          <a
            href={resolve(item.href as '/')}
            onclick={closeMenu}
            class="block rounded-md px-2 py-2 no-underline transition-colors {itemActive
              ? 'bg-outline'
              : 'hover:bg-surface'}">
            <div class="font-display text-[0.82rem] text-foreground">{item.label}</div>
            <div class="font-mono text-[0.64rem] text-muted">{item.description}</div>
          </a>
        {/each}
      </div>
    </div>
  {/if}
</div>
