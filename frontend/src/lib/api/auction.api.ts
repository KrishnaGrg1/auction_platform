import { AuctionService } from '#/gen/auction_platform/v1/auction_pb'
import { createClient } from '@connectrpc/connect'

import { createConnectTransport } from '@connectrpc/connect-web'

export function createAuctionClient(token?: string | null) {
  const transport = createConnectTransport({
    baseUrl: import.meta.env.VITE_AUCTION_URL,
    interceptors: [
      (next) => async (req) => {
        if (token) {
          req.header.set('Authorization', `Bearer ${token}`)
        }
        return next(req)
      },
    ],
  })

  return createClient(AuctionService, transport)
}
