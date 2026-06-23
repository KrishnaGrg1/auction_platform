import type { GetAuctionsListSchema } from '#/lib/schema/auction.schema'
import {
  bidAuction,
  createAuction,
  getAuctionById,
  getAuctionsList,
  getMe,
} from '#/lib/services/auction.services'
import { useMutation, useQuery, queryOptions } from '@tanstack/react-query'
import { useNavigate } from '@tanstack/react-router'
import { toast } from 'sonner'

// ── create auction ────────────────────────────────────────────────────────
export function useCreateAuction() {
  const navigate = useNavigate()
  return useMutation({
    mutationFn: createAuction,
    onSuccess: (data) => {
      navigate({ to: `/dashboard/auction/${data.auction?.id}` })
      toast.success(data.message)
    },
    onError: (err: Error) => {
      toast.error(err.message)
    },
  })
}

// ── get auctions list ─────────────────────────────────────────────────────
// useQuery — not useMutation
export function useGetAuctionsList(params?: Partial<GetAuctionsListSchema>) {
  return useQuery({
    queryKey: ['auctions', params],
    queryFn: () =>
      getAuctionsList({
        data: {
          page: params?.page ?? 1,
          page_size: params?.page_size ?? 20,
          status: params?.status,
          type: params?.type,
        },
      }),
    staleTime: 1000 * 30, // 30 seconds
  })
}

// ── get auction by id ─────────────────────────────────────────────────────
export function useGetAuctionById(auctionId: string) {
  return useQuery({
    queryKey: ['auction', auctionId],
    queryFn: () => getAuctionById({ data: { auction_id: auctionId } }),
    enabled: !!auctionId,
    staleTime: 1000 * 30,
  })
}

// ── bid auction ───────────────────────────────────────────────────────────
// mutation is correct here — bidding IS an action
export function useBidAuction() {
  return useMutation({
    mutationFn: bidAuction,
    onSuccess: (data) => {
      toast.success(data.message)
    },
    onError: (err: Error) => {
      toast.error(err.message)
    },
  })
}

// ── get me ────────────────────────────────────────────────────────────────
export const meQueryOptions = () =>
  queryOptions({
    queryKey: ['me'],
    queryFn: () => getMe(),
    staleTime: 1000 * 60 * 5, // 5 minutes
  })
