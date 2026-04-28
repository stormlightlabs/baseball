<script lang="ts">
  import { KOFI_URL, PATREON_URL } from '$lib/common/constants';

  let {
    adClient = '',
    adSlot = '',
    variant = 'horizontal'
  }: { adClient?: string; adSlot?: string; variant?: 'horizontal' | 'vertical' } = $props();

  let baitEl = $state<HTMLDivElement | undefined>();
  let pushFailed = $state(false);

  const adBlocked = $derived.by(() => {
    if (!adClient || !adSlot) return true;
    if (!baitEl) return false;
    const style = globalThis.getComputedStyle(baitEl);
    return baitEl.offsetHeight === 0 || style.display === 'none' || style.visibility === 'hidden' || pushFailed;
  });

  $effect(() => {
    if (!baitEl || adBlocked || !adClient || !adSlot) return;
    try {
      ((globalThis as unknown as { adsbygoogle: unknown[] }).adsbygoogle =
        (globalThis as unknown as { adsbygoogle: unknown[] }).adsbygoogle || []).push({});
    } catch {
      pushFailed = true;
    }
  });
</script>

<!-- Bait element: adblockers hide/collapse ad-named elements, revealing their presence -->
<div
  bind:this={baitEl}
  class="adsense ads ad advert pointer-events-none absolute h-px w-px overflow-hidden opacity-0"
  aria-hidden="true">
</div>

{#if variant === 'vertical'}
  {#if adBlocked}
    <!-- Vertical fallback: compact Patreon CTA + Ko-fi link -->
    <div class="rounded-lg border border-outline bg-crust p-3">
      <div class="mb-2 flex items-center gap-1.5">
        <span class="flex items-center text-warning">
          <i class="i-tabler-heart-filled"></i>
        </span>
        <span class="font-mono text-xxs tracking-[0.08em] text-muted uppercase">Support Big Fly</span>
      </div>

      <a
        href={PATREON_URL}
        rel="external"
        target="_blank"
        class="mb-2 flex w-full items-center justify-center gap-2 rounded-md border border-[#FF424D]/40 bg-[#FF424D]/10 px-3 py-2 text-xs font-medium text-[#FF424D] no-underline transition-all duration-150 hover:bg-[#FF424D]/20 active:scale-[0.96]">
        <span class="flex items-center">
          <i class="i-tabler-brand-patreon"></i>
        </span>
        Become a Patron
      </a>

      <div class="flex items-center justify-center gap-1">
        <span class="text-xxs text-muted">or</span>
        <a
          href={KOFI_URL}
          rel="external"
          target="_blank"
          class="inline-flex items-center gap-1 text-xxs text-[#29ABE0] no-underline transition-colors duration-150 hover:text-[#29ABE0]/80">
          <span class="flex items-center">
            <i class="i-tabler-coffee"></i>
          </span>
          Ko-fi
        </a>
      </div>
    </div>
  {:else}
    <!-- Vertical ad slot (300×250) + Patreon link -->
    <div class="overflow-hidden rounded-lg border border-outline bg-crust">
      <div class="min-h-62.5 w-full">
        <ins
          class="adsbygoogle block"
          data-ad-client={adClient}
          data-ad-slot={adSlot}
          data-ad-format="rectangle"
          data-full-width-responsive="false"></ins>
      </div>

      <div class="flex items-center justify-end gap-1 border-t border-outline/50 px-3 py-1.5">
        <span class="text-xxs text-muted">Support us on</span>
        <a
          href={PATREON_URL}
          rel="external"
          target="_blank"
          class="inline-flex items-center gap-1 text-xxs text-[#FF424D] no-underline transition-colors duration-150 hover:text-[#FF424D]/80">
          <span class="flex items-center">
            <i class="i-tabler-brand-patreon"></i>
          </span>
          Patreon
        </a>
      </div>
    </div>
  {/if}
{:else if adBlocked}
  <!-- Horizontal fallback: Patreon as primary CTA, Ko-fi secondary -->
  <div class="rounded-lg border border-outline bg-crust p-4">
    <div class="mb-2.5 flex items-center gap-1.5">
      <span class="flex animate-pulse items-center text-rose-400">
        <i class="i-tabler-heart-filled"></i>
      </span>
      <span class="font-mono text-xxs tracking-[0.08em] text-muted uppercase">Support Big Fly</span>
    </div>

    <p class="mb-3 text-sm text-foreground/70" style="text-wrap: pretty">
      Enjoying the app? Help keep it free and ad-supported by backing us on Patreon.
    </p>

    <a
      href={PATREON_URL}
      rel="external"
      target="_blank"
      class="mb-3 flex w-full items-center justify-center gap-2 rounded-md border border-[#FF424D]/40 bg-[#FF424D]/10 px-4 py-2.5 text-sm font-medium text-[#FF424D] no-underline transition-all duration-150 hover:bg-[#FF424D]/20 active:scale-[0.96]">
      <span class="flex items-center">
        <i class="i-tabler-brand-patreon"></i>
      </span>
      Become a Patron
    </a>

    <div class="flex items-center justify-center gap-1">
      <span class="text-xs text-muted">or support via</span>
      <a
        href={KOFI_URL}
        rel="external"
        target="_blank"
        class="inline-flex items-center gap-1 text-xs text-[#29ABE0] no-underline transition-colors duration-150 hover:text-[#29ABE0]/80">
        <span class="flex items-center">
          <i class="i-tabler-coffee"></i>
        </span>
        Ko-fi
      </a>
    </div>
  </div>
{:else}
  <!-- Horizontal ad banner + Patreon link below -->
  <div class="overflow-hidden rounded-lg border border-outline bg-crust">
    <div class="min-h-22.5 w-full">
      <ins
        class="adsbygoogle block"
        data-ad-client={adClient}
        data-ad-slot={adSlot}
        data-ad-format="auto"
        data-full-width-responsive="true"></ins>
    </div>

    <div class="flex items-center justify-end gap-1 border-t border-outline/50 px-3 py-1.5">
      <span class="text-xs text-muted">Support us on</span>
      <a
        href={PATREON_URL}
        rel="external"
        target="_blank"
        class="inline-flex items-center gap-1 text-xs text-[#FF424D] no-underline transition-colors duration-150 hover:text-[#FF424D]/80">
        <span class="flex items-center">
          <i class="i-tabler-brand-patreon"></i>
        </span>
        Patreon
      </a>
    </div>
  </div>
{/if}
