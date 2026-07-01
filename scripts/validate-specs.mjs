import { readFileSync, readdirSync, statSync } from 'node:fs'
import { basename, join } from 'node:path'
import Ajv2020 from 'ajv/dist/2020.js'
import addFormats from 'ajv-formats'

const schemaNames = [
  'agent',
  'artifact',
  'approval',
  'blueprint',
  'capability',
  'error',
  'event',
  'incubation',
  'policy',
  'task'
]

const ajv = new Ajv2020({
  allErrors: true,
  strict: false
})
addFormats(ajv)

const schemas = new Map()
for (const name of schemaNames) {
  const path = join('specs', `${name}.schema.json`)
  const schema = readJson(path)
  const expectedId = `https://schemas.dreamworker.dev/${name}/v0.1/schema.json`

  if (schema.$id !== expectedId) {
    fail(`${path}: expected $id ${expectedId}, got ${schema.$id}`)
  }
  if (!ajv.validateSchema(schema)) {
    fail(`${path}: schema is invalid\n${ajv.errorsText(ajv.errors, { separator: '\n' })}`)
  }

  schemas.set(name, ajv.compile(schema))
}

for (const fixtureKind of ['valid', 'invalid']) {
  const fixtureDir = join('specs', 'fixtures', fixtureKind)
  const files = readdirSync(fixtureDir).filter((file) => file.endsWith('.json'))
  const seen = new Set(files.map((file) => basename(file, '.json')))

  for (const name of schemaNames) {
    if (!seen.has(name)) {
      fail(`${fixtureDir}: missing ${name}.json`)
    }
  }

  for (const file of files) {
    const name = basename(file, '.json')
    const validate = schemas.get(name)
    if (!validate) {
      fail(`${fixtureDir}/${file}: no matching schema`)
    }

    const fixture = readJson(join(fixtureDir, file))
    const isValid = validate(fixture)
    if (fixtureKind === 'valid' && !isValid) {
      fail(
        `${fixtureDir}/${file}: expected valid fixture\n${ajv.errorsText(validate.errors, {
          separator: '\n'
        })}`
      )
    }
    if (fixtureKind === 'invalid' && isValid) {
      fail(`${fixtureDir}/${file}: expected invalid fixture`)
    }
  }
}

console.log('Specs validation passed.')

function readJson(path) {
  if (!statSync(path).isFile()) {
    fail(`${path}: not a file`)
  }
  return JSON.parse(readFileSync(path, 'utf8'))
}

function fail(message) {
  console.error(message)
  process.exit(1)
}
