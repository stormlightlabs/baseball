import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';
import GithubSlugger from 'github-slugger';
import { mdsvex } from 'mdsvex';

const extensions = ['.svelte', '.md'];

function headingText(node) {
  if (node.type === 'text' && typeof node.value === 'string') return node.value;
  if (!Array.isArray(node.children)) return '';
  return node.children.map((child) => headingText(child)).join('');
}

function visit(node, callback) {
  callback(node);
  if (!Array.isArray(node.children)) return;
  for (const child of node.children) {
    visit(child, callback);
  }
}

function rehypeHeadingIds() {
  return addHeadingIds;
}

function addHeadingIds(tree) {
  const slugger = new GithubSlugger();
  visit(tree, (node) => {
    if (node.type !== 'element') return;
    if (!/^h[1-6]$/.test(node.tagName)) return;
    if (node.properties?.id) return;

    const text = headingText(node).trim();
    if (!text) return;

    const id = slugger.slug(text);

    node.properties = { ...node.properties, id };
  });
}

/** @type {import('@sveltejs/kit').Config} */
const config = {
  extensions,
  preprocess: [vitePreprocess(), mdsvex({ extensions: ['.md'], rehypePlugins: [rehypeHeadingIds] })],
  compilerOptions: { runes: ({ filename }) => (filename.split(/[/\\]/).includes('node_modules') ? undefined : true) },
  kit: { adapter: adapter({ fallback: 'index.html' }), alias: { $tailwind: 'src/routes/layout.css' } }
};

export default config;
