import { test, expect } from 'e2e';
import { readFile, writeFile, stat, access, mkdir, unlink } from 'node:fs/promises';
import { join } from 'node:path';
import { workspace, receiver, fixtureKey } from './support';

let w: Awaited<ReturnType<typeof workspace>>;
test.beforeEach(async () => { w = await workspace(); });
test.afterEach(async () => { await w?.close(); });
async function success(args: string[], options = {}) {
  const r = await w.run(args, options);
  expect(r.code, r.stderr).toBe(0);
  expect(r.stderr).toBe('');
  return r;
}
async function json(args: string[], options = {}) { return JSON.parse((await success(['--json', ...args], options)).stdout); }
async function failure(args: string[], message: string, options = {}) {
  const r = await w.run(args, options);
  expect(r.code, r.stdout + r.stderr).toBe(1);
  expect(r.stdout).toBe('');
  expect(r.stderr).toContain(message);
  return r;
}
function last() { return w.requests.at(-1)!; }
function variables(expected: any) { expect(last().variables ?? {}).toEqual(expected); expect(last().auth).toBe(fixtureKey); expect(last().method).toBe('POST'); expect(last().path).toBe('/graphql'); }
function mutation(field: string, expected: any) { variables(expected); expect(last().query.trim().startsWith('mutation')).toBe(true); expect(last().query).toContain(field); }
async function absent(path: string) { expect(await access(path).then(() => false, () => true)).toBe(true); }

// Native tests own schema validity and detailed health rules. These flows own the
// compiled executable, flag/config wiring, real process exits and request boundaries.
test('configuration precedence selects the correct endpoint and authentication header', async () => {
  expect((await json(['info'])).os.hostname).toBe('fixture-server');
  variables({});
  const custom = join(w.dir, 'custom.yaml');
  await writeFile(custom, `server: ${w.baseURL}\napi_key: synthetic-file-key\n`);
  await json(['--config', custom, 'info']);
  expect(last().auth).toBe('synthetic-file-key');
  const env = { UNRAID_SERVER: w.baseURL, UNRAID_API_KEY: 'synthetic-env-key' };
  await writeFile(custom, 'server: http://127.0.0.1:1\napi_key: synthetic-file-key\n');
  await json(['--config', custom, 'info'], { env });
  expect(last().auth).toBe('synthetic-env-key');
  await json(['--config', custom, '--server', w.baseURL, '--api-key', fixtureKey, 'info'], { env: { ...env, UNRAID_SERVER: 'http://127.0.0.1:1' } });
  variables({});
  const count = w.requests.length;
  await writeFile(custom, 'server: [broken');
  await failure(['--config', custom, 'info'], 'parse config');
  await writeFile(custom, `server: ${w.baseURL}\n`);
  await failure(['--config', custom, 'info'], 'API key is required');
  await writeFile(custom, `api_key: ${fixtureKey}\n`);
  await failure(['--config', custom, 'info'], 'server URL is required');
  expect(w.requests.length).toBe(count);
  expect((await success(['version'])).stdout).toMatch(/^unraidctl version \d+\.\d+\.\d+/);
});

test('configure honors the selected file, preserves defaults, and writes private credentials', async () => {
  const path = join(w.dir, 'nested', 'chosen.yaml');
  const original = await readFile(w.config, 'utf8');
  const r = await success(['--config', path, 'configure'], { stdin: `${w.baseURL}\nsynthetic-created-key\n` });
  expect(r.stdout).toContain(path);
  expect(r.stdout).not.toContain('synthetic-created-key');
  expect(await readFile(w.config, 'utf8')).toBe(original);
  expect(await readFile(path, 'utf8')).toContain('api_key: synthetic-created-key');
  expect((await stat(path)).mode & 0o777).toBe(0o600);
  expect((await stat(join(w.dir, 'nested'))).mode & 0o777).toBe(0o700);
  await success(['--config', path, 'configure'], { stdin: '\n\n' });
  await json(['--config', path, 'info']);
  expect(last().auth).toBe('synthetic-created-key');
  expect(w.requests.length).toBe(1);
  await success(['--server', w.baseURL, '--api-key', 'synthetic-flag-key', '--config', path, 'configure'], { stdin: '\n\n' });
  expect(await readFile(path, 'utf8')).toContain('api_key: synthetic-flag-key');
});

