import z from 'zod'

enum AuctionType {
  UNSPECIFIED = 0,
  ENGLISH = 1,
  DUTCH = 2,
}
export const createAuctionSchema = z.object({
  title: z.string().min(3, 'At least 3 characters'),
  description: z.string().min(5, 'At least 5 characters'),
  type: z.nativeEnum(AuctionType),
  starting_price: z.number().positive(),
  reserved_price: z.number().min(0),
  extend_on_bid: z.boolean(),
  extend_minutes: z.number().min(1),
  // ✅ accept any non-empty string — convert to ISO in server fn
  start_time: z.string().min(1, 'Start time is required'),
  end_time: z.string().min(1, 'End time is required'),
  drop_amount: z.number().min(0).optional(),
  drop_interval: z.number().min(0).optional(),
})

export type CreateAuctionSchema = z.infer<typeof createAuctionSchema>

export const getAuctionByIdSchema=z.object({
  auction_id:z.string().uuid()
})
export type GetAuctionByIdSchema=z.infer<typeof getAuctionByIdSchema>

export const bidAuctionSchema=z.object({
  auction_id:z.string().uuid(),
  amount: z.number().positive(),
    is_auto_bid: z.boolean()
})
export type BidAuctionSchema=z.infer<typeof bidAuctionSchema>