import { createFileRoute, Link, notFound } from '@tanstack/react-router'
import { useForm } from '@tanstack/react-form'
import { getAuctionById } from '#/lib/services/auction.services'
import {
  AuctionStatus,
  AuctionType,
} from '#/gen/auction_platform/v1/auction_pb'
import { Card, CardContent, CardHeader, CardTitle } from '#/components/ui/card'
import { Button } from '#/components/ui/button'
import { Badge } from '#/components/ui/badge'
import { Input } from '#/components/ui/input'
import { Label } from '#/components/ui/label'
import { Separator } from '#/components/ui/separator'
import { Alert, AlertDescription } from '#/components/ui/alert'
import { DashboardLayout } from '#/components/dashboard-layout'
import { ArrowLeft, Gavel, TrendingDown, ShieldCheck, Info } from 'lucide-react'
import { useBidAuction } from '#/hooks/use-auction'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { meQueryOptions } from '#/hooks/use-auction'
import { useAuctionSocket } from '#/hooks/use-auction-socket'
import { AuctionSocketStatus } from '#/components/auction-socket-status'
import { useState, useCallback } from 'react'
import type { AuctionEvent } from '#/hooks/use-auction-socket'
import { toast } from 'sonner'

export const Route = createFileRoute('/dashboard/auction/$id/')({
  component: AuctionDetailPage,
  loader: async ({ params }) => {
    try {
      const resp = await getAuctionById({ data: { auction_id: params.id } })
      if (!resp.auction) throw notFound()
      return resp
    } catch (err) {
      if (err && typeof err === 'object' && 'isNotFound' in err) throw err
      throw notFound()
    }
  },
  notFoundComponent: () => <AuctionNotFound />,
  pendingComponent: () => <AuctionDetailSkeleton />,
})

