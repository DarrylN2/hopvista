// Opt-in browser + packaged process seams (#24). No application API mocking.
import { test, before, after } from 'node:test';
import assert from 'node:assert/strict';
import http from 'node:http';
import fs from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import { spawn } from 'node:child_process';
import { once } from 'node:events';
import { createRequire } from 'node:module';
import { fileURLToPath } from 'node:url';
import { createHash } from 'node:crypto';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const source = process.env.HOPVISTA_PACKAGED_EXE;
const enabled = process.platform === 'win32' && process.arch === 'x64' && !!source;
const skip = enabled ? false : 'requires Windows x64 and HOPVISTA_PACKAGED_EXE';
let browser;

before(async () => {
  if (!enabled) return;
  assert.ok(path.isAbsolute(source), 'HOPVISTA_PACKAGED_EXE must be absolute');
  const image = await fs.readFile(source);
  assert.equal(image.readUInt16LE(0), 0x5a4d, 'expected a Windows executable');
  const peOffset = image.readUInt32LE(0x3c);
  assert.equal(image.readUInt32LE(peOffset), 0x4550, 'expected a PE signature');
  assert.equal(image.readUInt16LE(peOffset + 4), 0x8664, 'expected Windows x64');
  console.log(`Packaged executable SHA-256: ${createHash('sha256').update(image).digest('hex')}`);
  const { chromium } = createRequire(path.join(root, 'ui/package.json'))('playwright');
  browser = await chromium.launch({ channel: process.env.HOPVISTA_BROWSER || 'msedge', headless: true });
});
after(async () => { await browser?.close(); });

async function packagedProcess(t) {
  const directory = await fs.mkdtemp(path.join(os.tmpdir(), 'hopvista-ui-'));
  const processes = [];
  t.after(async () => {
    for (const process of processes) await process.stop();
    await fs.rm(directory, { recursive: true, force: true });
  });
  const executable = path.join(directory, 'hopvista.exe');
  await fs.copyFile(source, executable);
  return { launch: async () => {
    const process = await launch(t, executable);
    processes.push(process);
    return process;
  } };
}

