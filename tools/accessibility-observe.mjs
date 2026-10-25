// Rendered acceptance checks against exact API replays. Not native measurements,
// a real screen reader, or human comprehension acceptance.
import fs from 'node:fs';
import path from 'node:path';
import assert from 'node:assert/strict';
import { createRequire } from 'node:module';
import { fileURLToPath } from 'node:url';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const { chromium } = createRequire(path.join(root, 'ui/package.json'))('playwright');
const url = process.env.HOPVISTA_REVIEW_URL;
if (!url || new URL(url).hostname !== '127.0.0.1') throw new Error('Set HOPVISTA_REVIEW_URL to the loopback replay URL.');
const output = path.join(root, '.impeccable/review/issue-16');
const captures = path.join(root, '.tools/issue-16');
fs.mkdirSync(output, { recursive: true });
const manifest = JSON.parse(fs.readFileSync(path.join(captures, 'manifest.json'), 'utf8'));
const filter = process.env.HOPVISTA_REVIEW_MODES?.split(',');
const records = manifest.filter(item => !filter || filter.includes(item.mode));
if (!records.length) throw new Error('No review modes selected.');
const browser = await chromium.launch({ channel: process.env.HOPVISTA_BROWSER || 'msedge', headless: true });
const report = { browser: browser.version(), url, observations: [], failures: [] };
const reportPath = path.join(output, `observations-${new Date().toISOString().replaceAll(':', '-')}.json`);
const save = () => fs.writeFileSync(reportPath, JSON.stringify(report, null, 2));
try {
  const context = await browser.newContext();
  const page = await context.newPage();
  let current;
  await page.route('**/api/v1/runs', route => route.fulfill({ status: 202, json: { run_id: 'review' } }));
  await page.route('**/api/v1/runs/review', route => route.fulfill({ json: current }));
  for (const record of records) {
    current = JSON.parse(fs.readFileSync(path.join(captures, `run-${record.mode}.json`), 'utf8'));
    for (const width of [1440, 960, 760, 390, 320]) {
      try {
        await page.setViewportSize({ width, height: 1000 });
        await page.goto(url);
        await page.getByLabel('Target URL', { exact: true }).fill('https://fixture.example/review');
        await page.getByRole('button', { name: 'Start diagnostic run', exact: true }).press('Enter');
        await page.locator('.run-heading').waitFor();
        await page.locator('.technical > summary').press('Enter');
        const source = page.locator('.timing .evidence-link').last();
        const hasEvidence = await source.count() > 0;
        if (hasEvidence) {
          await source.press('Enter');
          await page.locator('.evidence-item.is-selected').waitFor();
          assert.equal(await source.getAttribute('title'), await page.locator('.is-selected h4').innerText());
          assert.equal(await page.locator('.is-selected').evaluate(item => item === document.activeElement), true);
          await page.keyboard.press('Tab');
          await page.keyboard.press('Enter');
          assert.equal(await source.evaluate(button => button === document.activeElement), true);
        }
        await page.getByText('Full run result JSON', { exact: true }).press('Enter');
        const observation = await page.evaluate(() => {
          const box = selector => document.querySelector(selector)?.getBoundingClientRect().toJSON();
          return { width: innerWidth, scrollWidth: document.documentElement.scrollWidth,
            summary: box('.summary-column'), path: box('.path-column'), timing: box('.timing'),
            hopCount: document.querySelectorAll('.hop-list li').length,
            focus: document.activeElement.textContent,
            finding: document.querySelector('.observation')?.textContent,
            answer: document.querySelector('.answer')?.textContent };
        });
        report.observations.push({ mode: record.mode, ...observation });
        assert.ok(observation.scrollWidth <= width, `Page overflow: ${observation.scrollWidth} > ${width}`);
        const family = current.families.find(item => item.family === 'ipv4');
        assert.equal(observation.hopCount, family?.paths.at(-1)?.trace.hops.length || 0);
        const finding = current.findings.find(item => item.family === 'ipv4');
        if (['running', 'stopping'].includes(current.state)) {
          assert.equal(observation.finding, undefined);
          assert.match(observation.answer, /Pending until the diagnostic run finishes/);
        } else if (finding) assert.equal(observation.finding, finding.observation);
        if (width <= 760) assert.ok(observation.summary.y < observation.path.y && observation.timing.y < observation.path.y);
        else assert.ok(observation.path.x < observation.summary.x);
        if (record.mode === 'stress' && [1440, 320].includes(width)) await page.screenshot({ path: path.join(output, `stress-${width}.png`) });
      } catch (error) { report.failures.push({ mode: record.mode, width, error: String(error) }); }
      save();
    }
    console.log(`Measured ${record.mode}`);
  }
  // In a real running browser, emulate the preference rather than only read CSS.
  await page.emulateMedia({ reducedMotion: 'reduce' });
  await page.locator('.timing .evidence-link').last().press('Enter');
  report.reducedMotion = await page.evaluate(() => ({
    matches: matchMedia('(prefers-reduced-motion: reduce)').matches,
    animations: document.getAnimations().length,
    selected: getComputedStyle(document.querySelector('.is-selected')).animationName,
    arrow: getComputedStyle(document.querySelector('.evidence-link svg')).transform,
  }));
  assert.deepEqual(report.reducedMotion, { matches: true, animations: 0, selected: 'none', arrow: 'none' });
  await context.close();
} finally {
  save();
  await browser.close();
}
console.log(JSON.stringify({ reportPath, cases: report.observations.length, failures: report.failures, reducedMotion: report.reducedMotion }, null, 2));
if (report.failures.length) process.exitCode = 1;
