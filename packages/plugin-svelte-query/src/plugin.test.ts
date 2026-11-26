import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { adapterOas } from '@kubb/adapter-oas'
import { createKubb } from 'kubb'
import { pluginFetch } from '@kubb/plugin-fetch'
import { pluginTs } from '@kubb/plugin-ts'
import { afterAll, beforeAll, describe, expect, test } from 'bun:test'
import ts from 'typescript'
import { pluginSvelteQuery } from './index.ts'

const packageRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const webNodeModules = path.resolve(packageRoot, '../../web/node_modules')
const svelteQueryDir = path.resolve(webNodeModules, '@tanstack/svelte-query')
const svelteQueryTypesPath = path.resolve(svelteQueryDir, 'dist/index.d.ts')

const testSpec = {
  openapi: '3.0.3',
  info: { title: 'Test API', version: '1.0.0' },
  components: {
    schemas: {
      User: {
        type: 'object',
        required: ['id', 'name'],
        properties: {
          id: { type: 'string' },
          name: { type: 'string' },
          parent: { $ref: '#/components/schemas/User' },
          children: { type: 'array', items: { $ref: '#/components/schemas/User' } },
        },
      },
    },
  },
  paths: {
    '/me': {
      get: {
        operationId: 'getMe',
        responses: {
          '200': {
            description: 'Current user',
            content: { 'application/json': { schema: { $ref: '#/components/schemas/User' } } },
          },
        },
      },
    },
    '/items': {
      get: {
        operationId: 'listItems',
        parameters: [
          { name: 'tag', in: 'query', required: false, schema: { type: 'string' } },
          { name: 'cursor', in: 'query', required: false, schema: { type: 'number' } },
        ],
        responses: {
          '200': {
            description: 'Items list',
            content: {
              'application/json': {
                schema: {
                  type: 'object',
                  required: ['items'],
                  properties: {
                    items: { type: 'array', items: { $ref: '#/components/schemas/User' } },
                    nextCursor: { type: 'number' },
                  },
                },
              },
            },
          },
        },
      },
      post: {
        operationId: 'createItem',
        requestBody: {
          required: true,
          content: {
            'application/json': {
              schema: {
                type: 'object',
                required: ['name'],
                properties: {
                  name: { type: 'string' },
                  count: { type: 'number' },
                },
              },
            },
          },
        },
        responses: {
          '201': {
            description: 'Created',
            content: {
              'application/json': {
                schema: {
                  type: 'object',
                  required: ['id'],
                  properties: { id: { type: 'string' } },
                },
              },
            },
          },
        },
      },
    },
    '/items/{id}': {
      get: {
        operationId: 'getItemById',
        parameters: [
          { name: 'id', in: 'path', required: true, schema: { type: 'string' } },
          { name: 'include', in: 'query', required: false, schema: { type: 'string' } },
          { name: 'x-trace', in: 'header', required: false, schema: { type: 'string' } },
        ],
        responses: {
          '200': {
            description: 'Item',
            headers: {
              'x-request-id': { schema: { type: 'string' } },
            },
            content: {
              'application/json': { schema: { $ref: '#/components/schemas/User' } },
            },
          },
        },
      },
    },
    '/ping': {
      post: {
        operationId: 'pingServer',
        responses: {
          '204': { description: 'No content' },
        },
      },
    },
    '/upload': {
      post: {
        operationId: 'uploadData',
        requestBody: {
          required: true,
          content: {
            'application/json': {
              schema: { type: 'object', required: ['payload'], properties: { payload: { type: 'string' } } },
            },
            'text/plain': {
              schema: { type: 'string' },
            },
          },
        },
        responses: {
          '200': {
            description: 'Uploaded',
            content: { 'application/json': { schema: { type: 'object', properties: { ok: { type: 'boolean' } } } } },
          },
        },
      },
    },
  },
}