test('configure rejects missing credentials without a file or a success claim', async () => {
  const path = join(w.dir, 'missing.yaml');
  await unlink(w.config);
  const r = await w.run(['--config', path, 'configure'], { stdin: `${w.baseURL}\n\n` });
  expect(r.code).toBe(1);
  expect(r.stderr).toContain('API key is required');
  expect(r.stdout).not.toContain('Configuration saved');
  await absent(path);
  expect(w.requests).toEqual([]);
});

test('configure reports a failed destination and preserves existing config bytes', async () => {
  const original = await readFile(w.config, 'utf8');
  const destination = join(w.dir, 'directory');
  await mkdir(destination);
  const r = await w.run(['--config', destination, 'configure'], { stdin: `${w.baseURL}\nsynthetic-created-key\n` });
  expect(r.code).toBe(1);
  expect(r.stderr).toContain('config');
  expect(r.stdout).not.toContain('Configuration saved');
  expect(await readFile(w.config, 'utf8')).toBe(original);
  expect(w.requests).toEqual([]);
});

test('array operations send exact mutation inputs and decode each returned field', async () => {
  expect((await json(['array', 'status'])).disks[0].id).toBe('disk:1');
  const human = (await success(['array', 'status'])).stdout;
  for (const value of ['data-one', 'parity-one', 'cache-one', 'flash-one']) expect(human).toContain(value);
  for (const [action, desired, state] of [['start', 'START', 'STARTED'], ['stop', 'STOP', 'STOPPED']]) {
    expect((await json(['array', action])).state).toBe(state);
    mutation('setState(', { input: { desiredState: desired } });
  }
  expect((await json(['array', 'add-disk', 'disk:9', '--slot', '2'])).state).toBe('ASSIGNED');
  mutation('addDiskToArray(', { input: { id: 'disk:9', slot: 2 } });
  await json(['array', 'add-disk', 'disk:9']);
  variables({ input: { id: 'disk:9' } });
  for (const [action, field, status] of [['mount-disk', 'mountArrayDisk(', 'MOUNTED'], ['unmount-disk', 'unmountArrayDisk(', 'UNMOUNTED']]) {
    expect((await json(['array', action, 'disk:1'])).status).toBe(status);
    mutation(field, { id: 'disk:1' });
  }
  expect(await json(['array', 'clear-stats', 'disk:1'])).toEqual({ cleared: true });
  mutation('clearArrayDiskStatistics(', { id: 'disk:1' });
  const count = w.requests.length;
  await failure(['array', 'remove-disk', 'disk:1', '--config', join(w.dir, 'absent.yaml')], 'Unraid WebGUI');
  expect(w.requests.length).toBe(count);
});

test('no-argument mutations reject stray operands before any request', async () => {
  for (const args of [['array', 'start', 'stray'], ['array', 'stop', 'stray'], ['docker', 'update-all', 'stray'], ['apikey', 'create', 'stray', '--name', 'fixture'], ['settings', 'update', 'stray', '--data', '{"fixture":true}']]) {
    await failure(args, 'unknown command');
  }
  expect(w.requests).toEqual([]);
});

test('docker reads preserve metadata, aliases, log filters, and empty or missing outcomes', async () => {
  expect((await json(['container', 'ls']))[0].id).toBe('container:1');
  const wide = (await success(['docker', 'list', '--wide'])).stdout;
  for (const value of ['fixture-app', 'available', 'WEB UI', '127.0.0.1:8080']) expect(wide).toContain(value);
  const inspected = await json(['docker', 'inspect', 'container:1']);
  expect(inspected.networkSettings).toEqual({ fixture: true });
  variables({ id: 'container:1' });
  expect((await success(['docker', 'inspect', 'container:1'])).stdout).toContain('Writable size');
  await failure(['docker', 'inspect', 'missing'], 'container not found');
  const logs = await json(['docker', 'logs', 'container:1', '--tail', '2', '--since', '2026-10-01T00:00:00Z']);
  expect(logs.cursor).toBe('fixture-cursor');
  expect(logs.lines[0].message).toBe('fixture log line');
  variables({ id: 'container:1', tail: 2, since: '2026-10-01T00:00:00Z' });
  expect((await success(['docker', 'logs', 'container:1'])).stdout).toContain('second line');
  w.state.data.docker.containers = [];
  expect(await json(['docker', 'list'])).toEqual([]);
  expect((await success(['docker', 'list'])).stdout).toContain('No containers found');
});