function AuctionDetailPage() {
  const { id: auctionId } = Route.useParams()
  const data = Route.useLoaderData()
  const auction = data.auction!
  const queryClient = useQueryClient()

  const {
    mutate: placeBid,
    isPending: isBidding,
    error: bidError,
  } = useBidAuction()
  const { data: userDetails, isPending: userPending } =
    useQuery(meQueryOptions())

  // ── live price state — updated by WebSocket ──────────────────────────
  const [livePrice, setLivePrice] = useState<number>(
    Number(auction.currentPrice),
  )
  const [liveBidderId, setLiveBidderId] = useState<string>(
    auction.currentBidderId ?? '',
  )
  const [liveStatus, setLiveStatus] = useState<AuctionStatus>(auction.status)

  // ── WebSocket event handlers ──────────────────────────────────────────
  const handleNewBid = useCallback(
    (event: AuctionEvent) => {
      if (event.amount !== undefined) {
        setLivePrice(event.amount)
      }
      if (event.user_id) {
        setLiveBidderId(event.user_id)
      }
      // invalidate query so full data refreshes in background
      queryClient.invalidateQueries({ queryKey: ['auction', auctionId] })
      toast.info(`New bid: ${formatCents(event.amount ?? 0)}`)
    },
    [auctionId, queryClient],
  )

  const handleAuctionEnded = useCallback(
    (event: AuctionEvent) => {
      setLiveStatus(AuctionStatus.ENDED)
      queryClient.invalidateQueries({ queryKey: ['auction', auctionId] })
      if (event.type === 'auction_won') {
        toast.success('Auction won!')
      } else {
        toast.info('Auction has ended')
      }
    },
    [auctionId, queryClient],
  )

  const handlePriceDropped = useCallback((event: AuctionEvent) => {
    if (event.amount !== undefined) {
      setLivePrice(event.amount)
    }
  }, [])

  const handleExtended = useCallback(
    (_event: AuctionEvent) => {
      queryClient.invalidateQueries({ queryKey: ['auction', auctionId] })
      toast.info('Auction time extended due to late bid')
    },
    [auctionId, queryClient],
  )

  // ── connect to WebSocket ──────────────────────────────────────────────
  const { status: socketStatus, reconnect } = useAuctionSocket({
    auctionId,
    userId: userDetails?.user?.id,
    enabled: liveStatus === AuctionStatus.ACTIVE,
    onNewBid: handleNewBid,
    onAuctionEnded: handleAuctionEnded,
    onPriceDropped: handlePriceDropped,
    onExtended: handleExtended,
  })

  // ── derived flags ─────────────────────────────────────────────────────
  const isDutch = auction.type === AuctionType.DUTCH
  const isOwner = userDetails?.user?.id === auction.sellerId
  const isActive = liveStatus === AuctionStatus.ACTIVE
  const isCurrentWinner = userDetails?.user?.id === liveBidderId

  if (userPending) return <AuctionDetailSkeleton />

  return (
    <DashboardLayout>
      <div className="max-w-4xl mx-auto space-y-5">
        <Link
          to="/dashboard/auction"
          className="inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground transition-colors"
        >
          <ArrowLeft className="size-3.5" />
          Back to auctions
        </Link>

        <div className="grid gap-5 lg:grid-cols-[1.4fr_1fr]">
          {/* ── Left — details ─────────────────────────────────────── */}
          <div className="space-y-5">
            <Card>
              <CardContent className="p-6">
                <div className="flex items-start justify-between gap-3 mb-3">
                  <h1 className="text-xl font-bold leading-tight">
                    {auction.title}
                  </h1>
                  <StatusBadge status={liveStatus} />
                </div>

                <div className="flex items-center gap-2 mb-4">
                  <Badge variant="outline" className="gap-1 text-xs">
                    {isDutch ? (
                      <>
                        <TrendingDown className="size-3" />
                        Dutch auction
                      </>
                    ) : (
                      <>
                        <Gavel className="size-3" />
                        English auction
                      </>
                    )}
                  </Badge>
                  {isOwner && (
                    <Badge variant="secondary" className="text-xs">
                      You're the seller
                    </Badge>
                  )}
                </div>

                <p className="text-sm text-muted-foreground leading-relaxed whitespace-pre-wrap">
                  {auction.description}
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="text-base">Auction details</CardTitle>
              </CardHeader>
              <CardContent>
                <dl className="grid grid-cols-2 gap-4 text-sm">
                  <DetailRow
                    label="Starting price"
                    value={formatCents(auction.startingPrice)}
                  />
                  <DetailRow
                    label="Reserve price"
                    value={
                      Number(auction.reservedPrice) > 0
                        ? formatCents(auction.reservedPrice)
                        : 'No reserve'
                    }
                  />
                  <DetailRow
                    label="Starts"
                    value={formatDate(auction.startTime)}
                  />
                  <DetailRow label="Ends" value={formatDate(auction.endTime)} />
                  {isDutch && (
                    <>
                      <DetailRow
                        label="Drops by"
                        value={formatCents(auction.dropAmount)}
                      />
                      <DetailRow
                        label="Drop interval"
                        value={`Every ${auction.dropInterval}s`}
                      />
                    </>
                  )}
                  {auction.extendOnBid && (
                    <DetailRow
                      label="Anti-snipe"
                      value={`Extends ${auction.extendMinutes}m on late bid`}
                    />
                  )}
                </dl>
              </CardContent>
            </Card>
          </div>

          {/* ── Right — bid panel ──────────────────────────────────── */}
          <div className="space-y-5">
            <Card className="sticky top-6">
              <CardContent className="p-6">
                {/* live connection indicator */}
                <div className="flex items-center justify-between mb-3">
                  <p className="text-xs text-muted-foreground">
                    {isDutch ? 'Current price' : 'Current bid'}
                  </p>
                  {isActive && (
                    <AuctionSocketStatus
                      status={socketStatus}
                      onRetry={reconnect}
                    />
                  )}
                </div>

                {/* live price — updates in real time */}
                <p className="text-3xl font-bold mb-1 tabular-nums transition-all">
                  {formatCents(livePrice)}
                </p>

                {isCurrentWinner && isActive && (
                  <Badge variant="default" className="gap-1 text-xs mb-3">
                    <ShieldCheck className="size-3" />
                    You're winning
                  </Badge>
                )}

                <Separator className="my-4" />

                {/* bid panel states — explicit per-status */}
                {isOwner ? (
                  <Alert>
                    <Info className="size-4" />
                    <AlertDescription>
                      You listed this auction — bidding is disabled for sellers.
                    </AlertDescription>
                  </Alert>
                ) : liveStatus === AuctionStatus.SCHEDULED ? (
                  <Alert>
                    <Info className="size-4" />
                    <AlertDescription>
                      This auction hasn't started yet.
                    </AlertDescription>
                  </Alert>
                ) : liveStatus === AuctionStatus.ENDED ? (
                  <Alert>
                    <Info className="size-4" />
                    <AlertDescription>This auction has ended.</AlertDescription>
                  </Alert>
                ) : liveStatus === AuctionStatus.CANCELLED ? (
                  <Alert>
                    <Info className="size-4" />
                    <AlertDescription>
                      This auction was cancelled.
                    </AlertDescription>
                  </Alert>
                ) : isActive ? (
                  <BidForm
                    auctionId={auction.id}
                    currentPrice={livePrice}
                    isDutch={isDutch}
                    placeBid={placeBid}
                    isBidding={isBidding}
                    bidError={bidError}
                  />
                ) : (
                  <Alert>
                    <Info className="size-4" />
                    <AlertDescription>
                      This auction is not available for bidding.
                    </AlertDescription>
                  </Alert>
                )}
              </CardContent>
            </Card>
          </div>
        </div>
      </div>
    </DashboardLayout>
  )
}

