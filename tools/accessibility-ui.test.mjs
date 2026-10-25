// Rendered production UI seam. API snapshots are controlled schema-1 fixtures,
// not native measurements or a substitute for screen-reader/human review.
import { test, before, after } from 'node:test';
import assert from 'node:assert/strict';
import http from 'node:http';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { createRequire } from 'node:module';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const { chromium } = createRequire(path.join(root, 'ui', 'package.json'))('playwright');
const fixture = JSON.parse(fs.readFileSync(path.join(root, 'internal/result/testdata/completed.json'), 'utf8'));
let browser, server, origin;
before(async () => {
  server = http.createServer((req, res) => {
    const url = new URL(req.url, 'http://127.0.0.1');
    const name = url.pathname === '/' ? 'index.html' : url.pathname.slice(1);
    if (!['index.html', 'app.js', 'style.css'].includes(name)) return res.writeHead(404).end();
    if (url.searchParams.has('block-clipboard')) res.setHeader('Permissions-Policy', 'clipboard-write=()');
    res.setHeader('Content-Type', name.endsWith('.js') ? 'text/javascript' : name.endsWith('.css') ? 'text/css' : 'text/html');
    res.end(fs.readFileSync(path.join(root, 'internal/localapi/static', name)));
  });
  await new Promise(resolve => server.listen(0, '127.0.0.1', resolve));
  origin = `http://127.0.0.1:${server.address().port}`;
  browser = await chromium.launch({ channel: process.env.HOPVISTA_BROWSER || 'msedge', headless: true });
});
after(async () => {
  await browser?.close();
  await new Promise(resolve => server?.close(resolve));
});

async function review(t, state = 'completed') {
  const context = await browser.newContext({ viewport: { width: 1440, height: 1000 } });
  t.after(() => context.close());
  const page = await context.newPage();
  const run = structuredClone(fixture);
  run.state = state;
  if (['running', 'stopping'].includes(state)) { run.findings = []; delete run.ended_at; }
  await page.route('**/api/v1/session', route => route.fulfill({ json: { mutation_token: 'ui-test' } }));
  await page.route('**/api/v1/runs', route => route.fulfill({ status: 202, json: { run_id: 'ui-test' } }));
  await page.route('**/api/v1/runs/ui-test', route => route.fulfill({ json: run }));
  await page.goto(origin);
  await page.getByLabel('Target URL', { exact: true }).fill('https://fixture.example/');
  return { page, run, context };
}

test('starting a diagnostic run keeps keyboard focus in the diagnostic workspace', async t => {
  const { page } = await review(t, 'running');
  await page.getByRole('button', { name: 'Start diagnostic run', exact: true }).press('Enter');
  await page.locator('.run-heading').waitFor();
  assert.equal(await page.evaluate(() => document.activeElement?.getAttribute('role')), 'status');
  assert.match(await page.locator('.status').innerText(), /running/i);
});

test('a failed start leaves an announced error and allows a keyboard retry', async t => {
  const { page } = await review(t);
  let rejected = true;
  await page.route('**/api/v1/runs', route => route.fulfill(rejected
    ? { status: 400, json: { error: { message: 'Controlled invalid target.' } } }
    : { status: 202, json: { run_id: 'ui-test' } }));
  const start = page.getByRole('button', { name: 'Start diagnostic run', exact: true });
  await start.press('Enter');
  await page.getByRole('alert').waitFor();
  assert.match(await page.getByRole('alert').innerText(), /Could not start.*Controlled invalid target/);
  assert.doesNotMatch(await page.locator('.status').innerText(), /Starting/);
  assert.equal(await start.isEnabled(), true);
  rejected = false;
  await page.keyboard.press('Shift+Tab');
  assert.equal(await page.evaluate(() => document.activeElement.textContent), 'Advanced');
  await page.keyboard.press('Shift+Tab');
  assert.equal(await page.evaluate(() => document.activeElement.textContent), 'Start diagnostic run');
  await page.keyboard.press('Enter');
  await page.locator('.run-heading').waitFor();
  assert.equal(await page.locator('.request-error').count(), 0);
});

