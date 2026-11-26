import type { ClientConfig } from './core'
import { serializePath } from './serializers/path'
import { serializeStyledQuery } from './serializers/queryStyles'

/**
 * The config `client.ts` starts with: what this spec needs from the runtime.
 */
export const defaultConfig: ClientConfig = { serializer: { query: serializeStyledQuery, path: serializePath } }
