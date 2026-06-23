import { AuthService } from '#/gen/auction_platform/v1/auth_pb'
import { createClient } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'

const transport = createConnectTransport({
  baseUrl: import.meta.env.VITE_AUTH_URL ?? 'http://localhost:8080',
})

export const authClient = createClient(AuthService, transport)