test('clipboard failure offers a keyboard-selectable address for manual copying', async t => {
  const { page } = await review(t);
  // Synthetic browser API rejection, never counted as native permission denial.
  await page.evaluate(() => Object.defineProperty(navigator, 'clipboard', {
    configurable: true, value: { writeText: async () => { throw new Error('Synthetic clipboard denial'); } },
  }));
  await page.getByRole('button', { name: 'Start diagnostic run', exact: true }).press('Enter');
  await page.locator('.run-heading').waitFor();
  await page.getByRole('button', { name: /^Copy selected IPv4 address/ }).press('Enter');
  await page.locator('.copy-feedback.error').waitFor();
  const manual = page.getByRole('textbox', { name: 'IPv4 address for manual copying', exact: true });
  assert.equal(await manual.count(), 1);
  assert.equal(await manual.inputValue(), '192.0.2.10');
  assert.deepEqual(await manual.evaluate(input => ({ focused: input === document.activeElement, selected: input.value.slice(input.selectionStart, input.selectionEnd) })),
    { focused: true, selected: '192.0.2.10' });
  assert.equal(await page.locator('main').getAttribute('data-state'), 'completed');
});

test('waiting for the clipboard preserves focus through the copy acknowledgement', async t => {
  const { page } = await review(t);
  await page.evaluate(() => Object.defineProperty(navigator, 'clipboard', {
    configurable: true, value: { writeText: () => new Promise(resolve => { window.finishCopy = resolve; }) },
  }));
  await page.getByRole('button', { name: 'Start diagnostic run', exact: true }).press('Enter');
  await page.locator('.run-heading').waitFor();
  const copy = page.getByRole('button', { name: /^Copy selected IPv4 address/ });
  await copy.press('Enter');
  await page.waitForFunction(() => document.querySelector('.copy-button')?.textContent === 'Copying…');
  assert.equal(await copy.evaluate(button => button === document.activeElement), true);
  await page.evaluate(() => window.finishCopy());
  await page.waitForFunction(() => document.querySelector('.copy-button')?.textContent === 'Copied');
  assert.equal(await copy.evaluate(button => button === document.activeElement), true);
  assert.match(await page.locator('.copy-address [role=status]').innerText(), /IPv4 address copied/);
});

test('cancellation retains focus while stopping and after the control disappears', async t => {
  const { page, run } = await review(t, 'running');
  await page.route('**/api/v1/runs/ui-test/cancel', route => {
    run.state = 'stopping';
    return route.fulfill({ json: run });
  });
  await page.getByRole('button', { name: 'Start diagnostic run', exact: true }).press('Enter');
  await page.locator('.run-heading').waitFor();
  await page.keyboard.press('Shift+Tab');
  assert.equal(await page.evaluate(() => document.activeElement.textContent), 'Advanced');
  await page.keyboard.press('Shift+Tab');
  assert.equal(await page.evaluate(() => document.activeElement.textContent), 'Cancel run');
  await page.keyboard.press('Enter');
  await page.waitForFunction(() => document.querySelector('main').dataset.state === 'stopping');
  assert.equal(await page.evaluate(() => document.activeElement.getAttribute('role')), 'status');
  assert.match(await page.locator('.answer').innerText(), /Pending until the diagnostic run finishes/);
  assert.equal(await page.getByRole('button', { name: 'Stopping…', exact: true }).isEnabled(), false);
  run.state = 'cancelled';
  await page.waitForFunction(() => document.querySelector('main').dataset.state === 'cancelled');
  assert.equal(await page.evaluate(() => document.activeElement.getAttribute('role')), 'status');
  assert.equal(await page.getByRole('button', { name: 'Start diagnostic run', exact: true }).isEnabled(), true);
});

test('an unconfirmed cancellation keeps the last result and exposes retry uncertainty', async t => {
  const { page } = await review(t, 'running');
  await page.route('**/api/v1/runs/ui-test/cancel', route => route.fulfill({ status: 503, json: { error: { message: 'Controlled cancellation failure.' } } }));
  await page.getByRole('button', { name: 'Start diagnostic run', exact: true }).press('Enter');
  await page.locator('.run-heading').waitFor();
  // A failed refresh must not erase the cancellation error or invent a terminal state.
  await page.route('**/api/v1/runs/ui-test', route => route.abort());
  await page.getByRole('button', { name: 'Cancel run', exact: true }).press('Enter');
  await page.getByRole('alert').waitFor();
  assert.doesNotMatch(await page.locator('.status').innerText(), /Requesting stop/);
  assert.equal(await page.locator('main').getAttribute('data-state'), 'running');
  assert.equal(await page.getByRole('button', { name: 'Cancel run', exact: true }).isEnabled(), true);
});

