import { bidAuction, createAuction, getAuctionById, getMe } from '#/lib/services/auction.services'
import { useMutation } from '@tanstack/react-query'
import { useNavigate } from '@tanstack/react-router'
import { toast } from 'sonner'
import { queryOptions } from '@tanstack/react-query'
export function useCreateAuction() {
  const navigate = useNavigate()
  return useMutation({
    mutationFn: createAuction,
    onSuccess: (data) => {
      navigate({ to: `/dashboard/auction/${data.auction?.id}/edit` })
      toast.success(data.message)
    },
    onError: (err: Error) => {
      toast.error(err.message)
    },
  })
}

export function useGetAuctionById(){
  return useMutation({
    mutationFn:getAuctionById,
    onSuccess:(data)=>{
      toast.success(data.message)
    },
    onError:(err:Error)=>{
      toast.error(err.message)
    }
  })
}

export function useBidAuction(){
  return useMutation({
    mutationFn:bidAuction,
    onSuccess:(data)=>{
      toast.success(data.message)
    },
    onError:(err:Error)=>{
      toast.error(err.message)
    }
  })
}



export const meQueryOptions = () =>
  queryOptions({
    queryKey: ['me'],
    queryFn: () => getMe(),
    staleTime: 1000 * 60 * 5,
  })