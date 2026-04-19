<script lang="ts">
  import { resolve } from '$app/paths';
  import { onMount } from 'svelte';
  import SingleColLayout from '$lib/layouts/SingleColLayout.svelte';
  import { apiHref, apiUrl } from '$lib/api';

  type AuthUser = { id: string; email: string; name?: string | null; avatar_url?: string | null };

  type ApiKey = {
    id: string;
    key_prefix: string;
    name?: string | null;
    created_at: string;
    last_used_at?: string | null;
    expires_at?: string | null;
    is_active: boolean;
  };

  type ApiKeyCreateResponse = { key: string; warning?: string };

  const githubAuthHref = apiHref('/auth/github');
  const codebergAuthHref = apiHref('/auth/codeberg');

  let loading = $state(true);
  let saving = $state(false);
  let user = $state<AuthUser | null>(null);
  let keys = $state<ApiKey[]>([]);
  let keyName = $state('');
  let createdKey = $state<string | null>(null);
  let error = $state<string | null>(null);

  onMount(async () => {
    await loadSession();
  });

  async function loadSession() {
    loading = true;
    error = null;
    createdKey = null;

    try {
      const meRes = await fetch(apiUrl('/auth/me'), { credentials: 'include' });
      if (!meRes.ok) {
        user = null;
        keys = [];
        return;
      }

      user = (await meRes.json()) as AuthUser;
      await loadKeys();
    } catch (err) {
      user = null;
      keys = [];
      error = err instanceof Error ? err.message : 'Unable to reach API';
    } finally {
      loading = false;
    }
  }

  async function loadKeys() {
    const res = await fetch(apiUrl('/auth/keys'), { credentials: 'include' });
    if (!res.ok) {
      throw new Error('Unable to load API keys');
    }
    keys = (await res.json()) as ApiKey[];
  }

  async function createKey() {
    if (!user || saving) return;
    saving = true;
    error = null;

    try {
      const res = await fetch(apiUrl('/auth/keys'), {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: keyName.trim() || null })
      });

      if (!res.ok) {
        throw new Error('Failed to create API key');
      }

      const payload = (await res.json()) as ApiKeyCreateResponse;
      createdKey = payload.key;
      keyName = '';
      await loadKeys();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to create API key';
    } finally {
      saving = false;
    }
  }

  async function revokeKey(id: string) {
    if (!user) return;
    const confirmed = globalThis.confirm('Revoke this key? This cannot be undone.');
    if (!confirmed) return;

    error = null;
    try {
      const res = await fetch(apiUrl(`/auth/keys/${id}`), { method: 'DELETE', credentials: 'include' });
      if (!res.ok) {
        throw new Error('Failed to revoke API key');
      }
      await loadKeys();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to revoke API key';
    }
  }

  async function logout() {
    try {
      await fetch(apiUrl('/auth/logout'), { method: 'POST', credentials: 'include' });
    } finally {
      await loadSession();
    }
  }

  function formatDate(value: string | null | undefined): string {
    if (!value) return 'Never';
    const parsed = new Date(value);
    if (Number.isNaN(parsed.getTime())) return value;
    return parsed.toLocaleString();
  }
</script>