function typecheckFiles(files: string[]): string[] {
  const compilerOptions: ts.CompilerOptions = {
    target: ts.ScriptTarget.ESNext,
    module: ts.ModuleKind.ESNext,
    moduleResolution: ts.ModuleResolutionKind.Bundler,
    strict: true,
    skipLibCheck: true,
    noEmit: true,
    allowImportingTsExtensions: true,
    paths: {
      '@tanstack/svelte-query': [svelteQueryTypesPath],
    },
  }
  const program = ts.createProgram(files, compilerOptions)
  const diagnostics = ts.getPreEmitDiagnostics(program)
  return diagnostics.map((d) => {
    const message = ts.flattenDiagnosticMessageText(d.messageText, '\n')
    if (d.file && d.start !== undefined) {
      const { line, character } = d.file.getLineAndCharacterOfPosition(d.start)
      return `${d.file.fileName}:${line + 1}:${character + 1} TS${d.code}: ${message}`
    }
    return `TS${d.code}: ${message}`
  })
}

async function runGenerate(root: string, plugins: any[], spec: object = testSpec) {
  const kubb = createKubb({
    root,
    input: JSON.stringify(spec),
    output: { path: './generated', clean: true },
    adapter: adapterOas(),
    plugins,
  })
  return kubb.generate({ lockfile: false })
}

let tempRoot: string
let fullOutRoot: string

beforeAll(async () => {
  tempRoot = fs.mkdtempSync(path.join(os.tmpdir(), 'plugin-svelte-query-test-'))
  fs.mkdirSync(path.join(tempRoot, 'node_modules/@tanstack'), { recursive: true })
  fs.symlinkSync(svelteQueryDir, path.join(tempRoot, 'node_modules/@tanstack/svelte-query'), 'dir')
  fs.symlinkSync(path.join(webNodeModules, 'svelte'), path.join(tempRoot, 'node_modules/svelte'), 'dir')
  fullOutRoot = path.join(tempRoot, 'main')
  fs.mkdirSync(fullOutRoot, { recursive: true })

  const result = await runGenerate(fullOutRoot, [
    pluginTs(),
    pluginFetch(),
    pluginSvelteQuery({
      hooks: true,
      infinite: { queryParam: 'cursor', nextParam: 'nextCursor', initialPageParam: 0 },
    }),
  ])
  if (!result.success) {
    throw new Error(`Generation failed: ${JSON.stringify(result, null, 2)}`)
  }
})

afterAll(() => {
  if (tempRoot && fs.existsSync(tempRoot)) {
    fs.rmSync(tempRoot, { recursive: true, force: true })
  }
})

