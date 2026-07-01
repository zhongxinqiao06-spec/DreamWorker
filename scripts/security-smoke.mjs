import { existsSync, readFileSync, readdirSync, statSync } from 'node:fs'
import { join } from 'node:path'

const failures = []

function fail(message) {
  failures.push(message)
}

function read(path) {
  return readFileSync(path, 'utf8')
}

function collectFiles(root, extensions, files = []) {
  if (!existsSync(root)) {
    return files
  }

  const stat = statSync(root)
  if (stat.isDirectory()) {
    for (const entry of readdirSync(root)) {
      collectFiles(join(root, entry), extensions, files)
    }
    return files
  }

  for (const extension of extensions) {
    if (root.endsWith(extension)) {
      files.push(root)
      break
    }
  }

  return files
}

function assertNoMatch(files, checks) {
  for (const file of files) {
    const content = read(file)
    for (const check of checks) {
      if (check.pattern.test(content)) {
        fail(`${file}: ${check.message}`)
      }
    }
  }
}

const rendererAndShared = [
  ...collectFiles('apps/desktop/renderer', ['.ts', '.vue', '.html', '.css']),
  ...collectFiles('apps/desktop/shared', ['.ts'])
]

assertNoMatch(rendererAndShared, [
  { pattern: /\bipcRenderer\b/, message: 'Renderer/shared must not reference ipcRenderer.' },
  { pattern: /\brequire\s*\(/, message: 'Renderer/shared must not call require().' },
  { pattern: /from\s+['"]node:/, message: 'Renderer/shared must not import node:* modules.' },
  {
    pattern: /import\s*\(\s*['"]node:/,
    message: 'Renderer/shared must not dynamically import node:* modules.'
  },
  { pattern: /\bfs\./, message: 'Renderer/shared must not access fs.' },
  { pattern: /\bprocess\./, message: 'Renderer/shared must not access process.' },
  {
    pattern: /\blocalStorage\b/,
    message: 'Renderer/shared must not persist workspace data in localStorage.'
  }
])

const preloadFiles = collectFiles('apps/desktop/preload', ['.ts'])
const preloadContent = preloadFiles.map((file) => read(file)).join('\n')
const exposedGlobals = [...preloadContent.matchAll(/exposeInMainWorld\(\s*['"]([^'"]+)['"]/g)].map(
  (match) => match[1]
)

if (exposedGlobals.length !== 1 || exposedGlobals[0] !== 'dreamworker') {
  fail(`Preload must expose only window.dreamworker, got: ${exposedGlobals.join(', ')}`)
}

assertNoMatch(preloadFiles, [
  {
    pattern: /exposeInMainWorld\(\s*['"]ipcRenderer['"]/,
    message: 'Preload must not expose raw IPC.'
  },
  {
    pattern: /\bipcRenderer\s*:/,
    message: 'Preload must not place ipcRenderer on exposed objects.'
  },
  { pattern: /\bprocess\s*:/, message: 'Preload must not expose process.' },
  { pattern: /\brequire\s*:/, message: 'Preload must not expose require.' },
  { pattern: /\bfs\s*:/, message: 'Preload must not expose fs.' }
])

const windowOptions = read('apps/desktop/main/window-options.ts')
for (const required of ['contextIsolation: true', 'nodeIntegration: false', 'sandbox: true']) {
  if (!windowOptions.includes(required)) {
    fail(`BrowserWindow security option missing: ${required}`)
  }
}

const rendererHtml = read('apps/desktop/renderer/index.html')
for (const required of [
  'Content-Security-Policy',
  "default-src 'self'",
  "script-src 'self'",
  "connect-src 'self'"
]) {
  if (!rendererHtml.includes(required)) {
    fail(`Renderer CSP requirement missing: ${required}`)
  }
}

const readme = read('README.md')
if (!readme.includes('UI 层所有面向用户可见的文字必须使用中文')) {
  fail('README must keep the UI Chinese copy rule.')
}

const rendererCopy = [
  read('apps/desktop/renderer/src/App.vue'),
  read('apps/desktop/renderer/src/stores/app-shell.ts'),
  ...collectFiles('apps/desktop/renderer/src/components', ['.vue']).map((file) => read(file))
].join('\n')
const oldEnglishPlaceholders = [
  'Idea Chat',
  'Incubator Workspace',
  'Shell State',
  'UI only',
  'not connected',
  'repo-bootstrap',
  'Mission intake',
  'Stage overview',
  'Event stream',
  'Deliverable'
]

for (const text of oldEnglishPlaceholders) {
  if (rendererCopy.includes(text)) {
    fail(`Renderer UI still contains old English placeholder: ${text}`)
  }
}

if (failures.length > 0) {
  console.error('Security smoke failed:')
  for (const failure of failures) {
    console.error(`- ${failure}`)
  }
  process.exit(1)
}

console.log('Security smoke passed.')
