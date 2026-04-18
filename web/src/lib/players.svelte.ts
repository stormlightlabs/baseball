import type { ParamValue } from '$lib/url-state.svelte';
import { SvelteMap } from 'svelte/reactivity';

type UrlParamMap = SvelteMap<string, ParamValue>;

export type PlayersUiSnapshot = { q: string; tab: string; page: number; perPage: number };

export class PlayersUiController {
  search = $state('');
  tab = $state('batting');
  page = $state(1);
  perPage = $state(20);

  #urlTab = $state('batting');
  #urlPage = $state(1);
  #urlPerPage = $state(20);
  #syncingFromUrl = $state(false);

  #toParamMap(entries: Array<[string, ParamValue]>): UrlParamMap {
    const map = new SvelteMap<string, ParamValue>();
    for (const [key, value] of entries) {
      map.set(key, value);
    }
    return map;
  }

  nextUrlParams = $derived.by((): UrlParamMap | null => {
    if (this.#syncingFromUrl) return null;
    if (this.tab !== this.#urlTab)
      return this.#toParamMap([
        ['tab', this.tab],
        ['page', 1]
      ]);
    if (this.perPage !== this.#urlPerPage)
      return this.#toParamMap([
        ['per_page', this.perPage],
        ['page', 1]
      ]);
    if (this.page !== this.#urlPage) return this.#toParamMap([['page', this.page]]);
    return null;
  });

  syncFromUrl(snapshot: PlayersUiSnapshot): void {
    this.#syncingFromUrl = true;
    this.search = snapshot.q;
    this.tab = snapshot.tab;
    this.page = snapshot.page;
    this.perPage = snapshot.perPage;
    this.#urlTab = snapshot.tab;
    this.#urlPage = snapshot.page;
    this.#urlPerPage = snapshot.perPage;
    this.#syncingFromUrl = false;
  }
}
