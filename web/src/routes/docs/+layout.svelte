<script lang="ts">
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import { DOC_GROUPS } from '$lib/docs/catalog';

  let { children } = $props();

  let search = $state('');

  const filteredGroups = $derived.by(() => {
    const query = search.trim().toLowerCase();
    return DOC_GROUPS.map((group) => {
      if (!query) return group;
      return {
        group: group.group,
        docs: group.docs.filter((doc) =>
          [doc.slug, doc.title, doc.navTitle].some((value) => value.toLowerCase().includes(query))
        )
      };
    }).filter((group) => group.docs.length > 0);
  });

  let currentSlug = $derived(page.params.slug ?? '');
  let activeDocData = $derived(
    (page.data as {
      group?: string;
      navTitle?: string;
      sourceFile?: string;
      toc?: Array<{ depth: 2 | 3; text: string; id: string }>;
    }) ?? {}
  );

  function jumpToHeading(id: string) {
    const heading = document.getElementById(id);
    if (!heading) return;

    heading.scrollIntoView({ behavior: 'smooth', block: 'start' });

    const path = currentSlug ? resolve(`/docs/${currentSlug}` as `/docs/${string}`) : resolve('/docs');
    history.replaceState(history.state, '', `${path}#${id}`);
  }
</script>

<main class="min-h-full bg-mantle">
  <div
    class="mx-auto grid w-full max-w-416 grid-cols-[18rem_minmax(0,1fr)_14rem] gap-4 px-6 pt-5 pb-7 max-[74rem]:grid-cols-[17rem_minmax(0,1fr)] max-[56rem]:grid-cols-1 max-[56rem]:px-[0.9rem] max-[56rem]:pt-[0.9rem] max-[56rem]:pb-[1.2rem]">
    <nav
      class="sticky top-19 flex max-h-[calc(100vh-6rem)] flex-col self-start overflow-hidden rounded-xl border border-outline bg-crust max-[56rem]:static max-[56rem]:max-h-none">
      <div class="border-b border-outline bg-crust p-3">
        <input
          class="w-full rounded-[0.6rem] border border-outline bg-surface px-[0.65rem] py-2 text-[0.8rem] text-foreground placeholder:text-muted/90"
          type="search"
          bind:value={search}
          placeholder="Search docs…"
          aria-label="Search docs" />
      </div>

      <div class="space-y-[0.9rem] overflow-auto px-3 pt-[0.6rem] pb-3">
        {#each filteredGroups as group (group.group)}
          <section>
            <h2 class="mb-[0.45rem] px-[0.35rem] font-mono text-[0.63rem] tracking-[0.08em] text-muted uppercase">
              {group.group}
            </h2>
            <ul class="m-0 list-none p-0">
              {#each group.docs as doc (doc.slug)}
                <li>
                  <a
                    href={resolve(`/docs/${doc.slug}` as `/docs/${string}`)}
                    class="block rounded-lg px-[0.45rem] py-[0.38rem] text-[0.8rem] leading-[1.35] no-underline transition-colors duration-150 {currentSlug ===
                    doc.slug
                      ? 'bg-primary/20 text-primary'
                      : 'text-foreground hover:bg-surface'}">
                    {doc.navTitle}
                  </a>
                </li>
              {/each}
            </ul>
          </section>
        {/each}
      </div>
    </nav>

    <section
      class="sticky top-19 h-[calc(100vh-6rem)] self-start overflow-auto rounded-xl border border-outline bg-crust p-4 pb-8 max-[56rem]:static max-[56rem]:h-auto max-[56rem]:max-h-none max-[56rem]:overflow-visible max-[56rem]:px-4 max-[56rem]:pt-[0.95rem] max-[56rem]:pb-[1.4rem]">
      {#if activeDocData.navTitle}
        <div
          class="mb-4 flex flex-wrap items-center gap-[0.45rem] border-b border-outline pb-[0.6rem] text-[0.74rem] text-muted">
          <span>{activeDocData.group}</span>
          <span>›</span>
          <span class="font-semibold text-foreground">{activeDocData.navTitle}</span>
          {#if activeDocData.sourceFile}
            <span
              class="ml-auto rounded-full border border-outline px-2 py-[0.18rem] font-mono text-[0.67rem] max-[56rem]:ml-0">
              {activeDocData.sourceFile}
            </span>
          {/if}
        </div>
      {/if}
      {@render children()}
    </section>

    <aside
      class="sticky top-19 max-h-[calc(100vh-6rem)] self-start overflow-auto rounded-xl border border-outline bg-crust p-[0.8rem] px-3 max-[74rem]:hidden max-[56rem]:static max-[56rem]:max-h-none">
      <h2 class="mb-2 font-mono text-[0.65rem] tracking-[0.08em] text-muted uppercase">On this page</h2>
      <nav class="flex flex-col gap-[0.3rem]">
        {#if activeDocData.toc?.length}
          {#each activeDocData.toc as item (`${item.id}-${item.depth}`)}
            <button
              type="button"
              onclick={() => jumpToHeading(item.id)}
              class="cursor-pointer border-0 bg-transparent p-0 text-left leading-[1.3] text-foreground no-underline transition-colors duration-150 hover:text-primary {item.depth ===
              3
                ? 'pl-3 text-[0.74rem] text-muted'
                : 'text-[0.78rem]'}">
              {item.text}
            </button>
          {/each}
        {:else}
          <p class="m-0 text-[0.76rem] text-muted">No headings found.</p>
        {/if}
      </nav>
    </aside>
  </div>
</main>

<style lang="postcss">
  @reference "$tailwind";

  :global(.doc-body) {
    @apply w-full max-w-none text-[0.93rem] leading-[1.62] text-foreground;
  }

  :global(.doc-body h1) {
    @apply mb-4 font-display text-[1.9rem] leading-[1.18] text-foreground;
  }

  :global(.doc-body h2) {
    @apply mt-8 mb-3 scroll-mt-20 font-display text-[1.22rem] leading-tight text-foreground;
  }

  :global(.doc-body h3) {
    @apply mt-[1.4rem] mb-[0.6rem] scroll-mt-20 text-[1.02rem] leading-tight text-foreground;
  }

  :global(.doc-body p),
  :global(.doc-body ul),
  :global(.doc-body ol),
  :global(.doc-body pre),
  :global(.doc-body table) {
    @apply mb-4;
  }

  :global(.doc-body ul) {
    @apply list-disc pl-5;
  }

  :global(.doc-body ol) {
    @apply list-decimal pl-5;
  }

  :global(.doc-body li + li) {
    @apply mt-[0.35rem];
  }

  :global(.doc-body a) {
    @apply text-primary underline-offset-[0.16em];
  }

  :global(.doc-body code) {
    @apply rounded-[0.35rem] border border-outline bg-surface px-[0.32rem] py-[0.1rem] font-mono text-[0.82em] text-primary;
  }

  :global(.doc-body pre) {
    @apply overflow-x-auto rounded-[0.6rem] border border-outline bg-surface px-[0.95rem] py-[0.85rem];
  }

  :global(.doc-body pre code) {
    @apply border-0 bg-transparent p-0 text-foreground;
  }

  :global(.doc-body table) {
    @apply w-full border-collapse border border-outline;
  }

  :global(.doc-body th),
  :global(.doc-body td) {
    @apply border-b border-outline px-[0.55rem] py-[0.45rem] text-left align-top;
  }

  :global(.doc-body th) {
    @apply bg-surface font-mono text-[0.75rem] tracking-[0.02em] text-muted uppercase;
  }
</style>
