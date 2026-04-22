<script lang="ts">
  import { resolve } from '$app/paths';

  const DOC_LINKS = [
    { href: '/docs/about', label: 'About' },
    { href: '/docs/apps', label: 'Apps' }
  ] as const;
</script>

{#snippet iconFor(label: 'About' | 'Apps')}
  {#if label === 'About'}
    <span class="flex items-center">
      <i class="i-tabler-info-circle"></i>
    </span>
  {:else if label === 'Apps'}
    <span class="flex items-center">
      <i class="i-tabler-apps"></i>
    </span>
  {/if}
{/snippet}

<footer class="border-t border-outline/70 bg-crust/70">
  <div
    class="mx-auto flex w-full max-w-6xl flex-wrap items-center gap-x-5 gap-y-2 px-4 py-2 sm:px-6 lg:justify-between lg:px-8">
    <div class="flex items-center gap-2 font-mono text-[0.62rem] tracking-[0.08em] text-muted uppercase">
      <span>Stormlight Labs</span>
    </div>

    <nav class="flex flex-wrap items-center gap-x-4 gap-y-1">
      <a
        href="https://stormlightlabs.org"
        target="_blank"
        rel="noopener noreferrer"
        class="inline-flex items-center gap-1 rounded text-sm text-muted no-underline transition-colors duration-150 hover:text-primary focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary">
        <span class="flex items-center">
          <i class="i-tabler-building"></i>
        </span>
        Company
      </a>
      <a
        href="https://github.com/stormlightlabs/baseball"
        target="_blank"
        rel="noopener noreferrer"
        class="inline-flex items-center gap-1 rounded text-sm text-muted no-underline transition-colors duration-150 hover:text-primary focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary">
        <span class="flex items-center">
          <i class="i-tabler-brand-github"></i>
        </span>
        Source
      </a>
      {#each DOC_LINKS as { href, label } (href)}
        <a
          href={resolve(href)}
          class="inline-flex items-center gap-1 rounded text-sm text-muted no-underline transition-colors duration-150 hover:text-primary focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary">
          {@render iconFor(label)}
          {label}
        </a>
      {/each}
      <a
        href="https://bsky.app/profile/bigfly.tech"
        target="_blank"
        class="inline-flex items-center gap-1 rounded text-sm text-muted no-underline transition-colors duration-150 hover:text-primary/80 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary">
        <span class="flex items-center">
          <i class="i-tabler-brand-bluesky"></i>
        </span>
        Bluesky
      </a>
      <a
        href="mailto:admin@bigfly.tech"
        class="inline-flex items-center gap-1 rounded text-sm text-muted no-underline transition-colors duration-150 hover:text-primary/80 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary">
        <span class="flex items-center">
          <i class="i-tabler-mail-code"></i>
        </span>
        Email
      </a>
    </nav>
  </div>
</footer>
