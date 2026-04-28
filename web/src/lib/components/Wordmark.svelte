<script lang="ts">
  import favicon from '$lib/assets/favicon.svg';

  let { size = 'sm' }: { size?: 'sm' | 'lg' } = $props();

  const iconSize = $derived(size === 'lg' ? 32 : 20);
  const gap = $derived(size === 'lg' ? 'gap-2.5' : 'gap-1.5');
</script>

<span class="inline-flex items-center {gap}">
  <img
    src={favicon}
    width={iconSize}
    height={iconSize}
    alt=""
    aria-hidden="true"
    class="shrink-0 rounded-full outline outline-1 outline-outline/60" />
  <span>Big <span class="text-primary">Fly</span></span>
</span>
