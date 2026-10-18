// Replay API-captured controlled results in the unchanged production UI.
// Generate captures using docs/implementation/issue-11.md or issue-12.md first.
import http from 'node:http';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const captureDir = process.env.HOPVISTA_JOURNEY_CAPTURE || path.join(root, '.tools', 'issue-11');
const modes = new Set(fs.readdirSync(captureDir).map(name => /^run-([a-z0-9]+)\.json$/.exec(name)?.[1]).filter(Boolean));
const instructions = [...modes].map(mode => `/${mode}`).join(', ');
const runs = new Map([...modes].map(mode => [mode, JSON.parse(fs.readFileSync(path.join(captureDir, `run-${mode}.json`), 'utf8'))]));
const server = http.createServer(async (req, res) => {
  const origin = `http://${server.address().address}:${server.address().port}`;
  if (req.headers.host !== new URL(origin).host) { res.writeHead(403).end(); return; }
  res.setHeader('Cache-Control', 'no-store');
  const json = value => { res.setHeader('Content-Type', 'application/json'); res.end(JSON.stringify(value)); };
  const pathname = new URL(req.url, origin).pathname;
  if (req.method === 'GET' && pathname === '/api/v1/session') return json({mutation_token: 'controlled-replay'});
  if (req.method === 'POST' && pathname === '/api/v1/runs') {
    if (req.headers.origin !== origin || req.headers['x-hopvista-token'] !== 'controlled-replay') { res.writeHead(403).end(); return; }
    let body = '';
    for await (const chunk of req) {
      body += chunk;
      if (body.length > 4096) { res.writeHead(413).end(); return; }
    }
    let mode;
    try { mode = new URL(JSON.parse(body).target_url).pathname.slice(1); } catch { /* rejected below */ }
    if (!modes.has(mode)) { res.statusCode = 400; return json({error:{message:`Use https://fixture.example with a captured path: ${instructions}.`}}); }
    res.statusCode = 202;
    return json({run_id: mode});
  }
  const match = pathname.match(/^\/api\/v1\/runs\/([a-z0-9]+)$/);
  if (req.method === 'GET' && match && modes.has(match[1])) return json(runs.get(match[1]));
  const asset = pathname === '/' ? 'index.html' : pathname.slice(1);
  if (req.method !== 'GET' || !['index.html', 'app.js', 'style.css'].includes(asset)) { res.writeHead(404).end(); return; }
  let body = fs.readFileSync(path.join(root, 'internal', 'localapi', 'static', asset));
  if (asset === 'index.html') {
    body = body.toString().replace('<body>', `<body><aside style="padding:8px 20px;background:#f5eddf;color:#745022;font:14px system-ui">Controlled API replay - simulated DNS/socket addresses, failures, ICMP and rule timings - real engine/API results - enter https://fixture.example with ${instructions}</aside>`);
  }
  res.setHeader('Content-Type', asset.endsWith('.js') ? 'text/javascript' : asset.endsWith('.css') ? 'text/css' : 'text/html');
  res.end(body);
});
server.listen(0, '127.0.0.1', () => console.log(`Controlled HTTPS replay: http://127.0.0.1:${server.address().port}/`));