test('every docker action selects its own result and sends the chosen container', async () => {
  for (const [action, field] of [['start', 'start'], ['stop', 'stop'], ['restart', 'restart'], ['pause', 'pause'], ['unpause', 'unpause'], ['update', 'updateContainer']]) {
    expect((await json(['docker', action, 'container:1'])).state).toBe(`RESULT_${field}`);
    mutation(`${field}(`, { id: 'container:1' });
  }
  expect((await json(['docker', 'update-all']))[0].state).toBe('UPDATED_ALL');
  mutation('updateAllContainers', {});
  expect(await json(['docker', 'remove', 'container:1', '--with-image'])).toEqual({ removed: true });
  mutation('removeContainer(', { id: 'container:1', withImage: true });
  await json(['docker', 'remove', 'container:1']);
  variables({ id: 'container:1', withImage: false });
  expect(await json(['docker', 'autostart', 'container:1', '--enable', '--wait', '10'])).toEqual({ updated: true });
  mutation('updateAutostartConfiguration(', { entries: [{ id: 'container:1', autoStart: true, wait: 10 }], persist: true });
  await json(['docker', 'autostart', 'container:1', '--disable', '--persist=false']);
  variables({ entries: [{ id: 'container:1', autoStart: false }], persist: false });
  const human = (await success(['--no-color', 'docker', 'restart', 'container:1'])).stdout;
  expect(human).toContain('Container restart complete');
  expect(human).not.toContain('\u001b');
  w.state.mutationOK = false;
  const r = await w.run(['docker', 'remove', 'container:1']);
  expect(r.code).toBe(0);
  expect(r.stderr).toContain('was not removed');
  expect(r.stdout).not.toContain('Container removed');
  expect(await json(['docker', 'autostart', 'container:1', '--enable'])).toEqual({ updated: false });
});

test('API key lifecycle preserves permission normalization, multiple IDs, and created key visibility', async () => {
  expect((await json(['api-key', 'ls']))[0].name).toBe('fixture-key');
  expect((await success(['apikey', 'list'])).stdout).not.toContain('synthetic-new-key');
  expect(await json(['apikey', 'roles'])).toEqual(['ADMIN', 'VIEWER']);
  expect((await json(['apikey', 'permissions'])).getAvailableAuthActions).toContain('UPDATE_ANY');
  const created = await json(['apikey', 'create', '--name', 'automation', '--description', 'fixture', '--role', 'viewer,admin', '--permission', 'docker:read_any,update_any', '--overwrite']);
  expect(created.key).toBe('synthetic-new-key');
  mutation('create(', { input: { name: 'automation', description: 'fixture', roles: ['VIEWER', 'ADMIN'], permissions: [{ resource: 'DOCKER', actions: ['READ_ANY', 'UPDATE_ANY'] }], overwrite: true } });
  expect((await json(['apikey', 'update', 'key:1', '--name', 'updated-key', '--role', 'viewer'])).name).toBe('updated-key');
  mutation('update(', { input: { id: 'key:1', name: 'updated-key', roles: ['VIEWER'] } });
  expect(await json(['apikey', 'delete', 'key:1', 'key:2'])).toEqual({ deleted: true });
  mutation('delete(', { input: { ids: ['key:1', 'key:2'] } });
  for (const [command, field] of [['add-role', 'addRole'], ['remove-role', 'removeRole']]) {
    expect(await json(['apikey', command, 'key:1', '--role', 'viewer'])).toEqual({ updated: true });
    mutation(`${field}(`, { input: { apiKeyId: 'key:1', role: 'VIEWER' } });
  }
  expect((await success(['apikey', 'create', '--name', 'fixture'])).stdout).toContain('Key:  synthetic-new-key');
});

test('settings read and file or inline updates preserve object values and restart warnings', async () => {
  expect((await json(['settings', 'show'])).settings.api.version).toBe('4.37.4');
  expect((await success(['settings', 'show', '--values'])).stdout).toContain('"fixture":true');
  const payload = { api: { sandbox: true, extraOrigins: ['http://fixture.invalid'] } };
  const path = join(w.dir, 'settings.json');
  await writeFile(path, JSON.stringify(payload));
  const updated = await json(['settings', 'update', '--file', path]);
  expect(updated.values).toEqual(payload);
  expect(updated.restartRequired).toBe(true);
  mutation('updateSettings(', { input: payload });
  const r = await success(['settings', 'update', '--data', JSON.stringify(payload)]);
  expect(r.stdout).toContain('Restart required: true');
  expect(r.stdout).toContain('fixture restart notice');
  expect(await readFile(path, 'utf8')).toBe(JSON.stringify(payload));
});

