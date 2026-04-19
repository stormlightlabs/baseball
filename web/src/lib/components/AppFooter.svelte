<script lang="ts">
  import { resolve } from '$app/paths';

  const DOC_LINKS = [
    { href: '/docs/about', label: 'About' },
    { href: '/docs/apps', label: 'Apps' }
  ] as const;
</script>

<footer class="border-t border-outline bg-crust">
  <div class="mx-auto grid w-full max-w-6xl gap-6 p-4 sm:px-6 lg:grid-cols-3 lg:gap-8 lg:px-8">
    <section class="space-y-2">
      <h2 class="font-mono text-xxs tracking-[0.08em] text-muted uppercase">Stormlight Labs</h2>
      <ul class="m-0 list-none space-y-1 p-0">
        <li>
          <a
            href="https://stormlightlabs.org"
            target="_blank"
            rel="noopener noreferrer"
            class="inline-flex rounded text-[0.82rem] text-foreground no-underline transition-colors duration-150 hover:text-primary focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary">
            Company
          </a>
        </li>
        <li>
          <a
            href="https://github.com/stormlightlabs/baseball"
            target="_blank"
            rel="noopener noreferrer"
            class="inline-flex gap-1 rounded text-[0.82rem] text-foreground no-underline transition-colors duration-150 hover:text-primary focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary">
            <span class="flex items-center">
              <i class="i-tabler-brand-github"></i>
            </span>
            Source
          </a>
        </li>
      </ul>
    </section>

    <section class="space-y-2">
      <h2 class="font-mono text-xxs tracking-[0.08em] text-muted uppercase">Learn</h2>
      <ul class="m-0 list-none space-y-1 p-0">
        {#each DOC_LINKS as { href, label } (href)}
          <li>
            <a
              href={resolve(href)}
              class="inline-flex rounded text-[0.82rem] text-foreground no-underline transition-colors duration-150 hover:text-primary focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary">
              {label}
            </a>
          </li>
        {/each}
      </ul>
    </section>

    <section class="space-y-2">
      <h2 class="flex items-center gap-1 font-mono text-xxs tracking-[0.08em] text-muted uppercase">
        <span class="flex items-center">
          <i class="i-tabler-info-circle"></i>
        </span>
        About Big Fly
      </h2>
      <p class="m-0 max-w-prose text-[0.82rem] leading-normal text-muted">
        Big Fly is a baseball research app & API backed by unified data from Lahman, Retrosheet, and MLB.
      </p>
    </section>
  </div>
</footer>
