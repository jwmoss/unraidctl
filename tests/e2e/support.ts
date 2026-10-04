import { createServer, type IncomingMessage, type ServerResponse } from 'node:http';
import { mkdtemp, mkdir, writeFile, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join, resolve } from 'node:path';
import { openSync, closeSync } from 'node:fs';
import { spawn } from 'node:child_process';

export const fixtureKey = 'synthetic-fixture-key';
export const disk = { id: 'disk:1', name: 'data-one', device: 'sda', size: 2147483648, status: 'DISK_OK', temp: 32, type: 'DATA', numErrors: 0 };
export const parity = { date: '2026-10-01', duration: 600, status: 'COMPLETED', errors: 0, progress: 100, running: false, paused: false, correcting: false };
export const array = { state: 'STARTED', capacity: { kilobytes: { total: '4000000000', used: '1000000000', free: '3000000000' } }, disks: [disk], parities: [{ ...disk, id: 'disk:2', name: 'parity-one', type: 'PARITY' }], caches: [{ ...disk, id: 'disk:3', name: 'cache-one', type: 'CACHE' }], boot: { ...disk, id: 'disk:4', name: 'flash-one', type: 'FLASH' }, parityCheckStatus: parity };
export const container = { id: 'container:1', names: ['/fixture-app'], state: 'RUNNING', status: 'Up (healthy)', image: 'fixture/app:stable', autoStart: true, autoStartOrder: 1, autoStartWait: 10, isUpdateAvailable: true, isRebuildReady: false, webUiUrl: 'http://fixture.invalid:8080', lanIpPorts: ['127.0.0.1:8080'], hostConfig: { networkMode: 'bridge' }, sizeRw: 1024, networkSettings: { fixture: true }, mounts: [{ destination: '/data' }] };
export const key = { id: 'key:1', name: 'fixture-key', description: 'automation', roles: ['VIEWER'], createdAt: '2026-10-01', permissions: [{ resource: 'DOCKER', actions: ['READ_ANY'] }] };
export const provider = { id: 'oidc:1', name: 'Fixture OIDC', clientId: 'fixture-client', issuer: 'https://identity.invalid', scopes: ['openid', 'profile'], buttonText: 'Fixture login', buttonVariant: 'primary' };
export const info = { versions: { core: { unraid: '7.3.2', api: '4.37.4', kernel: '6.18.38-Unraid' } }, os: { platform: 'linux', distro: 'Unraid OS', release: '7.3 x86_64', uptime: '2026-10-01T00:00:00Z', hostname: 'fixture-server' }, cpu: { manufacturer: 'Fixture', brand: 'Fixture CPU', cores: 8, speed: 3.4 } };
export const metrics = { cpu: { percentTotal: 12.5 }, memory: { total: 8589934592, available: 4294967296, active: 2147483648, percentTotal: 50 }, temperature: { sensors: [{ name: 'cpu-package', type: 'CPU', current: { value: 42, unit: 'CELSIUS', status: 'NORMAL' } }] }, network: [{ id: 'net:1', name: 'eth0', operstate: 'UP', bytesReceived: 2048, bytesSent: 1024, rxSec: 1024, txSec: 512, utilizationPercent: 12.5, receiveErrors: 2, transmitErrors: 1, receiveDropped: 3, transmitDropped: 4 }] };
export const overview = { unread: { total: 0, info: 0, warning: 0, alert: 0 }, archive: { total: 0, info: 0, warning: 0, alert: 0 } };

type Request = { query: string; variables?: Record<string, any>; auth: string | undefined; path: string; method: string | undefined };
export type State = { mode: string; data: Record<string, any>; pages: Record<string, any[]>; mutationOK: boolean; redirect: string; denied: string; pageErrorOffset?: number };