describe('Type-level checks', () => {
  test('Should typecheck all generated files without errors or TS2589 recursion on recursive OpenAPI schemas', () => {
    const queriesDir = path.join(fullOutRoot, 'generated/queries')
    const queryFiles = fs
      .readdirSync(queriesDir)
      .filter((f) => f.endsWith('.ts'))
      .map((f) => path.join(queriesDir, f))

    expect(queryFiles.length).toBeGreaterThan(0)
    const errors = typecheckFiles(queryFiles)
    expect(errors).toEqual([])
  })

  test('Should infer TData from select and enforce argument order (params?, options?, queryClient?) on query hooks', () => {
    const usageFile = path.join(fullOutRoot, 'generated/usage-queries.test-d.ts')
    fs.writeFileSync(
      usageFile,
      `import type { QueryClient } from '@tanstack/svelte-query'
import { createGetMe, createListItems, createGetItemById } from './queries/index.ts'

declare const client: QueryClient

const me = createGetMe()
const _meName: string | undefined = me.data?.name

const meSelected = createGetMe(() => ({ query: { select: (u) => u.children?.length ?? 0 } }))
const _meCount: number | undefined = meSelected.data

const itemsDefault = createListItems()
const _itemsCount: number | undefined = itemsDefault.data?.items.length

const itemsFull = createListItems(
  () => ({ query: { tag: 'active' } }),
  () => ({ query: { enabled: true, select: (res) => res.items.map((i) => i.name) } }),
  () => client,
)
const _names: string[] | undefined = itemsFull.data

const itemById = createGetItemById(() => ({ path: { id: '42' }, query: { include: 'parent' } }))
const _itemTitle: string | undefined = itemById.data?.name

// @ts-expect-error required params must not be omitted
createGetItemById()

// @ts-expect-error path.id must be string, not number
createGetItemById(() => ({ path: { id: 123 } }))

// @ts-expect-error params must come before options
createListItems(() => ({ query: { enabled: true } }))

void [_meName, _meCount, _itemsCount, _names, _itemTitle]
`,
    )

    const errors = typecheckFiles([usageFile])
    expect(errors).toEqual([])
  })

  test('Should reject non-accessor QueryClient and invalid mutation payloads at compile time', () => {
    const usageFile = path.join(fullOutRoot, 'generated/usage-errors.test-d.ts')
    fs.writeFileSync(
      usageFile,
      `import type { QueryClient } from '@tanstack/svelte-query'
import { createGetMe, createCreateItem } from './queries/index.ts'

declare const client: QueryClient

// @ts-expect-error queryClient must be Accessor<QueryClient>
createGetMe(undefined, client)

const mutation = createCreateItem()
mutation.mutate({ body: { name: 'ok', count: 1 } })

// @ts-expect-error body.name is required and must be string
mutation.mutate({ body: { count: 1 } })

// @ts-expect-error mutateAsync rejects invalid body shape
mutation.mutateAsync({ body: { name: 123 } })
`,
    )

    const errors = typecheckFiles([usageFile])
    expect(errors).toEqual([])
  })

  test('Should infer InfiniteData<T> on infinite queries and TContext from onMutate to onError/onSettled on mutations', () => {
    const usageFile = path.join(fullOutRoot, 'generated/usage-infinite-mutation.test-d.ts')
    fs.writeFileSync(
      usageFile,
      `import type { InfiniteData } from '@tanstack/svelte-query'
import { createListItemsInfinite, createCreateItem, createPingServer } from './queries/index.ts'

const inf = createListItemsInfinite(() => ({ query: { tag: 'v1' } }))
const _pages: InfiniteData<{ items: Array<unknown>; nextCursor?: number }> | undefined = inf.data

const infSelected = createListItemsInfinite(
  () => ({ query: { tag: 'v1' } }),
  () => ({ query: { select: (d) => d.pages.flatMap((p) => p.items) } }),
)
const _flatLen: number | undefined = infSelected.data?.length

const m = createCreateItem(() => ({
  mutation: {
    onMutate: () => ({ optimisticId: 'tmp-1' }),
    onError: (_err, _vars, ctx) => {
      const _id: string | undefined = ctx?.optimisticId
      void _id
    },
    onSettled: (_data, _err, _vars, ctx) => {
      const _id: string | undefined = ctx?.optimisticId
      void _id
    },
  },
}))
m.mutate({ body: { name: 'hello' } })

const ping = createPingServer()
ping.mutate(undefined)

void [_pages, _flatLen]
`,
    )

    const errors = typecheckFiles([usageFile])
    expect(errors).toEqual([])
  })

  test('Should work directly with createQuery and createMutation via exported *QueryOptions and *MutationOptions', async () => {
    const optionsOnlyRoot = path.join(tempRoot, 'options-only')
    await runGenerate(optionsOnlyRoot, [
      pluginTs(),
      pluginFetch(),
      pluginSvelteQuery({ hooks: false }),
    ])

    const usageFile = path.join(optionsOnlyRoot, 'generated/usage-direct.test-d.ts')
    fs.writeFileSync(
      usageFile,
      `import { createQuery, createMutation } from '@tanstack/svelte-query'
import { getMeQueryOptions, createItemMutationOptions } from './queries/index.ts'

const q = createQuery(() => getMeQueryOptions())
const _name: string | undefined = q.data?.name

const m = createMutation(() => createItemMutationOptions())
m.mutate({ body: { name: 'direct' } })

void _name
`,
    )

    const errors = typecheckFiles([usageFile])
    expect(errors).toEqual([])
  })

  test('Should expose full RequestResult types when returnType full is enabled', async () => {
    const fullResultRoot = path.join(tempRoot, 'full-result')
    await runGenerate(fullResultRoot, [
      pluginTs(),
      pluginFetch({ returnType: 'full' }),
      pluginSvelteQuery({ hooks: true, returnType: 'full' }),
    ])

    const usageFile = path.join(fullResultRoot, 'generated/usage-full.test-d.ts')
    fs.writeFileSync(
      usageFile,
      `import { createGetItemById } from './queries/index.ts'

const q = createGetItemById(() => ({ path: { id: '1' } }))
const _userName: string | undefined = q.data?.data.name
const _status: number | undefined = q.data?.response.status
const _reqId: string | null | undefined = q.data?.headers.get('x-request-id')

void [_userName, _status, _reqId]
`,
    )

    const errors = typecheckFiles([usageFile])
    expect(errors).toEqual([])
  })
})