test('keyboard entry, advanced settings, invalid bounds and the skip link remain usable', async t => {
  const { page } = await review(t);
  await page.reload();
  await page.keyboard.press('Tab');
  assert.equal(await page.evaluate(() => document.activeElement.textContent), 'Skip to diagnostic run');
  await page.keyboard.press('Enter');
  await page.keyboard.press('Tab');
  assert.equal(await page.evaluate(() => document.activeElement.id), 'target');
  await page.keyboard.type('https://fixture.example/');
  await page.keyboard.press('Tab');
  assert.equal(await page.evaluate(() => document.activeElement.textContent), 'Start diagnostic run');
  await page.keyboard.press('Tab');
  assert.equal(await page.evaluate(() => document.activeElement.textContent), 'Advanced');
  await page.keyboard.press('Enter');
  await page.keyboard.press('Tab');
  assert.equal(await page.evaluate(() => document.activeElement.id), 'hops');
  for (const value of ['0', '65', '1.5']) {
    await page.getByLabel('Maximum hops (1–64)', { exact: true }).fill(value);
    await page.keyboard.press('Enter');
    assert.equal(await page.locator('#hops').evaluate(input => input.checkValidity()), false);
    assert.equal(await page.locator('.run-heading').count(), 0);
  }
  await page.getByLabel('Maximum hops (1–64)', { exact: true }).fill('64');
  const request = page.waitForRequest(request => request.method() === 'POST' && new URL(request.url()).pathname === '/api/v1/runs');
  await page.keyboard.press('Enter');
  assert.deepEqual((await request).postDataJSON(), { target_url: 'https://fixture.example/', max_hops: 64 });
  await page.locator('.run-heading').waitFor();
});

test('each website phase opens its exact evidence and returns keyboard focus', async t => {
  const { page } = await review(t);
  await page.getByRole('button', { name: 'Start diagnostic run', exact: true }).press('Enter');
  await page.locator('.run-heading').waitFor();
  for (const id of ['exchange-1-dns', 'tcp-1', 'tls-1', 'http-1']) {
    const source = page.getByRole('button', { name: `View evidence ${id}`, exact: true });
    await source.press('Enter');
    await page.waitForFunction(id => document.activeElement.id === `evidence-${id}`, id);
    assert.match(await page.locator(`#evidence-${id}`).evaluate(item => getComputedStyle(item).outlineStyle), /solid/);
    await page.keyboard.press('Tab');
    assert.equal(await page.evaluate(() => document.activeElement.textContent), 'Return to measurement');
    await page.keyboard.press('Enter');
    assert.equal(await source.evaluate(button => button === document.activeElement), true);
  }
  await page.locator('.technical > summary').press('Enter');
  assert.equal(await page.locator('.technical').getAttribute('open'), null);
});

test('browser-enforced clipboard policy exposes the manual-copy fallback', async t => {
  const { page } = await review(t);
  await page.goto(`${origin}/?block-clipboard`);
  await page.getByLabel('Target URL', { exact: true }).fill('https://fixture.example/');
  await page.getByRole('button', { name: 'Start diagnostic run', exact: true }).press('Enter');
  await page.locator('.run-heading').waitFor();
  assert.equal(await page.evaluate(() => (document.permissionsPolicy || document.featurePolicy).allowsFeature('clipboard-write')), false);
  await page.getByRole('button', { name: /^Copy selected IPv4 address/ }).press('Enter');
  await page.locator('.copy-feedback.error').waitFor();
  assert.equal(await page.getByRole('textbox', { name: 'IPv4 address for manual copying', exact: true }).inputValue(), '192.0.2.10');
});

test('input boundaries contrast with the form surface', async t => {
  const { page } = await review(t);
  const ratio = await page.locator('#target').evaluate(input => {
    const luminance = color => color.match(/[\d.]+/g).slice(0, 3)
      .map(Number).map(value => value / 255)
      .map(value => value <= 0.04045 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4)
      .reduce((sum, value, index) => sum + value * [0.2126, 0.7152, 0.0722][index], 0);
    const style = getComputedStyle(input);
    const values = [luminance(style.borderTopColor), luminance(style.backgroundColor)].sort((a, b) => b - a);
    return (values[0] + 0.05) / (values[1] + 0.05);
  });
  assert.ok(ratio >= 3, `Input boundary contrast was ${ratio.toFixed(2)}:1; expected at least 3:1.`);
});