function response(query: string, variables: Record<string, any>, state: State) {
  if (query.trim().startsWith('mutation')) {
    const actions = ['setState', 'addDiskToArray', 'mountArrayDisk', 'unmountArrayDisk', 'clearArrayDiskStatistics'];
    for (const action of actions) if (new RegExp(`\\b${action}\\(`).test(query)) {
      const value = action === 'setState' ? { ...array, state: variables.input.desiredState === 'START' ? 'STARTED' : 'STOPPED' } : action === 'addDiskToArray' ? { ...array, state: 'ASSIGNED' } : action === 'clearArrayDiskStatistics' ? state.mutationOK : { ...disk, status: action === 'mountArrayDisk' ? 'MOUNTED' : 'UNMOUNTED' };
      return { array: { [action]: value } };
    }
    for (const action of ['start', 'stop', 'restart', 'pause', 'unpause', 'updateContainer', 'updateAllContainers', 'removeContainer', 'updateAutostartConfiguration']) if (new RegExp(`\\b${action}\\s*[({]`).test(query)) {
      const value = action === 'removeContainer' || action === 'updateAutostartConfiguration' ? state.mutationOK : action === 'updateAllContainers' ? [{ ...container, state: 'UPDATED_ALL' }] : { ...container, state: `RESULT_${action}` };
      return { docker: { [action]: value } };
    }
    for (const action of ['create', 'update', 'delete', 'addRole', 'removeRole']) if (query.includes(`${action}(`)) return { apiKey: { [action]: action === 'create' ? { ...key, key: 'synthetic-new-key' } : action === 'update' ? { ...key, name: 'updated-key' } : state.mutationOK } };
    if (query.includes('updateSettings(')) return { updateSettings: { restartRequired: true, warnings: ['fixture restart notice'], values: variables.input } };
    throw new Error('Unexpected mutation');
  }
  if (query.includes('list(filter:')) {
    const { type, offset, limit } = variables.filter;
    return { notifications: { overview, list: (state.pages[type] ?? []).slice(offset, offset + limit) } };
  }
  if (query.includes('container(id:')) return { docker: { container: variables.id === 'missing' ? null : container } };
  if (query.includes('logs(id:')) return { docker: { logs: { containerId: variables.id, cursor: 'fixture-cursor', lines: [{ timestamp: '2026-10-01T00:00:00Z', message: 'fixture log line' }, { message: 'second line' }] } } };
  if (query.includes('logFile(path:')) return { logFile: { path: variables.path, content: 'fixture log\nsecond line\n', totalLines: 2, startLine: variables.startLine ?? 1 } };
  if (query.includes('validateOidcSession(')) return { validateOidcSession: { valid: variables.token === 'synthetic-session', username: variables.token === 'synthetic-session' ? 'fixture-user' : '' } };
  // Return only requested top-level resources. Wrong CLI query wiring must fail.
  const tokens = query.slice(query.indexOf('{')).match(/\$?[A-Za-z_]\w*|[{}()]/g) ?? [];
  let depth = 0, argumentsDepth = 0;
  const data: Record<string, any> = {};
  for (const token of tokens) {
    if (token === '{') depth++;
    else if (token === '}') depth--;
    else if (token === '(') argumentsDepth++;
    else if (token === ')') argumentsDepth--;
    else if (depth === 1 && argumentsDepth === 0) {
      if (!(token in state.data)) throw new Error(`Unexpected fixture resource: ${token}`);
      data[token] = state.data[token];
    }
  }
  return data;
}

