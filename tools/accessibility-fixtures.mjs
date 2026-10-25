// Prepare exact API replays for issue #16. Does not run diagnostics or infer findings.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import { fileURLToPath } from 'node:url';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const output = path.join(root, '.tools', 'issue-16');
fs.mkdirSync(output, { recursive: true });
const manifest = [];
for (const issue of [11, 12, 13, 14]) {
  const source = path.join(root, '.tools', `issue-${issue}`);
  for (const name of fs.readdirSync(source).filter(name => /^run-[a-z0-9]+\.json$/.test(name))) {
    const bytes = fs.readFileSync(path.join(source, name));
    const mode = `i${issue}${name.slice(4, -5)}`;
    fs.writeFileSync(path.join(output, `run-${mode}.json`), bytes);
    manifest.push({ mode, source: `.tools/issue-${issue}/${name}`, sha256: crypto.createHash('sha256').update(bytes).digest('hex') });
  }
}

// Presentation stress only: preserve outcomes, numeric measurements and Go finding strings.
// This variant is never native diagnostic or human acceptance evidence by itself.
const stress = JSON.parse(fs.readFileSync(path.join(output, 'run-i13limit64.json'), 'utf8'));
stress.target_url = `https://fixture.example/${'long-path-'.repeat(45)}?query=${'a'.repeat(200)}`;
stress.families[0].exchanges[0].requested_url = stress.target_url;
const ids = new Set();
function visit(value) {
  if (!value || typeof value !== 'object') return;
  if (typeof value.id === 'string') ids.add(value.id);
  Object.values(value).forEach(visit);
}
visit(stress);
function lengthen(value) {
  if (typeof value === 'string') return ids.has(value) ? `${value}-${'long-identifier-'.repeat(12)}` : value;
  if (!value || typeof value !== 'object') return value;
  if (Array.isArray(value)) return value.map(lengthen);
  return Object.fromEntries(Object.entries(value).map(([key, item]) => [key, lengthen(item)]));
}
fs.writeFileSync(path.join(output, 'run-stress.json'), JSON.stringify(lengthen(stress), null, 2));
manifest.push({ mode: 'stress', source: 'i13limit64', limitation: 'Synthetic URL/ID length variant for layout only; original structured outcomes and Go finding strings retained.' });
fs.writeFileSync(path.join(output, 'manifest.json'), JSON.stringify(manifest, null, 2));
console.log(`Prepared ${manifest.length} review modes in ${output}. See manifest.json for provenance.`);
console.log('Replay with HOPVISTA_JOURNEY_CAPTURE set to that directory and node tools/https-journey-preview.mjs.');
