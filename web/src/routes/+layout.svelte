<script lang="ts">
  import favicon from '$lib/assets/favicon.svg';
  import '@fontsource-variable/google-sans';
  import '@fontsource-variable/google-sans-code';
  import '@fontsource-variable/inter';
  import './layout.css';
  let { children } = $props();
</script>

<svelte:head><link rel="icon" href={favicon} /></svelte:head>
{@render children()}
