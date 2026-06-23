import type { SocketStatus } from '#/hooks/use-auction-socket'
import { Badge } from '#/components/ui/badge'
import { Wifi, WifiOff, Loader2 } from 'lucide-react'

interface AuctionSocketStatusProps {
  status: SocketStatus
  onRetry?: () => void
}

// Small indicator shown in the bid panel — reusable anywhere
export function AuctionSocketStatus({
  status,
  onRetry,
}: AuctionSocketStatusProps) {
  if (status === 'connected') {
    return (
      <Badge
        variant="outline"
        className="gap-1.5 text-xs text-green-600 border-green-200 bg-green-50"
      >
        <Wifi className="size-3" />
        Live
      </Badge>
    )
  }

  if (status === 'connecting') {
    return (
      <Badge
        variant="outline"
        className="gap-1.5 text-xs text-amber-600 border-amber-200 bg-amber-50"
      >
        <Loader2 className="size-3 animate-spin" />
        Connecting…
      </Badge>
    )
  }

  return (
    <Badge
      variant="outline"
      className="gap-1.5 text-xs text-muted-foreground cursor-pointer hover:bg-muted"
      onClick={onRetry}
    >
      <WifiOff className="size-3" />
      {status === 'error'
        ? 'Connection error — click to retry'
        : 'Offline — click to retry'}
    </Badge>
  )
}