test('the real browser clipboard receives the selected IPv4 and feedback resets', {
  skip: process.env.HOPVISTA_LIVE_CLIPBOARD !== '1' ? 'opt-in: writes the system clipboard' : false,
}, async t => {
  const { page, context } = await review(t);
  // Permission is granted to this disposable test context, not through a user
  // permission prompt. Never read pre-existing user clipboard content.
  await context.grantPermissions(['clipboard-read', 'clipboard-write']);
  await page.evaluate(() => navigator.clipboard.writeText('HopVista controlled clipboard check'));
  await page.getByRole('button', { name: 'Start diagnostic run', exact: true }).press('Enter');
  await page.locator('.run-heading').waitFor();
  await page.getByRole('button', { name: /^Copy selected IPv4 address/ }).press('Enter');
  await page.waitForFunction(() => document.querySelector('.copy-button')?.textContent === 'Copied');
  assert.equal(await page.evaluate(() => navigator.clipboard.readText()), '192.0.2.10');
  await page.waitForFunction(() => document.querySelector('.copy-button')?.textContent === 'Copy IPv4');
});

test('the selected IPv4 and its scope have a text separator for assistive reading', async t => {
  const { page } = await review(t);
  await page.getByRole('button', { name: 'Start diagnostic run', exact: true }).press('Enter');
  await page.locator('.run-heading').waitFor();
  assert.match(await page.locator('.selected').textContent(), /192\.0\.2\.10, address scope: special use/);
});

test('Start retains readable contrast when a diagnostic run makes it available again', async t => {
  const { page, run } = await review(t, 'running');
  const start = page.getByRole('button', { name: 'Start diagnostic run', exact: true });
  await start.press('Enter');
  await page.locator('.run-heading').waitFor();
  // Settle the disabled state, then sample the browser's actual transition on
  // re-enabling. Moving CSS animation time makes the failing frame repeatable.
  await start.evaluate(button => button.getAnimations().forEach(animation => animation.finish()));
  run.state = 'completed';
  await page.waitForFunction(() => !document.querySelector('button[type=submit]').disabled);
  const ratio = await start.evaluate(button => {
    for (const animation of button.getAnimations()) { animation.pause(); animation.currentTime = 40; }
    const luminance = color => color.match(/[\d.]+/g).slice(0, 3).map(Number)
      .map(value => value / 255)
      .map(value => value <= 0.04045 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4)
      .reduce((sum, value, index) => sum + value * [0.2126, 0.7152, 0.0722][index], 0);
    const style = getComputedStyle(button);
    const levels = [luminance(style.color), luminance(style.backgroundColor)].sort((a, b) => b - a);
    return (levels[0] + 0.05) / (levels[1] + 0.05);
  });
  assert.ok(ratio >= 4.5, `Enabled Start text contrast was ${ratio.toFixed(2)}:1 during its state change.`);
});

test('copy acknowledgement keeps its text readable through the color change', async t => {
  const { page } = await review(t);
  await page.evaluate(() => Object.defineProperty(navigator, 'clipboard', {
    configurable: true, value: { writeText: async () => {} },
  }));
  await page.getByRole('button', { name: 'Start diagnostic run', exact: true }).press('Enter');
  await page.locator('.run-heading').waitFor();
  const copy = page.getByRole('button', { name: /^Copy selected IPv4 address/ });
  await copy.press('Enter');
  await page.waitForFunction(() => document.querySelector('.copy-button')?.textContent === 'Copied');
  const colors = await copy.evaluate(button => {
    for (const animation of button.getAnimations()) { animation.pause(); animation.currentTime = 40; }
    const style = getComputedStyle(button);
    return { text: style.color, background: style.backgroundColor };
  });
  assert.deepEqual(colors, { text: 'rgb(255, 255, 255)', background: 'rgb(36, 79, 205)' });
});

test('website timing explains the four steps in plain language before opening evidence', async t => {
  const { page } = await review(t);
  await page.getByRole('button', { name: 'Start diagnostic run', exact: true }).press('Enter');
  await page.locator('.run-heading').waitFor();
  const introduction = await page.locator('.timing > .section-help').innerText();
  assert.match(introduction, /DNS finds an address/);
  assert.match(introduction, /TCP opens a connection/);
  assert.match(introduction, /TLS secures it/);
  assert.match(introduction, /HTTP first byte waits for the response to start/);
  assert.match(introduction, /do not measure a full page load/);
  assert.equal(await page.locator('.technical').getAttribute('open'), null);
});
