import { createFileRoute, Link } from '@tanstack/react-router'
import { DashboardLayout } from '#/components/dashboard-layout'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from '#/components/ui/card'
import { Button } from '#/components/ui/button'
import { Badge } from '#/components/ui/badge'
import { Input } from '#/components/ui/input'
import { Search, PlusCircle, Gavel, TrendingDown } from 'lucide-react'
import { useGetAuctionsList } from '#/hooks/use-auction'
import {
  AuctionStatus,
  AuctionType,
} from '#/gen/auction_platform/v1/auction_pb'
import { useState } from 'react'
import { Alert, AlertDescription } from '#/components/ui/alert'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '#/components/ui/table'

function statusBadgeVariant(status: number) {
  switch (status) {
    case AuctionStatus.ACTIVE:
      return 'default' as const
    case AuctionStatus.SCHEDULED:
      return 'secondary' as const
    case AuctionStatus.ENDED:
      return 'outline' as const
    case AuctionStatus.CANCELLED:
      return 'destructive' as const
    default:
      return 'outline' as const
  }
}

function statusLabel(status: number): string {
  const labels: Record<number, string> = {
    [AuctionStatus.SCHEDULED]: 'Scheduled',
    [AuctionStatus.ACTIVE]: 'Live',
    [AuctionStatus.ENDED]: 'Ended',
    [AuctionStatus.CANCELLED]: 'Cancelled',
  }
  return labels[status] ?? 'Unknown'
}

function typeLabel(type: number): string {
  const labels: Record<number, string> = {
    [AuctionType.UNSPECIFIED]: 'Unspecified',
    [AuctionType.ENGLISH]: 'English',
    [AuctionType.DUTCH]: 'Dutch',
  }
  return labels[type] ?? 'Unknown'
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

function formatCents(cents: bigint | number): string {
  const value = Number(cents) / 100
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
  }).format(value)
}

export const Route = createFileRoute('/dashboard/auction/')({
  component: AuctionsListPage,
})

function AuctionsListPage() {
  const [filter, setFilter] = useState<number | null>(null)
  const [search, setSearch] = useState('')

  // ✅ useQuery — no useEffect needed, no manual trigger
  const {
    data: auctionsData,
    isLoading, // use isLoading not isPending
    isError,
    error,
  } = useGetAuctionsList({ page: 1, page_size: 50 })

  const allAuctions = auctionsData?.auctions ?? []

  // client-side filter by status + search
  const auctions = allAuctions.filter((a) => {
    const matchesFilter = filter === null || a.status === filter
    const matchesSearch =
      search === '' || a.title.toLowerCase().includes(search.toLowerCase())
    return matchesFilter && matchesSearch
  })

  return (
    <DashboardLayout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h1 className="text-2xl font-bold tracking-tight">Auctions</h1>
            <p className="text-sm text-muted-foreground mt-1">
              Manage and monitor your auction listings
            </p>
          </div>
          <Button asChild>
            <Link to="/dashboard/auction/create">
              <PlusCircle className="size-4" />
              Create auction
            </Link>
          </Button>
        </div>

        {/* Filters + Search */}
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
          <div className="relative flex-1 max-w-sm">
            <Search className="absolute left-2.5 top-2.5 size-4 text-muted-foreground" />
            <Input
              placeholder="Search auctions…"
              className="pl-8"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
            />
          </div>
          <div className="flex gap-2">
            <Button
              variant={filter === null ? 'default' : 'outline'}
              size="sm"
              className="text-xs"
              onClick={() => setFilter(null)}
            >
              All
            </Button>
            {[
              { label: 'Active', value: AuctionStatus.ACTIVE },
              { label: 'Scheduled', value: AuctionStatus.SCHEDULED },
              { label: 'Ended', value: AuctionStatus.ENDED },
            ].map((f) => (
              <Button
                key={f.value}
                variant={filter === f.value ? 'default' : 'outline'}
                size="sm"
                className="text-xs"
                onClick={() => setFilter(f.value)}
              >
                {f.label}
              </Button>
            ))}
          </div>
        </div>

        {/* Error state */}
        {isError && (
          <Alert variant="destructive">
            <AlertDescription>
              {error instanceof Error
                ? error.message
                : 'Failed to load auctions'}
            </AlertDescription>
          </Alert>
        )}

        {/* Table */}
        <Card>
          <CardHeader>
            <CardTitle className="text-base">All listings</CardTitle>
            <CardDescription>
              {auctions.length} auction{auctions.length !== 1 ? 's' : ''}
              {filter !== null ? ' (filtered)' : ''}
            </CardDescription>
          </CardHeader>
          <CardContent className="p-0">
            {isLoading ? (
              <div className="p-6 space-y-3">
                {[1, 2, 3, 4].map((i) => (
                  <div
                    key={i}
                    className="h-12 animate-pulse rounded bg-muted"
                  />
                ))}
              </div>
            ) : auctions.length === 0 ? (
              <div className="py-12 text-center text-sm text-muted-foreground">
                No auctions found.
              </div>
            ) : (
              <div className="overflow-x-auto">
                <Table className="w-full text-sm">
                  <TableHeader>
                    <TableRow className="border-b text-left text-muted-foreground">
                      <TableHead className="px-4 py-3 font-medium">
                        Title
                      </TableHead>
                      <TableHead className="px-4 py-3 font-medium">
                        Status
                      </TableHead>
                      <TableHead className="px-4 py-3 font-medium">
                        Type
                      </TableHead>
                      <TableHead className="px-4 py-3 font-medium">
                        Current price
                      </TableHead>
                      <TableHead className="px-4 py-3 font-medium">
                        Ends
                      </TableHead>
                      <TableHead className="px-4 py-3 font-medium" />
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {auctions.map((auction) => (
                      <TableRow
                        key={auction.id}
                        className="border-b last:border-0 hover:bg-muted/30 transition-colors"
                      >
                        <TableCell className="px-4 py-3 font-medium max-w-[200px] truncate">
                          {auction.title}
                        </TableCell>
                        <TableCell className="px-4 py-3">
                          <Badge
                            variant={statusBadgeVariant(auction.status)}
                            className="text-[10px] px-1.5 py-0"
                          >
                            {statusLabel(auction.status)}
                          </Badge>
                        </TableCell>
                        <TableCell className="px-4 py-3 text-muted-foreground">
                          <span className="flex items-center gap-1.5">
                            {auction.type === AuctionType.DUTCH ? (
                              <TrendingDown className="size-3" />
                            ) : (
                              <Gavel className="size-3" />
                            )}
                            {typeLabel(auction.type)}
                          </span>
                        </TableCell>
                        <TableCell className="px-4 py-3 font-semibold">
                          {formatCents(auction.currentPrice)}
                        </TableCell>
                        <TableCell className="px-4 py-3 text-muted-foreground">
                          {formatDate(auction.endTime)}
                        </TableCell>
                        <TableCell className="px-4 py-3">
                          <Button asChild variant="ghost" size="sm">
                            <Link
                              to="/dashboard/auction/$id"
                              params={{ id: auction.id }}
                            >
                              View
                            </Link>
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </DashboardLayout>
  )
}
