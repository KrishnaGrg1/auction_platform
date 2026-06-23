import { createServerFn } from '@tanstack/react-start'
import { ConnectError } from '@connectrpc/connect'
import { createAuctionClient } from '../api/auction.api'
import {
  bidAuctionSchema,
  createAuctionSchema,
  getAuctionByIdSchema,
  getAuctionsListSchema,
} from '../schema/auction.schema'
import { timestampFromDate } from '@bufbuild/protobuf/wkt'
import { AuctionType } from '#/gen/auction_platform/v1/auction_pb'
import { getAuthToken } from '../token'

export const createAuction = createServerFn({ method: 'POST' })
  .inputValidator((data) => createAuctionSchema.parse(data))
  .handler(async ({ data }) => {
    try {
      // datetime-local → "2026-06-20T14:30"
      // new Date()     → treats as LOCAL time → converts to UTC internally
      // timestampFromDate → protobuf Timestamp in UTC ✅
      const startDate = new Date(data.start_time)
      const endDate = new Date(data.end_time)

      if (isNaN(startDate.getTime())) {
        throw new Error('Invalid start time')
      }
      if (isNaN(endDate.getTime())) {
        throw new Error('Invalid end time')
      }
      if (endDate <= startDate) {
        throw new Error('End time must be after start time')
      }
      const isDutch = data.type === AuctionType.DUTCH

      const token = getAuthToken()
      const auctionClient = createAuctionClient(token ?? undefined)
      const res = await auctionClient.createAuction({
        title: data.title,
        description: data.description,
        type: data.type,
        startingPrice: BigInt(data.starting_price),
        reservedPrice: BigInt(data.reserved_price),
        extendOnBid: data.extend_on_bid,
        extendMinutes: data.extend_minutes,
        startTime: timestampFromDate(startDate),
        endTime: timestampFromDate(endDate),
        dropAmount: isDutch ? BigInt(data.drop_amount ?? 0) : BigInt(0),
        dropInterval: isDutch ? (data.drop_interval ?? 0) : 1,
      })

      return res
    } catch (error: unknown) {
      if (error instanceof ConnectError) {
        const msg =
          typeof error.rawMessage === 'string'
            ? error.rawMessage
            : error.message

        throw new Error(msg)
      }
      if (error instanceof Error) {
        throw error
      }
      throw new Error('Failed to create auction')
    }
  })

// export const getAllAuction=createServerFn({method:'GET'})

export const getAuctionById = createServerFn({ method: 'POST' })
  .inputValidator((data) => getAuctionByIdSchema.parse(data))
  .handler(async ({ data }) => {
    try {
      const token = getAuthToken()
      const auctionClient = createAuctionClient(token ?? undefined)
      const resp = await auctionClient.getAuctionDetailsById({
        auctionId: data.auction_id,
      })
      return resp
    } catch (error: unknown) {
      if (error instanceof ConnectError) {
        const msg =
          typeof error.rawMessage === 'string'
            ? error.rawMessage
            : error.message

        throw new Error(msg)
      }
      if (error instanceof Error) {
        throw error
      }
      throw new Error('Failed to create auction')
    }
  })

export const bidAuction = createServerFn({ method: 'POST' })
  .inputValidator((data) => bidAuctionSchema.parse(data))
  .handler(async ({ data }) => {
    try {
      const token = getAuthToken()
      const auctionClient = createAuctionClient(token ?? undefined)
      const resp = await auctionClient.bidAuction({
        auctionId: data.auction_id,
        amount: BigInt(data.amount),
        isAutoBid: data.is_auto_bid,
      })
      return resp
    } catch (error: unknown) {
      if (error instanceof ConnectError) {
        const msg =
          typeof error.rawMessage === 'string'
            ? error.rawMessage
            : error.message

        throw new Error(msg)
      }
      if (error instanceof Error) {
        throw error
      }
      throw new Error('Failed to create auction')
    }
  })

// ✅ GET not POST for fetching
export const getAuctionsList = createServerFn({ method: 'GET' })
  .inputValidator((data) => getAuctionsListSchema.parse(data))
  .handler(async ({ data }) => {
    try {
      const token = getAuthToken()
      const auctionClient = createAuctionClient(token ?? undefined)

      const req: Record<string, unknown> = {
        page: data.page,
        pageSize: data.page_size,
      }
      if (data.status && data.status !== 0) req.status = data.status
      if (data.type && data.type !== 0) req.type = data.type

      const resp = await auctionClient.getAuctionsList(req)
      return resp
    } catch (error: unknown) {
      if (error instanceof ConnectError) {
        throw new Error(
          typeof error.rawMessage === 'string'
            ? error.rawMessage
            : error.message,
        )
      }
      if (error instanceof Error) throw error
      throw new Error('Failed to fetch auctions')
    }
  })

export const getMe = createServerFn({ method: 'POST' }).handler(async () => {
  try {
    const token = getAuthToken()
    const authClient = createAuctionClient(token ?? undefined)
    const res = await authClient.getMe({})
    return res
  } catch (error: unknown) {
    if (error instanceof ConnectError) {
      const msg =
        typeof error.rawMessage === 'string' ? error.rawMessage : error.message

      throw new Error(msg)
    }

    throw new Error('Failed to Get Me')
  }
})
