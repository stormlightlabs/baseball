<script lang="ts">
  import { DOCS_BY_SLUG } from '$lib/docs/catalog';

  let { data } = $props();

  let doc = $derived(DOCS_BY_SLUG[data.slug]);
  let DocComponent = $derived(doc?.component);
</script>

{#if DocComponent}
  <article class="doc-body">
    <DocComponent />
  </article>
{/if}
