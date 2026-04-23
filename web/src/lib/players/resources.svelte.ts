import type { PaginatedResponse } from '$lib/api';
import { toErrorMessage } from '$lib/leaders/utils';

type ResourceOptions = { clearError?: boolean; resetOnLoad?: boolean };

export class AsyncValueResource<T> {
  value = $state<T | null>(null);
  loading = $state(false);
  error = $state<string | null>(null);

  #requestVersion = 0;

  clear(nextValue: T | null = null): void {
    this.#requestVersion += 1;
    this.value = nextValue;
    this.loading = false;
    this.error = null;
  }

  async load(loader: () => Promise<T>, options: ResourceOptions = {}): Promise<void> {
    const { clearError = true } = options;
    const requestVersion = ++this.#requestVersion;

    this.loading = true;
    if (clearError) {
      this.error = null;
    }

    try {
      const next = await loader();
      if (requestVersion !== this.#requestVersion) return;
      this.value = next;
    } catch (error) {
      if (requestVersion !== this.#requestVersion) return;
      this.error = toErrorMessage(error, 'Request failed');
    } finally {
      if (requestVersion === this.#requestVersion) {
        this.loading = false;
      }
    }
  }
}

export class AsyncListResource<T> {
  items = $state<T[]>([]);
  loading = $state(false);
  error = $state<string | null>(null);

  #requestVersion = 0;

  clear(nextItems: T[] = []): void {
    this.#requestVersion += 1;
    this.items = nextItems;
    this.loading = false;
    this.error = null;
  }

  async load(loader: () => Promise<T[]>, options: ResourceOptions = {}): Promise<void> {
    const { clearError = true, resetOnLoad = false } = options;
    const requestVersion = ++this.#requestVersion;

    this.loading = true;
    if (clearError) {
      this.error = null;
    }
    if (resetOnLoad) {
      this.items = [];
    }

    try {
      const nextItems = await loader();
      if (requestVersion !== this.#requestVersion) return;
      this.items = nextItems;
    } catch (error) {
      if (requestVersion !== this.#requestVersion) return;
      this.error = toErrorMessage(error, 'Request failed');
    } finally {
      if (requestVersion === this.#requestVersion) {
        this.loading = false;
      }
    }
  }
}

export class AsyncPaginatedListResource<T> {
  items = $state<T[]>([]);
  total = $state(0);
  loading = $state(false);
  error = $state<string | null>(null);

  #requestVersion = 0;

  clear(nextItems: T[] = [], nextTotal = 0): void {
    this.#requestVersion += 1;
    this.items = nextItems;
    this.total = nextTotal;
    this.loading = false;
    this.error = null;
  }

  async load(loader: () => Promise<PaginatedResponse<T>>, options: ResourceOptions = {}): Promise<void> {
    const { clearError = true, resetOnLoad = false } = options;
    const requestVersion = ++this.#requestVersion;

    this.loading = true;
    if (clearError) {
      this.error = null;
    }
    if (resetOnLoad) {
      this.items = [];
      this.total = 0;
    }

    try {
      const next = await loader();
      if (requestVersion !== this.#requestVersion) return;
      this.items = next.data;
      this.total = next.total;
    } catch (error) {
      if (requestVersion !== this.#requestVersion) return;
      this.error = toErrorMessage(error, 'Request failed');
    } finally {
      if (requestVersion === this.#requestVersion) {
        this.loading = false;
      }
    }
  }
}