<SingleColLayout>
  <section class="rounded-xl border border-outline bg-crust p-5">
    <h1 class="mb-1 font-display text-2xl text-foreground">Account</h1>
    <p class="text-sm text-muted">
      Sign in to manage API keys used by your apps and scripts.
      <a class="text-primary no-underline hover:underline" href={resolve('/docs/api-auth' as `/docs/${string}`)}>
        Auth docs
      </a>
    </p>
  </section>

  {#if loading}
    <section class="mt-4 rounded-xl border border-outline bg-crust p-5">
      <p class="text-sm text-muted">Loading session…</p>
    </section>
  {:else if !user}
    <section class="mt-4 rounded-xl border border-outline bg-crust p-5">
      <h2 class="mb-2 text-lg text-foreground">Sign in</h2>
      <p class="mb-4 text-sm text-muted">Use an OAuth provider to create/revoke API keys.</p>
      <div class="flex flex-wrap gap-2">
        <button
          type="button"
          onclick={() => globalThis.location.assign(githubAuthHref)}
          class="rounded-md bg-primary px-3 py-2 text-sm font-semibold text-white no-underline hover:bg-primary/90">
          Continue with GitHub
        </button>
        <button
          type="button"
          onclick={() => globalThis.location.assign(codebergAuthHref)}
          class="rounded-md border border-outline px-3 py-2 text-sm font-semibold text-foreground no-underline hover:bg-surface">
          Continue with Codeberg
        </button>
      </div>
      {#if error}
        <p class="mt-4 text-sm text-warning">{error}</p>
      {/if}
    </section>
  {:else}
    <section class="mt-4 rounded-xl border border-outline bg-crust p-5">
      <div class="mb-4 flex flex-wrap items-center gap-3">
        {#if user.avatar_url}
          <img src={user.avatar_url} alt="" class="h-10 w-10 rounded-full border border-outline object-cover" />
        {/if}
        <div>
          <div class="text-sm font-semibold text-foreground">{user.name || user.email}</div>
          <div class="font-mono text-xs text-muted">{user.email}</div>
        </div>
        <button
          type="button"
          class="ml-auto rounded-md border border-outline px-3 py-2 text-sm text-foreground hover:bg-surface"
          onclick={logout}>
          Log out
        </button>
      </div>

      <form
        class="grid gap-2 rounded-lg border border-outline bg-mantle p-3 sm:grid-cols-[1fr_auto]"
        onsubmit={(event) => {
          event.preventDefault();
          void createKey();
        }}>
        <input
          type="text"
          bind:value={keyName}
          placeholder="Key name (optional)"
          class="rounded-md border border-outline bg-surface px-3 py-2 text-sm text-foreground placeholder:text-muted" />
        <button
          type="submit"
          disabled={saving}
          class="rounded-md bg-primary px-3 py-2 text-sm font-semibold text-white disabled:cursor-not-allowed disabled:opacity-60">
          {saving ? 'Creating…' : 'Create API key'}
        </button>
      </form>

      {#if createdKey}
        <div class="mt-3 rounded-lg border border-secondary/40 bg-secondary/10 p-3">
          <p class="mb-2 text-sm text-foreground">Copy this key now. It is shown only once.</p>
          <code class="block overflow-x-auto rounded bg-mantle px-2 py-2 font-mono text-xs text-secondary">
            {createdKey}
          </code>
        </div>
      {/if}

      {#if error}
        <p class="mt-3 text-sm text-warning">{error}</p>
      {/if}
    </section>

    <section class="mt-4 overflow-hidden rounded-xl border border-outline bg-crust">
      <div class="border-b border-outline px-5 py-3">
        <h2 class="text-lg text-foreground">API keys</h2>
      </div>
      {#if keys.length === 0}
        <p class="px-5 py-4 text-sm text-muted">No keys yet.</p>
      {:else}
        <div class="overflow-x-auto">
          <table class="min-w-full text-sm">
            <thead>
              <tr class="border-b border-outline bg-surface/60 text-left font-mono text-xs text-muted uppercase">
                <th class="px-4 py-2">Name</th>
                <th class="px-4 py-2">Prefix</th>
                <th class="px-4 py-2">Created</th>
                <th class="px-4 py-2">Last used</th>
                <th class="px-4 py-2">Expires</th>
                <th class="px-4 py-2">Status</th>
                <th class="px-4 py-2"></th>
              </tr>
            </thead>
            <tbody>
              {#each keys as key (key.id)}
                <tr class="border-b border-outline/70 last:border-b-0">
                  <td class="px-4 py-2 text-foreground">{key.name || 'Unnamed key'}</td>
                  <td class="px-4 py-2 font-mono text-xs text-muted">{key.key_prefix}…</td>
                  <td class="px-4 py-2 text-muted">{formatDate(key.created_at)}</td>
                  <td class="px-4 py-2 text-muted">{formatDate(key.last_used_at)}</td>
                  <td class="px-4 py-2 text-muted">{formatDate(key.expires_at)}</td>
                  <td class="px-4 py-2 text-muted">{key.is_active ? 'Active' : 'Revoked'}</td>
                  <td class="px-4 py-2 text-right">
                    {#if key.is_active}
                      <button
                        type="button"
                        class="rounded-md border border-warning/50 px-2 py-1 text-xs text-warning hover:bg-warning/10"
                        onclick={() => void revokeKey(key.id)}>
                        Revoke
                      </button>
                    {/if}
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {/if}
    </section>
  {/if}
</SingleColLayout>