test('settings reject non-object JSON without a mutation', async () => {
  for (const value of ['null', '[]', 'true', '"text"', '42']) await failure(['settings', 'update', '--data', value], 'JSON object');
  expect(w.requests).toEqual([]);
});

test('SSO commands preserve providers, configuration, and token validity', async () => {
  expect(await json(['sso', 'status'])).toEqual({ enabled: true, providers: 1, publicProviders: 1 });
  expect((await json(['sso', 'providers']))[0].clientId).toBe('fixture-client');
  expect((await json(['sso', 'public-providers']))[0].buttonText).toBe('Fixture login');
  expect((await json(['sso', 'config'])).defaultAllowedOrigins).toEqual(['http://fixture.invalid']);
  expect(await json(['sso', 'validate-token', 'synthetic-session'])).toEqual({ valid: true, username: 'fixture-user' });
  variables({ token: 'synthetic-session' });
  expect(await json(['sso', 'validate-token', 'synthetic-invalid'])).toEqual({ valid: false, username: '' });
  expect((await success(['sso', 'public-providers'])).stdout).toContain('Fixture login');
});

test('reporting commands deliver typed JSON and human output through the executable', async () => {
  const cases: [string[], (v: any) => any, any, string][] = [
    [['info'], v => v.cpu.cores, 8, 'Fixture CPU'],
    [['ups', 'status'], v => v[0].status, 'ONLINE', 'fixture-ups'],
    [['disk', 'ls'], v => v[0].smartStatus, 'PASSED', 'Fixture disk'],
    [['parity', 'status'], v => v.status, 'COMPLETED', 'COMPLETED'],
    [['parity', 'history'], v => v[0].errors, 0, '2026-10-01'],
    [['metrics', 'cpu'], v => v.percentTotal, 12.5, '12.5%'],
    [['metrics', 'memory'], v => v.available, 4294967296, '4.0 GiB'],
    [['metrics', 'temperature'], v => v.sensors[0].current.value, 42, 'cpu-package'],
    [['metrics', 'network'], v => v[0].receiveErrors, 2, '12.50%'],
    [['vm', 'ls'], v => v[0].name, 'fixture-vm', 'fixture-vm'],
    [['share', 'ls'], v => v[0].free, 3000000, 'fixture-share'],
    [['log', 'ls'], v => v[0].size, 2048, 'graphql-api.log'],
  ];
  for (const [args, select, expected, text] of cases) {
    expect(select(await json(args))).toEqual(expected);
    expect((await success(args)).stdout).toContain(text);
    expect(last().query.trim().startsWith('query')).toBe(true);
    variables({});
  }
  const log = await json(['logs', 'tail', '/var/log/graphql-api.log', '--lines', '2', '--start-line', '3']);
  expect(log.content).toBe('fixture log\nsecond line\n');
  variables({ path: '/var/log/graphql-api.log', lines: 2, startLine: 3 });
  const destination = join(w.dir, 'log.json');
  await success(['--json', 'log', 'show', '/var/log/graphql-api.log'], { output: destination });
  expect(JSON.parse(await readFile(destination, 'utf8')).totalLines).toBe(2);
  expect((await success(['log', 'show', '/var/log/graphql-api.log'])).stdout).toBe('fixture log\nsecond line\n');
});