export async function workspace() {
  const dir = await mkdtemp(join(tmpdir(), 'unraidctl-e2e-'));
  const requests: Request[] = [];
  const state: State = {
    mode: 'ok', pages: {}, mutationOK: true, redirect: '', denied: '',
    data: { info, array, docker: { containers: [container] }, upsDevices: [{ name: 'fixture-ups', model: 'Fixture UPS', status: 'ONLINE', battery: { chargeLevel: 100, estimatedRuntime: 600 }, power: { loadPercentage: 25, currentPower: 100 } }], disks: [{ id: 'disk:1', device: 'sda', name: 'Fixture disk', type: 'HDD', size: 2199023255552, smartStatus: 'PASSED', temperature: 32 }], parityHistory: [parity], metrics, notifications: { overview, warningsAndAlerts: [] }, shares: [{ name: 'fixture-share', comment: 'Fixture content', free: 3000000, used: 1000000 }], vms: { id: 'vm-manager', domain: [{ name: 'fixture-vm', state: 'RUNNING' }] }, apiKeys: [key], apiKeyPossibleRoles: ['ADMIN', 'VIEWER'], apiKeyPossiblePermissions: key.permissions, getAvailableAuthActions: ['READ_ANY', 'UPDATE_ANY'], logFiles: [{ name: 'graphql-api.log', path: '/var/log/graphql-api.log', size: 2048, modifiedAt: '2026-10-01' }], isSSOEnabled: true, oidcProviders: [provider], publicOidcProviders: [provider], oidcConfiguration: { defaultAllowedOrigins: ['http://fixture.invalid'], providers: [provider] }, settings: { id: 'settings:1', api: { version: '4.37.4', sandbox: false, plugins: ['fixture-plugin'] }, sso: { id: 'sso:1', oidcProviders: [provider] }, unified: { values: { fixture: true } } } },
  };
  state.data = structuredClone(state.data);
  const server = createServer(async (req, res) => {
    try {
      let raw = '';
      for await (const chunk of req) raw += chunk;
      const body = JSON.parse(raw || '{}');
      requests.push({ ...body, auth: req.headers['x-api-key'] as string | undefined, method: req.method, path: req.url! });
      if (req.method !== 'POST' || !['/graphql', '/redirected/graphql'].includes(req.url!)) { res.writeHead(404); res.end('Unexpected fixture route'); return; }
      if (['same-origin', 'canonical-origin'].includes(state.mode) && req.url === '/graphql') { res.writeHead(307, { Location: state.mode === 'canonical-origin' ? state.redirect : '/redirected/graphql' }); res.end(); return; }
      if (state.pageErrorOffset !== undefined && body.variables?.filter?.offset === state.pageErrorOffset) { res.end(JSON.stringify({ errors: [{ message: 'fixture page unavailable' }] })); return; }
      if (state.mode === 'redirect') { res.writeHead(307, { Location: state.redirect }); res.end(); return; }
      if (state.mode === 'disconnect') { req.socket.destroy(); return; }
      res.setHeader('Content-Type', 'application/json');
      if (state.mode === 'http-error') { res.writeHead(401); res.end(`Unauthorized ${fixtureKey}`); return; }
      if (state.mode === 'invalid-json') { res.end('{broken'); return; }
      if (state.mode === 'gql-error' || (state.denied && body.query.includes(state.denied))) { res.end(JSON.stringify({ errors: [{ message: `permission denied ${fixtureKey}` }] })); return; }
      res.end(JSON.stringify({ data: response(body.query, body.variables ?? {}, state) }));
    } catch (err) { res.writeHead(500); res.end(String(err)); }
  });
  await new Promise<void>(resolve => server.listen(0, '127.0.0.1', resolve));
  const baseURL = `http://127.0.0.1:${(server.address() as any).port}`;
  const config = join(dir, '.config', 'unraidctl', 'config.yaml');
  await mkdir(join(dir, '.config', 'unraidctl'), { recursive: true, mode: 0o700 });
  await writeFile(config, `server: ${baseURL}\napi_key: ${fixtureKey}\n`, { mode: 0o600 });
  return {
    dir, baseURL, config, state, requests,
    run(args: string[], options: { env?: Record<string, string>; stdin?: string; output?: string } = {}) {
      return new Promise<{ code: number | null; stdout: string; stderr: string }>((resolveRun, reject) => {
        const outputFD = options.output ? openSync(options.output, 'w', 0o600) : undefined;
        const child = spawn(resolve('unraidctl'), args, { cwd: dir, env: { PATH: process.env.PATH, HOME: dir, NO_COLOR: '1', ...options.env }, stdio: ['pipe', outputFD ?? 'pipe', 'pipe'] });
        if (outputFD !== undefined) closeSync(outputFD);
        let stdout = '', stderr = '';
        const timer = setTimeout(() => { child.kill('SIGKILL'); reject(new Error(`CLI timed out: ${args.join(' ')}`)); }, 5000);
        child.stdout?.on('data', chunk => { stdout += chunk; });
        child.stderr.on('data', chunk => { stderr += chunk; });
        child.on('error', err => { clearTimeout(timer); reject(err); });
        child.on('close', code => {
          clearTimeout(timer);
          resolveRun({ code, stdout, stderr });
        });
        child.stdin.end(options.stdin ?? '');
      });
    },
    async close() { server.closeAllConnections(); await new Promise<void>(resolve => server.close(() => resolve())); await rm(dir, { recursive: true, force: true }); },
  };
}

export async function receiver(handler: (req: IncomingMessage, res: ServerResponse) => void) {
  const server = createServer(handler);
  await new Promise<void>(resolve => server.listen(0, '127.0.0.1', resolve));
  return { url: `http://127.0.0.1:${(server.address() as any).port}`, async close() { server.closeAllConnections(); await new Promise<void>(resolve => server.close(() => resolve())); } };
}
