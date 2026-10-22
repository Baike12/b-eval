// Copy the built frontend into the directory the Go binary embeds.
//
// The embed directive cannot reach a parent directory, so the bundle has to
// be copied into internal/web/dist before the backend is compiled. Doing it
// here means `npm run build` alone produces a binary with a working UI.
import { cp, mkdir, rm, writeFile } from 'node:fs/promises'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const here = dirname(fileURLToPath(import.meta.url))
const source = resolve(here, '..', 'dist')
const target = resolve(here, '..', '..', 'internal', 'web', 'dist')

await rm(target, { recursive: true, force: true })
await mkdir(target, { recursive: true })
await cp(source, target, { recursive: true })
// The placeholder is restored because the embed directive needs the
// directory to hold at least one file. Without it a fresh checkout cannot
// compile the backend until the frontend has been built once.
await writeFile(resolve(target, '.gitkeep'), '')

console.log(`copied the frontend bundle into ${target}`)