async function launch(t, executable) {
  const env = Object.fromEntries(Object.entries(process.env).filter(([key]) => key.toLowerCase() !== 'path'));
  env.PATH = path.join(process.env.SystemRoot, 'System32');
  const child = spawn(executable, [], { cwd: path.dirname(executable), env, windowsHide: true, stdio: ['ignore', 'pipe', 'pipe'] });
  // Attach immediately so a startup failure is observed, and wait for complete
  // process/stream closure before deleting the isolated executable directory.
  const closed = once(child, 'close');
  closed.catch(() => {});
  let stopped = false;
  const stop = async () => {
    if (stopped) return;
    stopped = true;
    if (child.exitCode === null && child.signalCode === null) child.kill();
    await closed;
  };
  t.after(stop);
  child.stdout.resume();
  let output = '';
  const origin = await new Promise((resolve, reject) => {
    const timer = setTimeout(() => reject(new Error('packaged listener startup timed out')), 10_000);
    const fail = error => { clearTimeout(timer); reject(error); };
    child.once('error', fail);
    child.once('exit', () => fail(new Error('packaged process exited before listener startup')));
    child.stderr.on('data', data => {
      output += data.toString();
      const match = output.match(/HopVista: (http:\/\/127\.0\.0\.1:\d+)\//);
      if (match) { clearTimeout(timer); resolve(match[1]); }
    });
  }).catch(async error => { await stop(); throw error; });
  return { origin, stop };
}

async function targetServer(t, handler) {
  const server = http.createServer(handler);
  t.after(async () => {
    server.closeAllConnections();
    await new Promise((resolve, reject) => server.close(error => error ? reject(error) : resolve()));
  });
  server.listen(0, '127.0.0.1');
  await once(server, 'listening');
  return `http://127.0.0.1:${server.address().port}`;
}

async function workspace(t, origin, expectedStartRejections = 0) {
  const context = await browser.newContext({ viewport: { width: 1440, height: 1000 } });
  const page = await context.newPage();
  const errors = [], badResponses = [], failedRequests = [], requests = [];
  let rejectedStarts = 0;
  const origins = new Set([origin]);
  page.on('pageerror', error => errors.push(error.message));
  page.on('console', message => {
    if (!['error', 'warning'].includes(message.type())) return;
    // Chromium also logs expected HTTP rejections as resource errors.
    if (expectedStartRejections && message.location().url === `${origin}/api/v1/runs`
      && /^Failed to load resource:.*400/.test(message.text())) return;
    errors.push(message.text());
  });
  page.on('request', request => requests.push(request.url()));
  page.on('requestfailed', request => failedRequests.push(`${request.method()} ${request.url()}: ${request.failure()?.errorText}`));
  page.on('response', response => {
    if (response.status() < 400) return;
    if (response.url() === `${origin}/api/v1/runs` && response.request().method() === 'POST' && response.status() === 400) rejectedStarts++;
    else badResponses.push(`${response.status()} ${response.url()}`);
  });
  t.after(async () => {
    await context.close();
    assert.deepEqual(errors, [], 'browser errors');
    assert.deepEqual(badResponses, [], 'unexplained HTTP failures');
    assert.deepEqual(failedRequests, [], 'failed browser requests');
    assert.equal(rejectedStarts, expectedStartRejections, 'expected input rejections');
    assert.ok(requests.every(url => origins.has(new URL(url).origin)), 'UI requests must remain on the packaged listeners');
  });
  const open = async nextOrigin => {
    origins.add(nextOrigin);
    await page.goto(nextOrigin);
    await page.getByRole('button', { name: 'Start diagnostic run', exact: true }).waitFor();
    await page.waitForFunction(() => !document.querySelector('button[type=submit]').disabled);
  };
  await open(origin);
  return { page, context, open };
}

async function start(page, target, hops = 1) {
  await page.getByLabel('Target URL', { exact: true }).fill(target);
  await page.locator('.advanced > summary').click();
  await page.getByLabel('Maximum hops (1–64)', { exact: true }).fill(String(hops));
  await page.locator('.advanced > summary').click();
  const response = page.waitForResponse(r => r.url().endsWith('/api/v1/runs') && r.request().method() === 'POST');
  await page.getByRole('button', { name: 'Start diagnostic run', exact: true }).click();
  const accepted = await response;
  assert.equal(accepted.status(), 202);
  const body = await accepted.json();
  assert.equal(body.status_url, `/api/v1/runs/${body.run_id}`);
  return body;
}

async function terminalResult(page, context, origin, started, state = 'completed') {
  await page.waitForFunction(expected => document.querySelector('main')?.dataset.state === expected, state, { timeout: 25_000 });
  const response = await context.request.get(origin + started.status_url);
  assert.equal(response.status(), 200);
  const run = await response.json();
  assert.equal(run.schema_version, 1);
  assert.equal(run.state, state);
  assert.equal(run.run_id, started.run_id);
  // The user-visible export must be the actual terminal API result.
  if (await page.locator('.technical').getAttribute('open') === null) await page.getByText('Technical evidence', { exact: true }).click();
  await page.getByText('Full run result JSON', { exact: true }).click();
  assert.deepEqual(JSON.parse(await page.locator('.technical > details pre').innerText()), run);
  return run;
}

test('packaged UI submits a real redirected run and opens its evidence', { skip, timeout: 45_000 }, async t => {
  const target = await targetServer(t, (req, res) => {
    if (req.url === '/start') { res.writeHead(302, { Location: '/final' }); res.end(); }
    else { res.writeHead(503); res.end(); }
  });
  const packaged = await packagedProcess(t);
  const { origin } = await packaged.launch();
  const { page, context } = await workspace(t, origin);
  const started = await start(page, `${target}/start`, 4);
  const run = await terminalResult(page, context, origin, started);
  assert.equal(run.target_url, `${target}/start`);
  assert.equal(run.max_hops, 4);
  assert.match(await page.locator('.run-heading').innerText(), /Hop limit 4/);
  const family = run.families.find(f => f.family === 'ipv4');
  assert.deepEqual(family.exchanges.map(x => x.response_status), [302, 503]);
  for (const [index, exchange] of family.exchanges.entries()) {
    assert.equal(exchange.selected_ip, '127.0.0.1');
    assert.equal(exchange.actual_peer_ip, '127.0.0.1');
    const rendered = page.locator('.exchange').nth(index);
    assert.match(await rendered.innerText(), /IPv4 127\.0\.0\.1.*loopback.*Actual peer 127\.0\.0\.1/);
    assert.match(await rendered.locator('.exchange-footer').innerText(), new RegExp(`HTTP status ${exchange.response_status}`));
  }
  assert.ok(family.paths[0].trace.hops.length > 0, 'real native trace evidence');
  const finding = run.findings.find(f => f.family === 'ipv4');
  assert.ok(finding);
  assert.equal(await page.locator('.observation').innerText(), finding.observation);
  assert.ok((await page.locator('.limit').innerText()).includes(finding.limit));
  const id = family.exchanges[1].http.id;
  const link = page.getByRole('button', { name: `View evidence ${id}`, exact: true });
  await link.click();
  const evidence = page.locator(`[id="evidence-${id}"]`);
  assert.equal(await evidence.evaluate(element => element === document.activeElement), true);
  await evidence.getByText('Raw evidence', { exact: true }).click();
  assert.deepEqual(JSON.parse(await evidence.locator('pre').innerText()), family.exchanges[1].http);
  await evidence.getByRole('button', { name: 'Return to measurement', exact: true }).click();
  assert.equal(await link.evaluate(element => element === document.activeElement), true);
});

test('packaged UI recovers from rejected input and retains incomplete HTTP evidence', { skip, timeout: 45_000 }, async t => {
  const target = await targetServer(t, (req, res) => {
    if (req.url === '/broken') { res.writeHead(302, { Location: '/abort' }); res.end(); }
    else if (req.url === '/abort') req.socket.destroy();
    else { res.writeHead(204); res.end(); }
  });
  const packaged = await packagedProcess(t);
  const { origin } = await packaged.launch();
  const { page, context } = await workspace(t, origin, 1);
  await page.getByLabel('Target URL', { exact: true }).fill('http://[::1]/');
  const rejection = page.waitForResponse(r => r.url().endsWith('/api/v1/runs') && r.request().method() === 'POST');
  await page.getByRole('button', { name: 'Start diagnostic run', exact: true }).click();
  const rejected = await rejection;
  assert.equal(rejected.status(), 400);
  const error = (await rejected.json()).error;
  assert.equal(error.code, 'ipv6_not_supported_in_v1');
  await page.getByRole('alert').waitFor();
  assert.ok((await page.getByRole('alert').innerText()).includes(error.message));
  assert.equal(await page.getByRole('button', { name: 'Start diagnostic run', exact: true }).isEnabled(), true);

  const started = await start(page, `${target}/broken`);
  const run = await terminalResult(page, context, origin, started);
  const family = run.families.find(f => f.family === 'ipv4');
  assert.equal(family.outcome, 'error');
  assert.equal(family.exchanges.length, 2);
  assert.equal(family.exchanges[0].response_status, 302);
  const incomplete = family.exchanges[1];
  assert.equal(incomplete.tcp.outcome, 'completed');
  assert.equal(incomplete.http.outcome, 'error');
  assert.equal(incomplete.http.duration_ms, undefined);
  assert.equal(incomplete.response_status, undefined);
  assert.match(await page.locator('.run-description').innerText(), /Some measurements are incomplete/);
  assert.match(await page.locator('.exchange').first().innerText(), /HTTP status 302/);
  assert.equal(await page.locator('.exchange').nth(1).locator('[data-outcome=error]').innerText(), 'error');
  assert.equal(await page.locator('.observation').innerText(), run.findings.find(f => f.family === 'ipv4').observation);
  assert.equal(await page.getByRole('alert').count(), 0);

  const next = await start(page, `${target}/healthy`);
  const nextRun = await terminalResult(page, context, origin, next);
  assert.notEqual(nextRun.run_id, run.run_id);
  assert.equal(nextRun.families.find(f => f.family === 'ipv4').exchanges[0].response_status, 204);
  assert.equal(await page.locator('.exchange').count(), 1);
  assert.equal(await page.locator('.run-heading h2').innerText(), `${target}/healthy`);
});

test('packaged UI cancels held HTTP work, runs again, and starts fresh after process restart', { skip, timeout: 45_000 }, async t => {
  let receivedHold;
  const held = new Promise(resolve => { receivedHold = resolve; });
  const target = await targetServer(t, (req, res) => {
    if (req.url === '/cancel') { res.writeHead(302, { Location: '/hold' }); res.end(); }
    else if (req.url === '/hold') receivedHold(); // Intentionally leave response headers unwritten.
    else { res.writeHead(200); res.end(); }
  });
  const packaged = await packagedProcess(t);
  const first = await packaged.launch();
  const { page, context, open } = await workspace(t, first.origin);
  const started = await start(page, `${target}/cancel`);
  await held;
  await page.getByRole('button', { name: 'Cancel run', exact: true }).waitFor();
  const cancellation = page.waitForResponse(r => r.url().endsWith(`${started.status_url}/cancel`) && r.request().method() === 'POST');
  await page.getByRole('button', { name: 'Cancel run', exact: true }).click();
  const cancelledResponse = await cancellation;
  assert.equal(cancelledResponse.status(), 200);
  assert.ok(['stopping', 'cancelled'].includes((await cancelledResponse.json()).state));
  const cancelled = await terminalResult(page, context, first.origin, started, 'cancelled');
  const family = cancelled.families.find(f => f.family === 'ipv4');
  assert.equal(family.exchanges[0].response_status, 302);
  assert.equal(family.exchanges[0].http.outcome, 'completed');
  assert.equal(family.exchanges[1].http.outcome, 'cancelled');
  assert.equal(family.exchanges[1].http.duration_ms, undefined);
  assert.deepEqual(cancelled.findings, []);
  assert.match(await page.locator('.exchange').first().innerText(), /HTTP status 302/);
  assert.match(await page.locator('.answer').innerText(), /No observed finding/);

  const next = await start(page, `${target}/after-cancel`);
  const completed = await terminalResult(page, context, first.origin, next);
  assert.equal(completed.families.find(f => f.family === 'ipv4').exchanges[0].response_status, 200);
  await first.stop();
  const second = await packaged.launch();
  // The same disposable browser context navigates to the relaunched process.
  await open(second.origin);
  assert.equal(await page.locator('.run-heading').count(), 0);
  assert.equal(await page.getByLabel('Target URL', { exact: true }).inputValue(), '');
  const old = await context.request.get(second.origin + next.status_url);
  assert.equal(old.status(), 404);
  assert.equal((await old.json()).error.code, 'not_found');
  const fresh = await start(page, `${target}/after-restart`);
  const freshRun = await terminalResult(page, context, second.origin, fresh);
  assert.notEqual(freshRun.run_id, completed.run_id);
  assert.equal(freshRun.families.find(f => f.family === 'ipv4').exchanges[0].response_status, 200);
});
