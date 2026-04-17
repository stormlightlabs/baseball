// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
declare global {
  namespace App {
    // interface Error {}
    // interface Locals {}
    // interface PageData {}
    // interface PageState {}
    // interface Platform {}
  }
}

// Fontsource packages ship CSS only; declare them so TS doesn't complain.
declare module '@fontsource-variable/google-sans' {}
declare module '@fontsource-variable/google-sans-code' {}
declare module '@fontsource-variable/inter' {}

export {};
