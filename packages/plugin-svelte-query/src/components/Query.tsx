import { ast } from '@internals/ast'
import type { FunctionParameterNode, ResolverTs } from '@kubb/plugin-ts'
import { createFunctionParameter, createFunctionParameters, functionPrinter } from '@kubb/plugin-ts'
import { Function } from '@kubb/renderer-jsx'
import type { KubbReactNode } from '@kubb/renderer-jsx'
import { buildResponseTypes, resolvePageParamType } from '@internals/tanstack-query'
import type { HookResultProps, InfiniteConfig } from '@internals/tanstack-query'
import { buildClientOptionType, buildOperationComments, requestGroupOrder } from '@internals/shared'

export type QueryProps = {
  name: string
  queryOptionsName: string
  queryKeyTypeName: string
  node: ast.OperationNode
  tsResolver: ResolverTs
  /**
   * Emits the `createInfiniteQuery` variant of the hook instead of `createQuery`.
   */
  infinite?: Pick<InfiniteConfig, 'initialPageParam' | 'queryParam'>
} & HookResultProps

const declarationPrinter = functionPrinter({ mode: 'declaration' })

function buildParamsAccessor(node: ast.OperationNode, tsResolver: ResolverTs): { param: FunctionParameterNode; isOptional: boolean } | null {
  const groups = ast.http.getRequestGroups(node)
  const names = requestGroupOrder.filter((key) => groups[key])
  if (names.length === 0) return null

  const requiredByGroup = {
    path: groups.required.path,
    query: groups.required.query,
    body: groups.body,
    headers: groups.required.headers,
    cookies: false,
  }
  const isOptional = names.every((name) => !requiredByGroup[name])
  const optionsName = tsResolver.response.options(node)

  return {
    isOptional,
    param: createFunctionParameter({
      name: 'params',
      type: `Accessor<${optionsName}>`,
      optional: isOptional,
    }),
  }
}

export function Query({
  name,
  queryKeyTypeName,
  queryOptionsName,
  node,
  tsResolver,
  infinite,
  returnType,
  typedHeaders,
}: QueryProps): KubbReactNode {
  const { TData, TError } = buildResponseTypes(node, tsResolver, { returnType, typedHeaders })
  const optionsName = infinite ? 'CreateInfiniteQueryOptions' : 'CreateQueryOptions'
  const hookName = infinite ? 'createInfiniteQuery' : 'createQuery'

  const pageParamType = infinite
    ? resolvePageParamType(node, { resolver: tsResolver, initialPageParam: infinite.initialPageParam, queryParam: infinite.queryParam }).pageParamType
    : null

  const generics = infinite
    ? [`TData = InfiniteData<${TData}>`, `TQueryKey extends QueryKey = ${queryKeyTypeName}`, `TPageParam = ${pageParamType}`]
    : [`TData = ${TData}`, `TQueryKey extends QueryKey = ${queryKeyTypeName}`]

  const optionsGenerics = infinite ? [TData, TError, 'TData', 'TQueryKey', 'TPageParam'] : [TData, TError, 'TData', 'TQueryKey']
  const observerOptionsType = `${optionsName}<${optionsGenerics.join(', ')}>`

  const requestParam = buildParamsAccessor(node, tsResolver)
  const queryOptionsArgs = requestParam ? `${requestParam.isOptional ? 'params?.()' : 'params()'}, config` : 'config'

  const paramsNode = createFunctionParameters({
    params: [
      ...(requestParam ? [requestParam.param] : []),
      createFunctionParameter({
        name: 'options',
        type: `Accessor<{
  query?: Partial<${observerOptionsType}>
  client?: ${buildClientOptionType()}
}>`,
        optional: true,
      }),
      createFunctionParameter({
        name: 'queryClient',
        type: 'Accessor<QueryClient>',
        optional: true,
      }),
    ],
  })
  const paramsSignature = declarationPrinter.print(paramsNode) ?? ''

  return (
    <Function name={name} export generics={generics.join(', ')} params={paramsSignature} JSDoc={{ comments: buildOperationComments(node) }}>
      {`
       return ${hookName}(() => {
         const { query = {}, client: config = {} } = options?.() ?? {}
         return {
           ...${queryOptionsName}(${queryOptionsArgs}),
           ...query,
         } as unknown as ${observerOptionsType}
       }, queryClient)
       `}
    </Function>
  )
}
