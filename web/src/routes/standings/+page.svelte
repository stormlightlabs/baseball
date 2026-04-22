<script lang="ts">
  import StandingsPanel from '$lib/components/StandingsPanel.svelte';

  const CURRENT_SEASON = new Date().getFullYear();
</script>

<div class="mx-auto w-full max-w-6xl px-4 py-6 sm:px-6 lg:px-8">
  <div class="mb-4 rounded-lg border border-outline bg-crust p-4">
    <h1 class="font-display text-[1.1rem] text-foreground">Standings</h1>
    <p class="mt-1 font-sans text-[0.72rem] text-muted">
      Division standings for season {CURRENT_SEASON}, sourced from local `current_season` data when available.
    </p>
  </div>

  <StandingsPanel season={CURRENT_SEASON} title="Current standings" />
</div>
