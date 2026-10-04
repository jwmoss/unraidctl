# Compiled CLI end-to-end tests

Run `npm ci --ignore-scripts`, then `npm run test:e2e` (or `make test-e2e`).
The Make target builds the real Go executable with the normal version flags.
TesterArmy `e2e` 0.17.0 runs 21 deterministic process tests. No model is required.

Each process has a temporary home, an isolated config, synthetic keys, and a
five-second timeout. The harness passes only PATH, HOME, NO_COLOR, and explicit
fixture overrides. It does not inherit Unraid credentials or the user's config.
All server URLs use local HTTP fixtures. No test contacts a real Unraid server.
The fixture rejects unknown HTTP routes and returns only the requested top-level
GraphQL resources. It records query documents, variables, methods, paths, and
headers. Parsing, command dispatch, rendering, and config writes use product code.
Output redirection uses an actual process stdout file descriptor.

## Command matrix

The suite invokes all **57 executable leaves**: 53 project commands and four
Cobra completion generators. Aliases below share their canonical command owner.
The `help` command and hidden Cobra completion protocol are not counted as
product flows. Help-only tests do not contribute to this coverage.

| Commands | Automated proof |
| --- | --- |
| `version` | Built version has a semantic version, without a server request. |
| `configure` | Custom/default config path; blank-value reuse; flag defaults; private new file and directory; missing-key rejection; failed destination; unchanged default file. |
| `completion bash`, `completion zsh`, `completion fish`, `completion powershell` | Real shell-script output without config or server requests. |
| `info` | Typed CPU/version/host output; human output; flags, environment, and file precedence; missing credentials; malformed config; transport errors; cross-port redirect refusal and working relative/case-normalized redirects. |
| `health` | Seven real reads; valid JSON and exit codes for OK, ATTENTION, and INCOMPLETE; available data survives resource denial; error-key redaction; no mutations. |
| `array status` | JSON and human output include data, parity, cache, and boot devices. |
| `array start`, `array stop` | Exact desired-state input; distinct returned state; stray operands cannot send a mutation; server denial cannot claim success. |
| `array add-disk` | Disk ID with explicit slot or omitted slot; missing ID and malformed numeric option make no request. |
| `array mount-disk`, `array unmount-disk` | Exact ID; distinct mutation field and returned mount state. |
| `array clear-stats` | Exact ID and boolean receipt. |
| `array remove-disk` | Nonzero WebGUI guidance without config, credentials, or a server request. |
| `docker list` (`container ls`) | Metadata, wide columns, JSON, and empty result. |
| `docker inspect` | Exact ID, nested metadata, human details, missing-container error, invalid argument count. |
| `docker logs` | Container ID, tail/since filters, timestamped/plain lines, cursor, quiet text, and quiet JSON. |
| `docker start`, `docker stop`, `docker restart`, `docker pause`, `docker unpause`, `docker update` | Exact ID and distinct response field for every action; real human success output; missing operand and denied mutation. |
| `docker update-all` | All-container response; no invented ID; stray operands cannot send a mutation. |
| `docker remove` | Explicit/default image-removal choice; false receipt warns without a success claim. |
| `docker autostart` | Enable/disable, wait, persistence/default omission, boolean receipt, mutually exclusive options. |
| `metrics cpu`, `metrics memory`, `metrics temperature`, `metrics network` | Typed JSON and human output; null CPU/memory/temperature produce explicit failures. |
| `ups status` | Raw status, reported measurements, nullable measurements, human uncertainty. |
| `disk list` (`ls`) | Byte-based size, SMART status, nullable temperature, JSON and text. |
| `parity status`, `parity history` | Status/history output; null status remains null in JSON and explicit in text; empty history. |
| `vm list` (`ls`) | Domain list, human output, resource-denial error. |
| `share list` (`ls`) | Share name/capacity JSON and text. |
| `notification list` (`notifications ls`) | 101 unread records across pages, archive inclusion, sorting, unique IDs, exact filters, page failure with no partial JSON success. |
| `notification alerts` (`notify alerts`) | Deduplicated warnings/alerts resource, counts, JSON and text; no dismiss mutation. |
| `apikey list`, `apikey roles`, `apikey permissions` | Key metadata, role list, permission metadata, available auth actions; normal list does not invent a key secret. |
| `apikey create` | Name/description/roles/permissions/overwrite inputs; normalization; created key returned; invalid or missing input; stray operands rejected. |
| `apikey update` | Exact ID and replacement name/roles; returned metadata. |
| `apikey delete` | Multiple IDs and boolean receipt; missing IDs and denied mutation. |
| `apikey add-role`, `apikey remove-role` | Exact ID and normalized role; distinct mutation fields; missing role rejected. |
| `log list` (`logs ls`) | Available log metadata, JSON and text. |
| `log show` (`logs tail`) | Exact path, line/start-line filters, raw bytes, JSON, actual stdout redirection artifact; missing operand. |
| `settings show` | API version, unified JSON, explicit raw-values display. |
| `settings update` | File/inline object input, exact mutation values, restart/warning receipt, unchanged input file; missing/conflicting/malformed/non-object inputs; denied mutation. |
| `sso status`, `sso providers`, `sso public-providers`, `sso config` | Provider count, private/public metadata, button text, configured origins. |
| `sso validate-token` | Exact token variable; valid/invalid receipt and username; missing token. |

## Test ownership and limits

Existing Go tests remain the primary owners for every GraphQL document's
compatibility with pinned API 4.37.3 and 4.37.4 schemas, detailed capacity units,
all UPS alarm combinations, all health rules, pagination boundaries, and client
context timeouts. A native HTTP-client redirect table owns implicit/explicit HTTP80
and HTTPS443 equivalence, zero-padded ports, scheme/host/port refusal, and the
ten-redirect limit. It replaces only the network RoundTripper and still invokes
Client.Query, HTTP.Client.Do, and the production callback. The compiled process
flow owns real localhost hostname/scheme case and forwarded-header behavior. These process tests own compiled dispatch, config/flag wiring,
real exit codes, valid stdout artifacts, credential handling, and mutation safety.
A representative health alarm proves the process exit contract; it does not
replay the native alarm table. API 4.37.4 fixture metadata is synthetic.

All mutation proof is fixture-only. It verifies request and result contracts.
It does not establish actual array state, container state, disk assignment,
credential permissions, settings persistence, or OIDC behavior on a live server.
Live permissions, enabled feature flags, TLS/proxy configuration, API version
changes, actual notifications, physical storage health, and VM Manager state
require separate authorized server checks. No live smoke or mutation runs here.
The native disabled-VM diagnostic remains the primary owner for that error.

The CLI provides human and JSON output. It has no CSV mode, notification file
output flag, notification mutation, VM mutation, disk-removal mutation, or dry-run
flag. Shell redirection covers stdout artifacts; `configure` owns config output,
and settings updates consume a real JSON file. Arbitrary JSON object settings
keys and server enum constraints remain server validation responsibilities.
Shell scripts are generated here; interactive installation and completion in a
user shell remain separate platform checks. No host restart or storage change
is needed to run this suite.