/* ─────────────────────────── Bid form ─────────────────────────── */
function BidForm({
  auctionId,
  currentPrice,
  isDutch,
  placeBid,
  isBidding,
  bidError,
}: {
  auctionId: string
  currentPrice: number | bigint
  isDutch: boolean
  placeBid: (input: {
    data: { auction_id: string; amount: number; is_auto_bid: boolean }
  }) => void
  isBidding: boolean
  bidError: unknown
}) {
  const minBid = Number(currentPrice) + 100

  const form = useForm({
    defaultValues: { amount: isDutch ? Number(currentPrice) : minBid },
    onSubmit: async ({ value }) => {
      placeBid({
        data: {
          auction_id: auctionId,
          amount: value.amount,
          is_auto_bid: false,
        },
      })
    },
  })

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault()
        e.stopPropagation()
        form.handleSubmit()
      }}
      className="space-y-4"
    >
      {bidError !== null && bidError !== undefined && (
        <Alert variant="destructive">
          <AlertDescription>
            {bidError instanceof Error
              ? bidError.message
              : 'Failed to place bid.'}
          </AlertDescription>
        </Alert>
      )}

      {isDutch ? (
        <div className="space-y-2">
          <p className="text-xs text-muted-foreground">
            Accept the current price to win this item instantly.
          </p>
          <Button type="submit" className="w-full" disabled={isBidding}>
            {isBidding ? 'Placing…' : `Accept at ${formatCents(currentPrice)}`}
          </Button>
        </div>
      ) : (
        <>
          <form.Field
            name="amount"
            validators={{
              onChange: ({ value }) =>
                value < minBid
                  ? `Must be at least ${formatCents(minBid)}`
                  : undefined,
            }}
          >
            {(field) => (
              <div className="space-y-1.5">
                <Label htmlFor={field.name}>Your bid (¢)</Label>
                <Input
                  id={field.name}
                  name={field.name}
                  type="number"
                  min={minBid}
                  value={field.state.value}
                  aria-invalid={field.state.meta.errors.length > 0}
                  onChange={(e) => field.handleChange(Number(e.target.value))}
                  onBlur={field.handleBlur}
                />
                <p className="text-xs text-muted-foreground">
                  Minimum bid: {formatCents(minBid)}
                </p>
                {field.state.meta.errors.length > 0 && (
                  <p className="text-xs text-destructive">
                    {String(field.state.meta.errors[0])}
                  </p>
                )}
              </div>
            )}
          </form.Field>

          <form.Subscribe selector={(s) => s.canSubmit}>
            {(canSubmit) => (
              <Button
                type="submit"
                className="w-full"
                disabled={!canSubmit || isBidding}
              >
                {isBidding ? 'Placing bid…' : 'Place bid'}
              </Button>
            )}
          </form.Subscribe>
        </>
      )}
    </form>
  )
}

/* ─────────────────────────── Shared ───────────────────────────── */

function StatusBadge({ status }: { status: AuctionStatus }) {
  if (status === AuctionStatus.ACTIVE)
    return <Badge variant="default">Live</Badge>
  if (status === AuctionStatus.SCHEDULED)
    return <Badge variant="secondary">Scheduled</Badge>
  if (status === AuctionStatus.ENDED)
    return <Badge variant="outline">Ended</Badge>
  if (status === AuctionStatus.CANCELLED)
    return <Badge variant="destructive">Cancelled</Badge>
  return <Badge variant="outline">Unknown ({status})</Badge>
}

function DetailRow({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <dt className="text-xs text-muted-foreground mb-0.5">{label}</dt>
      <dd className="font-medium">{value}</dd>
    </div>
  )
}

function AuctionNotFound() {
  return (
    <div className="min-h-screen flex items-center justify-center px-4">
      <Card className="max-w-md w-full">
        <CardContent className="py-10 text-center">
          <p className="font-medium text-sm">Auction not found</p>
          <p className="text-xs text-muted-foreground mt-1 mb-4">
            This listing may have been removed or the link is incorrect.
          </p>
          <Button asChild size="sm" variant="outline">
            <Link to="/dashboard/auction">
              <ArrowLeft className="size-3.5" />
              Back to auctions
            </Link>
          </Button>
        </CardContent>
      </Card>
    </div>
  )
}

function AuctionDetailSkeleton() {
  return (
    <div className="min-h-screen py-8 px-4 bg-muted/30 animate-pulse">
      <div className="max-w-4xl mx-auto">
        <div className="h-4 w-32 mb-5 bg-muted rounded" />
        <div className="grid gap-5 lg:grid-cols-[1.4fr_1fr]">
          <div className="space-y-5">
            <Card>
              <CardContent className="p-6 space-y-3">
                <div className="h-6 w-3/4 bg-muted rounded" />
                <div className="h-4 w-1/3 bg-muted rounded" />
                <div className="h-16 w-full bg-muted rounded" />
              </CardContent>
            </Card>
          </div>
          <Card>
            <CardContent className="p-6 space-y-3">
              <div className="h-3 w-20 bg-muted rounded" />
              <div className="h-9 w-32 bg-muted rounded" />
              <div className="h-10 w-full mt-4 bg-muted rounded" />
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}

function formatCents(cents: bigint | number): string {
  const value = Number(cents) / 100
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
  }).format(value)
}

function formatDate(ts?: { seconds: bigint }): string {
  if (!ts) return '—'
  const date = new Date(Number(ts.seconds) * 1000)
  return new Intl.DateTimeFormat('en-US', {
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
  }).format(date)
}