describe('Runtime behavior of generated modules', () => {
  test('Should build distinct query, infinite query, and mutation keys and preserve falsy request bodies', async () => {
    const {
      getMeQueryKey,
      listItemsQueryKey,
      listItemsInfiniteQueryKey,
      getItemByIdQueryKey,
      createItemMutationKey,
    } = await import(path.join(fullOutRoot, 'generated/queries/index.ts'))

    expect(getMeQueryKey()).toEqual([{ url: '/me' }])
    expect(listItemsQueryKey()).toEqual([{ url: '/items' }])
    expect(listItemsQueryKey({ query: { tag: 'a' } })).toEqual([{ url: '/items' }, { tag: 'a' }])
    expect(listItemsInfiniteQueryKey({ query: { tag: 'a' } })).toEqual([{ url: '/items', infinite: true }, { tag: 'a' }])
    expect(getItemByIdQueryKey({ path: { id: '42' } })).toEqual([{ url: '/items/:id', params: { id: '42' } }])
    expect(createItemMutationKey()).toEqual([{ url: '/items' }])

    // Falsy body preservation check on a spec with primitive body
    const falsyRoot = path.join(tempRoot, 'falsy-body')
    await runGenerate(
      falsyRoot,
      [pluginTs(), pluginFetch(), pluginSvelteQuery()],
      {
        openapi: '3.0.3',
        info: { title: 'Falsy', version: '1.0.0' },
        paths: {
          '/search': {
            get: {
              operationId: 'searchWithBody',
              requestBody: {
                required: false,
                content: { 'application/json': { schema: { type: 'number' } } },
              },
              responses: { '200': { description: 'OK', content: { 'application/json': { schema: { type: 'string' } } } } },
            },
          },
        },
      },
    )
    const { searchWithBodyQueryKey } = await import(path.join(falsyRoot, 'generated/queries/createSearchWithBody.ts'))
    expect(searchWithBodyQueryKey({ body: 0 })).toEqual([{ url: '/search' }, 0])
    expect(searchWithBodyQueryKey({})).toEqual([{ url: '/search' }])
  })

  test('Should execute queryOptions().queryFn against mocked fetch with interpolated path, query params, headers, AbortSignal, and error propagation', async () => {
    const { getItemByIdQueryOptions } = await import(path.join(fullOutRoot, 'generated/queries/createGetItemById.ts'))

    const calls: Array<{ url: string; method: string; headers: Headers; signal?: AbortSignal | null }> = []
    const originalFetch = globalThis.fetch

    try {
      globalThis.fetch = (async (input: RequestInfo | URL, init?: RequestInit) => {
        const req = input instanceof Request ? input : new Request(input, init)
        calls.push({ url: req.url, method: req.method, headers: req.headers, signal: req.signal })
        if (req.url.includes('fail')) {
          return new Response(JSON.stringify({ message: 'Not Found' }), { status: 404, headers: { 'content-type': 'application/json' } })
        }
        return new Response(JSON.stringify({ id: '42', name: 'Alice' }), {
          status: 200,
          headers: { 'content-type': 'application/json' },
        })
      }) as typeof fetch

      const controller = new AbortController()
      const opts = getItemByIdQueryOptions(
        { path: { id: '42' }, query: { include: 'parent' }, headers: { 'x-trace': 'trace-1' } },
        { baseURL: 'http://localhost' },
      )
      const data = await opts.queryFn!({ queryKey: opts.queryKey, signal: controller.signal, meta: undefined } as any)

      expect(data).toEqual({ id: '42', name: 'Alice' })
      expect(calls[0]?.url).toContain('/items/42')
      expect(calls[0]?.url).toContain('include=parent')
      expect(calls[0]?.signal?.aborted).toBe(false)
      controller.abort()
      expect(calls[0]?.signal?.aborted).toBe(true)
      expect(calls[0]?.headers.get('x-trace')).toBe('trace-1')

      const failController = new AbortController()
      const failOpts = getItemByIdQueryOptions({ path: { id: 'fail' } }, { baseURL: 'http://localhost' })
      await expect(failOpts.queryFn!({ queryKey: failOpts.queryKey, signal: failController.signal, meta: undefined } as any)).rejects.toBeDefined()
    } finally {
      globalThis.fetch = originalFetch
    }
  })

  test('Should pass pageParam into infiniteQueryOptions().queryFn and compute next cursor via getNextPageParam', async () => {
    const { listItemsInfiniteQueryOptions } = await import(path.join(fullOutRoot, 'generated/queries/createListItemsInfinite.ts'))

    const requestedUrls: string[] = []
    const originalFetch = globalThis.fetch

    try {
      globalThis.fetch = (async (input: RequestInfo | URL) => {
        const url = typeof input === 'string' ? input : input instanceof URL ? input.toString() : input.url
        requestedUrls.push(url)
        return new Response(JSON.stringify({ items: [{ id: '1', name: 'One' }], nextCursor: 3 }), {
          status: 200,
          headers: { 'content-type': 'application/json' },
        })
      }) as typeof fetch

      const opts = listItemsInfiniteQueryOptions({ query: { tag: 'news' } }, { baseURL: 'http://localhost' })
      const page = await opts.queryFn!({ queryKey: opts.queryKey, pageParam: 2, direction: 'forward', signal: new AbortController().signal } as any)

      expect(requestedUrls[0]).toContain('tag=news')
      expect(requestedUrls[0]).toContain('cursor=2')
      expect(page).toEqual({ items: [{ id: '1', name: 'One' }], nextCursor: 3 })
      expect(opts.getNextPageParam({ items: [], nextCursor: 5 }, [], 0, [])).toBe(5)
      expect(opts.getNextPageParam({ items: [] }, [], 0, [])).toBeUndefined()
    } finally {
      globalThis.fetch = originalFetch
    }
  })

  test('Should execute mutationOptions().mutationFn for mutations with body, without body, and with multiple content types', async () => {
    const { createItemMutationOptions } = await import(path.join(fullOutRoot, 'generated/queries/createCreateItem.ts'))
    const { pingServerMutationOptions } = await import(path.join(fullOutRoot, 'generated/queries/createPingServer.ts'))
    const { uploadDataMutationOptions } = await import(path.join(fullOutRoot, 'generated/queries/createUploadData.ts'))

    const requests: Array<{ url: string; method: string; body: string; contentType: string | null }> = []
    const originalFetch = globalThis.fetch

    try {
      globalThis.fetch = (async (input: RequestInfo | URL, init?: RequestInit) => {
        const req = input instanceof Request ? input : new Request(input, init)
        const body = req.body ? await req.clone().text() : ''
        requests.push({
          url: req.url,
          method: req.method,
          body,
          contentType: req.headers.get('content-type'),
        })
        if (req.url.endsWith('/ping')) {
          return new Response(null, { status: 204 })
        }
        if (req.url.endsWith('/upload')) {
          return new Response(JSON.stringify({ ok: true }), { status: 200, headers: { 'content-type': 'application/json' } })
        }
        return new Response(JSON.stringify({ id: 'item-1' }), { status: 201, headers: { 'content-type': 'application/json' } })
      }) as typeof fetch

      const created = await createItemMutationOptions({ baseURL: 'http://localhost' }).mutationFn!({ body: { name: 'Widget', count: 3 } }, {} as any)
      expect(created).toEqual({ id: 'item-1' })
      expect(requests[0]?.method).toBe('POST')
      expect(requests[0]?.body).toBe(JSON.stringify({ name: 'Widget', count: 3 }))

      await pingServerMutationOptions({ baseURL: 'http://localhost' }).mutationFn!(undefined, {} as any)
      expect(requests[1]?.url).toContain('/ping')
      expect(requests[1]?.method).toBe('POST')

      const uploaded = await uploadDataMutationOptions({ baseURL: 'http://localhost' }).mutationFn!(
        { body: 'raw-text', contentType: 'text/plain' },
        {} as any,
      )
      expect(uploaded).toEqual({ ok: true })
      expect(requests[2]?.contentType).toContain('text/plain')
    } finally {
      globalThis.fetch = originalFetch
    }
  })

  test('Should keep params, options, and queryClient lazy inside the accessor passed to createQuery, createInfiniteQuery, and createMutation', async () => {
    const mockRoot = path.join(tempRoot, 'lazy-mock')
    const mockModulePath = path.join(mockRoot, 'mock-svelte-query.ts')
    fs.mkdirSync(mockRoot, { recursive: true })
    fs.writeFileSync(
      mockModulePath,
      `export const captured = { query: null as any, infinite: null as any, mutation: null as any }
export const queryOptions = (opts: any) => opts
export const infiniteQueryOptions = (opts: any) => opts
export const mutationOptions = (opts: any) => opts
export const createQuery = (optsFn: any, clientFn: any) => {
  captured.query = { optsFn, clientFn }
  return { optsFn, clientFn }
}
export const createInfiniteQuery = (optsFn: any, clientFn: any) => {
  captured.infinite = { optsFn, clientFn }
  return { optsFn, clientFn }
}
export const createMutation = (optsFn: any, clientFn: any) => {
  captured.mutation = { optsFn, clientFn }
  return { optsFn, clientFn }
}
`,
    )

    await runGenerate(mockRoot, [
      pluginTs(),
      pluginFetch(),
      pluginSvelteQuery({
        hooks: true,
        infinite: { queryParam: 'cursor', nextParam: 'nextCursor', initialPageParam: 0 },
        query: { importPath: mockModulePath },
        mutation: { importPath: mockModulePath },
      }),
    ])

    const { captured } = await import(mockModulePath)
    const { createListItems } = await import(path.join(mockRoot, 'generated/queries/createListItems.ts'))
    const { createListItemsInfinite } = await import(path.join(mockRoot, 'generated/queries/createListItemsInfinite.ts'))
    const { createCreateItem } = await import(path.join(mockRoot, 'generated/queries/createCreateItem.ts'))

    let tag = 'first'
    let enabled = false
    let paramsCalls = 0
    let optionsCalls = 0
    let clientCalls = 0
    const fakeClient = { id: 'custom-client' } as any

    createListItems(
      () => {
        paramsCalls++
        return { query: { tag } }
      },
      () => {
        optionsCalls++
        return { query: { enabled } }
      },
      () => {
        clientCalls++
        return fakeClient
      },
    )

    // None of the accessors are called synchronously during hook setup
    expect(paramsCalls).toBe(0)
    expect(optionsCalls).toBe(0)
    expect(clientCalls).toBe(0)

    const firstOpts = captured.query.optsFn()
    expect(firstOpts.queryKey).toEqual([{ url: '/items' }, { tag: 'first' }])
    expect(firstOpts.enabled).toBe(false)
    expect(captured.query.clientFn()).toBe(fakeClient)

    // Mutate reactive variables and re-evaluate accessor
    tag = 'second'
    enabled = true
    const secondOpts = captured.query.optsFn()
    expect(secondOpts.queryKey).toEqual([{ url: '/items' }, { tag: 'second' }])
    expect(secondOpts.enabled).toBe(true)

    createListItemsInfinite(() => ({ query: { tag } }))
    expect(captured.infinite.optsFn().queryKey).toEqual([{ url: '/items', infinite: true }, { tag: 'second' }])

    createCreateItem(() => ({ mutation: { mutationKey: ['custom-mut'] } }))
    expect(captured.mutation.optsFn().mutationKey).toEqual(['custom-mut'])
  })
})