test('notification pagination retrieves complete unread and archive sets with stable sorting', async () => {
  w.state.pages.UNREAD = Array.from({ length: 101 }, (_, i) => ({ id: `unread-${i}`, subject: `Unread ${i}`, importance: 'INFO', timestamp: i === 100 ? '2026-10-03T00:00:00Z' : '2026-10-01T00:00:00Z', type: 'UNREAD' }));
  w.state.pages.ARCHIVE = [{ id: 'archive-1', subject: 'Archived fixture', importance: 'WARNING', timestamp: '2026-10-02T00:00:00Z', type: 'ARCHIVE' }];
  const unread = await json(['notification', 'list']);
  expect(unread.length).toBe(101);
  expect(new Set(unread.map((n: any) => n.id)).size).toBe(101);
  expect(unread[0].id).toBe('unread-100');
  expect(w.requests.map(r => r.variables?.filter)).toEqual([{ type: 'UNREAD', offset: 0, limit: 100 }, { type: 'UNREAD', offset: 100, limit: 100 }]);
  w.requests.length = 0;
  const all = await json(['notifications', 'ls', '--all']);
  expect(all.length).toBe(102);
  expect(all.slice(0, 2).map((n: any) => n.id)).toEqual(['unread-100', 'archive-1']);
  expect(w.requests.map(r => r.variables?.filter.type)).toEqual(['UNREAD', 'UNREAD', 'ARCHIVE']);
  w.state.data.notifications = { overview: { unread: { total: 2, info: 0, warning: 1, alert: 1 }, archive: { total: 0, info: 0, warning: 0, alert: 0 } }, warningsAndAlerts: [{ id: 'alert-1', subject: 'Fixture alarm', importance: 'ALERT', type: 'UNREAD', timestamp: '2026-10-03T00:00:00Z' }] };
  expect((await json(['notify', 'alerts'])).warningsAndAlerts.length).toBe(1);
  expect((await success(['notification', 'alerts'])).stdout).toContain('Fixture alarm');
  expect(w.requests.every(r => !r.query.trim().startsWith('mutation'))).toBe(true);
  w.state.pageErrorOffset = 100;
  await failure(['notification', 'list', '--json'], 'offset 100');
});

test('health preserves available data and process exit status for healthy, alarm, and incomplete reports', async () => {
  expect((await json(['health'])).status).toBe('OK');
  expect(w.requests.length).toBe(7);
  w.state.data.upsDevices[0].status = 'ONLINE REPLACEBATT';
  let r = await w.run(['health', '--json']);
  expect(r.code).toBe(1);
  let report = JSON.parse(r.stdout);
  expect(report.status).toBe('ATTENTION');
  expect(report.checks.find((c: any) => c.component === 'UPS: fixture-ups').status).toBe('ALERT');
  w.state.denied = 'upsDevices';
  r = await w.run(['health', '--json']);
  expect(r.code).toBe(1);
  report = JSON.parse(r.stdout);
  expect(report.status).toBe('INCOMPLETE');
  expect(report.info.os.hostname).toBe('fixture-server');
  expect(report.errors[0]).toContain('UPS:');
  expect(report.errors[0]).not.toContain(fixtureKey);
  expect(w.requests.length).toBe(21);
  expect(w.requests.every(r => !r.query.trim().startsWith('mutation'))).toBe(true);
});

test('nullable or unavailable resources stay explicit rather than inventing data', async () => {
  for (const metric of ['cpu', 'memory', 'temperature']) {
    w.state.data.metrics[metric] = null;
    await failure(['metrics', metric, '--json'], 'unavailable');
  }
  w.state.data.upsDevices = [{ name: 'fixture-ups', status: 'COMMLOST', battery: { chargeLevel: null, estimatedRuntime: null }, power: { loadPercentage: null } }];
  expect((await success(['ups', 'status'])).stdout).toContain('unknown');
  w.state.data.disks[0].temperature = null;
  expect((await success(['disk', 'list'])).stdout).toContain('unknown');
  w.state.data.array.parityCheckStatus = null;
  expect(await json(['parity', 'status'])).toBe(null);
  await failure(['parity', 'status'], 'unavailable');
  w.state.data.parityHistory = [];
  expect((await success(['parity', 'history'])).stdout).toContain('No parity check history');
  w.state.denied = 'vms';
  await failure(['vm', 'list'], 'permission denied');
});

