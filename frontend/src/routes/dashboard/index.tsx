import { createFileRoute, Link } from '@tanstack/react-router'
import { Card, CardContent, CardHeader, CardTitle } from '#/components/ui/card'
import { Badge } from '#/components/ui/badge'
import { Button } from '#/components/ui/button'
import { DashboardLayout } from '#/components/dashboard-layout'
import {
  Gavel,
  TrendingUp,
  Users,
  Clock,
  PlusCircle,
  ArrowRight,
} from 'lucide-react'
import { useGetAuctionsList, meQueryOptions } from '#/hooks/use-auction'
import { AuctionStatus } from '#/gen/auction_platform/v1/auction_pb'
import { useQuery } from '@tanstack/react-query'
import { useMemo } from 'react'

export const Route = createFileRoute('/dashboard/')({
  component: DashboardPage,
})

function DashboardPage() {
  const { data: auctionsData, isLoading: isPending } = useGetAuctionsList({
    page: 1,
    page_size: 50,
  })
  console.log('auctions:', auctionsData)
  const { data: userDetails } = useQuery(meQueryOptions())

  const auctions = auctionsData?.auctions ?? []

  const activeAuctions = useMemo(
    () => auctions.filter((a) => a.status === AuctionStatus.ACTIVE),
    [auctions],
  )
  const scheduledAuctions = useMemo(
    () => auctions.filter((a) => a.status === AuctionStatus.SCHEDULED),
    [auctions],
  )
  const endedAuctions = useMemo(
    () => auctions.filter((a) => a.status === AuctionStatus.ENDED),
    [auctions],
  )

  return (
    <DashboardLayout>
      <div className="space-y-6">
        {/* Page header */}
        <div className="flex items-center justify-between gap-4">
          <div>
            <h1 className="text-2xl font-bold tracking-tight">
              Welcome
              {userDetails?.user?.firstName
                ? `, ${userDetails.user.firstName}`
                : ''}
            </h1>
            <p className="text-sm text-muted-foreground mt-1">
              Overview of your auction activity
            </p>
          </div>
          <Button asChild>
            <Link to="/dashboard/auction/create">
              <PlusCircle className="size-4" />
              New auction
            </Link>
          </Button>
        </div>

        {/* Stats grid */}
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <StatCard
            icon={Gavel}
            label="Active auctions"
            value={String(activeAuctions.length)}
            trend={
              scheduledAuctions.length > 0
                ? `${scheduledAuctions.length} upcoming`
                : '0 scheduled'
            }
            trendUp={scheduledAuctions.length > 0}
          />
          <StatCard
            icon={TrendingUp}
            label="Total auctions"
            value={String(auctions.length)}
            trend={`${activeAuctions.length} live now`}
            trendUp={activeAuctions.length > 0}
          />
          <StatCard
            icon={Users}
            label="Ended"
            value={String(endedAuctions.length)}
            trend="Completed listings"
            trendUp={false}
          />
          <StatCard
            icon={Clock}
            label="Ending soon"
            value={String(activeAuctions.length)}
            trend="Active right now"
            trendUp
          />
        </div>

        {/* Recent auctions & quick actions */}
        <div className="grid gap-6 lg:grid-cols-[1.5fr_1fr]">
          {/* Recent auctions */}
          <Card>
            <CardHeader className="flex flex-row items-center justify-between">
              <CardTitle className="text-base">Recent auctions</CardTitle>
              <Button asChild variant="ghost" size="sm">
                <Link to="/dashboard/auction">
                  View all <ArrowRight className="ml-1 size-3.5" />
                </Link>
              </Button>
            </CardHeader>
            <CardContent>
              {isPending ? (
                <div className="space-y-3">
                  {[1, 2, 3].map((i) => (
                    <div
                      key={i}
                      className="h-16 animate-pulse rounded-lg bg-muted"
                    />
                  ))}
                </div>
              ) : auctions.length === 0 ? (
                <div className="py-8 text-center text-sm text-muted-foreground">
                  No auctions yet. Create your first listing to get started.
                </div>
              ) : (
                <div className="space-y-3">
                  {auctions.slice(0, 5).map((auction) => (
                    <Link
                      key={auction.id}
                      to="/dashboard/auction/$id"
                      params={{ id: auction.id }}
                      className="flex items-center justify-between gap-4 rounded-lg border p-3 hover:bg-muted/50 transition-colors"
                    >
                      <div className="min-w-0 flex-1">
                        <p className="text-sm font-medium truncate">
                          {auction.title}
                        </p>
                      </div>
                      <div className="text-right shrink-0">
                        <p className="text-sm font-semibold">
                          {formatCents(auction.currentPrice)}
                        </p>
                        <Badge
                          variant={
                            auction.status === AuctionStatus.ACTIVE
                              ? 'default'
                              : auction.status === AuctionStatus.SCHEDULED
                                ? 'secondary'
                                : 'outline'
                          }
                          className="text-[10px] px-1.5 py-0"
                        >
                          {statusLabel(auction.status)}
                        </Badge>
                      </div>
                    </Link>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>

          {/* Quick actions */}
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Quick actions</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              <Button
                asChild
                variant="outline"
                className="w-full justify-start"
              >
                <Link to="/dashboard/auction/create">
                  <PlusCircle className="mr-2 size-4" />
                  List a new auction
                </Link>
              </Button>
              <Button
                asChild
                variant="outline"
                className="w-full justify-start"
              >
                <Link to="/dashboard/auction">
                  <ListChecks className="mr-2 size-4" />
                  Browse all auctions
                </Link>
              </Button>
            </CardContent>
          </Card>
        </div>
      </div>
    </DashboardLayout>
  )
}

function StatCard({
  icon: Icon,
  label,
  value,
  trend,
  trendUp,
}: {
  icon: React.ComponentType<{ className?: string }>
  label: string
  value: string
  trend: string
  trendUp: boolean
}) {
  return (
    <Card>
      <CardContent className="p-5">
        <div className="flex items-center justify-between mb-3">
          <span className="rounded-lg bg-amber/10 p-2">
            <Icon className="size-4 text-amber" />
          </span>
        </div>
        <p className="text-2xl font-bold">{value}</p>
        <p className="text-xs text-muted-foreground mt-1">{label}</p>
        <p
          className={`text-xs mt-1 ${trendUp ? 'text-green-600' : 'text-muted-foreground'}`}
        >
          {trend}
        </p>
      </CardContent>
    </Card>
  )
}

function ListChecks({ className }: { className?: string }) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      className={className}
    >
      <path d="M11 12H3" />
      <path d="M16 6H3" />
      <path d="M16 18H3" />
      <path d="M18 9l2 2 4-4" />
    </svg>
  )
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

function formatCents(cents: bigint | number): string {
  const value = Number(cents) / 100
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
  }).format(value)
}
