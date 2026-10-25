// Supporting browser measurements, not a WCAG certification or human review.
import fs from 'node:fs';
import path from 'node:path';
import { createRequire } from 'node:module';
import { fileURLToPath } from 'node:url';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const { chromium } = createRequire(path.join(root, 'ui/package.json'))('playwright');
const url = process.env.HOPVISTA_REVIEW_URL;
if (!url || new URL(url).hostname !== '127.0.0.1') throw new Error('Set HOPVISTA_REVIEW_URL to the loopback replay URL.');
const output = path.join(root, '.impeccable/review/issue-16');
fs.mkdirSync(output, { recursive: true });
const browser = await chromium.launch({ channel: process.env.HOPVISTA_BROWSER || 'msedge', headless: true });
const pairs = new Map();
const report = { browser: browser.version(), failures: [] };
try {
  const page = await browser.newPage();
  let run;
  await page.route('**/api/v1/runs', route => route.fulfill({ status: 202, json: { run_id: 'review' } }));
  await page.route('**/api/v1/runs/review', route => route.fulfill({ json: run }));
  for (const mode of ['i11completed', 'i12tcptimeout', 'i12noipv4', 'i12running', 'i12stopping', 'i14nativecancel', 'stress']) {
    run = JSON.parse(fs.readFileSync(path.join(root, `.tools/issue-16/run-${mode}.json`), 'utf8'));
    await page.goto(url);
    await page.getByLabel('Target URL', { exact: true }).fill('https://fixture.example/review');
    await page.getByRole('button', { name: 'Start diagnostic run', exact: true }).press('Enter');
    await page.locator('.run-heading').waitFor();
    await page.locator('.technical > summary').press('Enter');
    const readings = await page.evaluate(() => {
      const luminance = color => color.match(/[\d.]+/g).slice(0, 3).map(Number)
        .map(value => value / 255)
        .map(value => value <= 0.04045 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4)
        .reduce((sum, value, index) => sum + value * [0.2126, 0.7152, 0.0722][index], 0);
      const background = element => {
        for (let parent = element; parent; parent = parent.parentElement) {
          const color = getComputedStyle(parent).backgroundColor;
          if (color !== 'rgba(0, 0, 0, 0)' && color !== 'transparent') return color;
        }
        return 'rgb(255, 255, 255)';
      };
      const result = [];
      const walker = document.createTreeWalker(document.body, NodeFilter.SHOW_TEXT);
      let node;
      while ((node = walker.nextNode())) {
        const element = node.parentElement;
        if (!node.textContent.trim() || !element.checkVisibility({ checkVisibilityCSS: true, checkOpacity: true })
          || !element.getClientRects().length || element.closest('.visually-hidden') || element.getBoundingClientRect().width <= 1) continue;
        const style = getComputedStyle(element);
        const foreground = style.color;
        const surface = background(element);
        const levels = [luminance(foreground), luminance(surface)].sort((a, b) => b - a);
        const large = parseFloat(style.fontSize) >= 24 || (parseFloat(style.fontSize) >= 18.666 && parseFloat(style.fontWeight) >= 700);
        result.push({ foreground, background: surface, ratio: (levels[0] + 0.05) / (levels[1] + 0.05),
          threshold: large ? 3 : 4.5, disabled: !!element.closest(':disabled,[aria-disabled=true]'), sample: node.textContent.trim().slice(0, 70) });
      }
      const placeholder = getComputedStyle(document.querySelector('#target'), '::placeholder').color;
      const levels = [luminance(placeholder), 1].sort((a, b) => b - a);
      result.push({ foreground: placeholder, background: 'rgb(255, 255, 255)', ratio: (levels[0] + 0.05) / (levels[1] + 0.05), threshold: 4.5, sample: 'placeholder' });
      return result;
    });
    for (const reading of readings) {
      pairs.set(`${reading.foreground}/${reading.background}/${reading.disabled}`, reading);
      if (reading.ratio < reading.threshold && !reading.disabled) report.failures.push({ mode, ...reading });
    }
  }
  await page.setViewportSize({ width: 320, height: 1000 });
  // Explicitly synthetic rejection; native browser policy is checked separately in UI tests.
  await page.evaluate(() => Object.defineProperty(navigator, 'clipboard', { value: { writeText: async () => { throw Error('Synthetic denial'); } } }));
  await page.getByRole('button', { name: /^Copy selected IPv4 address/ }).press('Enter');
  await page.locator('.manual-copy input').waitFor();
  report.fallback = await page.evaluate(() => ({ width: innerWidth, scrollWidth: document.documentElement.scrollWidth,
    address: document.activeElement.value, selected: document.activeElement.value.slice(document.activeElement.selectionStart, document.activeElement.selectionEnd) }));
  await page.screenshot({ path: path.join(output, 'copy-fallback-320.png') });
  const coarse = await browser.newContext({ hasTouch: true, viewport: { width: 390, height: 844 } });
  const touchPage = await coarse.newPage();
  await touchPage.goto(url);
  await touchPage.getByLabel('Target URL', { exact: true }).fill('https://fixture.example/i13highsilent');
  await touchPage.getByRole('button', { name: 'Start diagnostic run', exact: true }).press('Enter');
  await touchPage.locator('.run-heading').waitFor();
  report.targets = await touchPage.evaluate(() => ({ coarse: matchMedia('(pointer: coarse)').matches,
    controls: [...document.querySelectorAll('button,summary,input,a')]
      .filter(element => element.checkVisibility() && element.getClientRects().length && !element.classList.contains('skip-link'))
      .map(element => ({ label: element.getAttribute('aria-label') || element.textContent || element.id,
        width: element.getBoundingClientRect().width, height: element.getBoundingClientRect().height })) }));
} finally {
  report.pairs = [...pairs.values()];
  fs.writeFileSync(path.join(output, `metrics-${new Date().toISOString().replaceAll(':', '-')}.json`), JSON.stringify(report, null, 2));
  await browser.close();
}
console.log(JSON.stringify({ pairs: pairs.size, failures: report.failures, fallback: report.fallback,
  shortCoarse: report.targets.controls.filter(control => control.height < 44) }, null, 2));
if (report.failures.length || report.fallback.scrollWidth > report.fallback.width) process.exitCode = 1;
