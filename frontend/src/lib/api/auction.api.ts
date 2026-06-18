import { createClient } from "@connectrpc/connect"
import { createConnectTransport } from "@connectrpc/connect-web"
import { AuctionService } from "#/gen/auction_platform/v1/auction_pb"

const transport = createConnectTransport({
    baseUrl: import.meta.env.VITE_AUCTION_URL 
})

export const authClient    = createClient(AuctionService,    transport)