test('invalid options, missing operands, and conflicting inputs stop before any request', async () => {
  const cases: [string[], string][] = [
    [['not-a-command'], 'unknown command'], [['info', '--bogus'], 'unknown flag'],
    [['docker', 'start'], 'accepts 1 arg'], [['array', 'add-disk'], 'accepts 1 arg'], [['docker', 'inspect', 'a', 'b'], 'accepts 1 arg'],
    [['apikey', 'delete'], 'requires at least 1 arg'], [['sso', 'validate-token'], 'accepts 1 arg'], [['log', 'show'], 'accepts 1 arg'],
    [['docker', 'autostart', 'container:1'], 'exactly one'], [['docker', 'autostart', 'container:1', '--enable', '--disable'], 'exactly one'],
    [['apikey', 'create'], '--name is required'], [['apikey', 'add-role', 'key:1'], '--role is required'], [['apikey', 'remove-role', 'key:1'], '--role is required'],
    [['apikey', 'create', '--name', 'fixture', '--permission', 'malformed'], 'RESOURCE:ACTION'],
    [['settings', 'update'], 'exactly one'], [['settings', 'update', '--data', '{}', '--file', 'missing'], 'exactly one'],
    [['settings', 'update', '--data', '{broken'], 'valid JSON'], [['settings', 'update', '--file', join(w.dir, 'missing')], 'read settings file'],
    [['notification', 'list', 'stray'], 'unknown command'], [['array', 'add-disk', 'disk:1', '--slot', 'bad'], 'invalid argument'],
  ];
  for (const [args, message] of cases) await failure(args, message);
  expect(w.requests).toEqual([]);
});

test('transport and GraphQL failures produce clear errors without disclosing the supplied key', async () => {
  for (const [mode, text] of [['http-error', 'status 401'], ['gql-error', 'permission denied'], ['invalid-json', 'parse response'], ['disconnect', 'request failed']]) {
    w.state.mode = mode;
    const r = await failure(['info', '--json'], text);
    expect(r.stderr).not.toContain(fixtureKey);
    if (mode === 'http-error' || mode === 'gql-error') expect(r.stderr).toContain('[REDACTED]');
  }
  w.state.mode = 'gql-error';
  for (const args of [['array', 'stop'], ['docker', 'restart', 'container:1'], ['apikey', 'delete', 'key:1'], ['settings', 'update', '--data', '{\"fixture\":true}']]) {
    await failure(args, 'permission denied');
    expect(last().query.trim().startsWith('mutation')).toBe(true);
  }
  w.state.mode = 'ok';
  await failure(['--server', 'http://127.0.0.1:1', 'info'], 'request failed');
  await failure(['--server', '://invalid', 'info'], 'failed to create request');
});

test('HTTP redirects cannot forward the API key to a different origin', async () => {
  const received: string[] = [];
  const other = await receiver((req, res) => { received.push(String(req.headers['x-api-key'] ?? '')); res.end(JSON.stringify({ data: { info: { os: { hostname: 'redirect-target' } } } })); });
  try {
    w.state.mode = 'redirect';
    w.state.redirect = `${other.url}/graphql`;
    const r = await w.run(['info']);
    expect(received).toEqual([]);
    expect(r.code).toBe(1);
    expect(r.stderr).toContain('redirect');
    expect(w.requests.length).toBe(1);
  } finally { await other.close(); }
  w.state.mode = 'same-origin';
  expect((await json(['info'])).os.hostname).toBe('fixture-server');
  expect(last().path).toBe('/redirected/graphql');
  expect(last().auth).toBe(fixtureKey);
  const local = w.baseURL.replace('127.0.0.1', 'localhost');
  w.state.mode = 'canonical-origin';
  w.state.redirect = `${local.replace('localhost', 'LOCALHOST')}/redirected/graphql`;
  expect((await json(['--server', local, 'info'])).os.hostname).toBe('fixture-server');
  expect(last().path).toBe('/redirected/graphql');
  expect(last().auth).toBe(fixtureKey);
  w.state.redirect = w.state.redirect.replace('http:', 'HTTP:');
  expect((await json(['--server', local, 'info'])).os.hostname).toBe('fixture-server');
  expect(last().auth).toBe(fixtureKey);
});

test('quiet text output and JSON pipes preserve their separate output contracts', async () => {
  expect((await success(['--quiet', 'docker', 'logs', 'container:1'])).stdout).toBe('');
  expect((await json(['--quiet', 'docker', 'logs', 'container:1'])).lines[0].message).toBe('fixture log line');
  const quiet = await success(['--quiet', 'array', 'start']);
  expect(quiet.stdout).not.toContain('Array state set');
  expect(quiet.stdout).toContain('data-one');
});

test('shell completion generation works before configuration and makes no server request', async () => {
  await unlink(w.config);
  for (const [shell, marker] of [['bash', '__start_unraidctl'], ['zsh', '#compdef unraidctl'], ['fish', 'complete -c unraidctl'], ['powershell', 'Register-ArgumentCompleter']]) {
    expect((await success(['completion', shell])).stdout).toContain(marker);
  }
  expect(w.requests).toEqual([]);
});