describe('Plugin configuration behavior', () => {
  test('Should emit only options/key helpers in queries/ by default and add create* hooks only when hooks is true', async () => {
    const defaultRoot = path.join(tempRoot, 'cfg-default')
    await runGenerate(defaultRoot, [pluginTs(), pluginFetch(), pluginSvelteQuery()])

    const files = fs.readdirSync(path.join(defaultRoot, 'generated/queries')).sort()
    expect(files).toContain('createGetMe.ts')
    expect(files).toContain('createCreateItem.ts')

    const getMeContent = fs.readFileSync(path.join(defaultRoot, 'generated/queries/createGetMe.ts'), 'utf8')
    expect(getMeContent).toContain('export const getMeQueryKey')
    expect(getMeContent).toContain('export function getMeQueryOptions')
    expect(getMeContent).not.toContain('export function createGetMe')
    expect(getMeContent).not.toContain('createQuery')

    const hooksFiles = fs.readdirSync(path.join(fullOutRoot, 'generated/queries')).sort()
    expect(hooksFiles).toContain('createGetMe.ts')
    expect(hooksFiles).toContain('createCreateItem.ts')
    expect(fs.existsSync(path.join(fullOutRoot, 'generated/.kubb/withQueryKey.ts'))).toBe(false)
  })

  test('Should emit infinite query files only when both infinite and hooks are enabled and operation has queryParam', async () => {
    const infNoHooksRoot = path.join(tempRoot, 'cfg-inf-no-hooks')
    await runGenerate(infNoHooksRoot, [
      pluginTs(),
      pluginFetch(),
      pluginSvelteQuery({
        hooks: false,
        infinite: { queryParam: 'cursor', initialPageParam: 0 },
      }),
    ])
    const filesNoHooks = fs.readdirSync(path.join(infNoHooksRoot, 'generated/queries'))
    expect(filesNoHooks.some((f) => f.includes('Infinite'))).toBe(false)

    const filesWithHooks = fs.readdirSync(path.join(fullOutRoot, 'generated/queries'))
    expect(filesWithHooks).toContain('createListItemsInfinite.ts')
    expect(filesWithHooks).not.toContain('createGetMeInfinite.ts')
  })

  test('Should respect query false, mutation false, custom methods, importPath, key transformers, and resolver overrides', async () => {
    const customRoot = path.join(tempRoot, 'cfg-custom')
    await runGenerate(customRoot, [
      pluginTs(),
      pluginFetch(),
      pluginSvelteQuery({
        hooks: true,
        query: { methods: ['GET', 'POST'], importPath: 'custom-svelte-query' },
        mutation: false,
        queryKey: ({ node }) => [JSON.stringify(node.operationId)],
        resolver: {
          query: {
            name: (node) => `custom${node.operationId}`,
          },
        },
      }),
    ])

    const files = fs.readdirSync(path.join(customRoot, 'generated/queries'))
    // POST /ping was reclassified as query because query.methods includes POST and mutation is false
    expect(files).toContain('createPingServer.ts')
    const pingContent = fs.readFileSync(path.join(customRoot, 'generated/queries/createPingServer.ts'), 'utf8')
    expect(pingContent).toContain("from 'custom-svelte-query'")
    expect(pingContent).toContain('["pingServer"] as const')
    expect(pingContent).toContain('export function custompingServer')
  })

  test('Should fail with invalidPluginOptions when returnType full is combined with a client returning data', async () => {
    const invalidRoot = path.join(tempRoot, 'cfg-invalid')
    const kubb = createKubb({
      root: invalidRoot,
      input: JSON.stringify(testSpec),
      output: { path: './generated', clean: true },
      adapter: adapterOas(),
      plugins: [
        pluginTs(),
        pluginFetch({ returnType: 'data' }),
        pluginSvelteQuery({ returnType: 'full' }),
      ],
    })

    const output = await kubb.safeBuild()
    const errorDiagnostics = output.diagnostics.filter((d) => d.severity === 'error')
    expect(errorDiagnostics.length).toBeGreaterThan(0)
    expect(errorDiagnostics[0]?.message).toContain('returnType')
  })
})
