import { getAuthToken } from '#/lib/token'
import { useEffect, useRef, useCallback, useState } from 'react'

// ── Event types matching your Go backend AuctionEvent struct ─────────────
export type AuctionEventType =
  | 'connected'
  | 'new_bid'
  | 'auction_won'
  | 'auction_ended'
  | 'auction_cancelled'
  | 'auction_no_reserve'
  | 'price_dropped'   // Dutch price drop
  | 'extended'        // anti-snipe extension
  | 'pong'

export interface AuctionEvent {
  type:       AuctionEventType
  auction_id: string
  user_id?:   string
  amount?:    number
  message?:   string
  timestamp:  string
}

// ── Connection states ─────────────────────────────────────────────────────
export type SocketStatus = 'connecting' | 'connected' | 'disconnected' | 'error'

interface UseAuctionSocketOptions {
  auctionId:   string
  userId?:     string
  enabled?:    boolean                        // default true
  onNewBid?:   (event: AuctionEvent) => void
  onAuctionEnded?: (event: AuctionEvent) => void
  onPriceDropped?: (event: AuctionEvent) => void
  onExtended?: (event: AuctionEvent) => void
  onConnected?: () => void
  onDisconnected?: () => void
}

interface UseAuctionSocketReturn {
  status:       SocketStatus
  lastEvent:    AuctionEvent | null
  isConnected:  boolean
  disconnect:   () => void
  reconnect:    () => void
}

const WS_BASE_URL = import.meta.env.VITE_WS_URL ?? 'ws://localhost:8081'
const MAX_RETRIES       = 5
const INITIAL_RETRY_MS  = 1000   // 1s → 2s → 4s → 8s → 16s (exponential)

// ── Main hook ─────────────────────────────────────────────────────────────
export function useAuctionSocket({
  auctionId,
  userId,
  enabled = true,
  onNewBid,
  onAuctionEnded,
  onPriceDropped,
  onExtended,
  onConnected,
  onDisconnected,
}: UseAuctionSocketOptions): UseAuctionSocketReturn {

  const wsRef         = useRef<WebSocket | null>(null)
  const retryCountRef = useRef(0)
  const retryTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const mountedRef    = useRef(true)

  const [status,    setStatus]    = useState<SocketStatus>('disconnected')
  const [lastEvent, setLastEvent] = useState<AuctionEvent | null>(null)

  // stable refs so handlers never go stale without reconnecting
  const onNewBidRef        = useRef(onNewBid)
  const onAuctionEndedRef  = useRef(onAuctionEnded)
  const onPriceDroppedRef  = useRef(onPriceDropped)
  const onExtendedRef      = useRef(onExtended)
  const onConnectedRef     = useRef(onConnected)
  const onDisconnectedRef  = useRef(onDisconnected)

  useEffect(() => { onNewBidRef.current       = onNewBid       }, [onNewBid])
  useEffect(() => { onAuctionEndedRef.current = onAuctionEnded }, [onAuctionEnded])
  useEffect(() => { onPriceDroppedRef.current = onPriceDropped }, [onPriceDropped])
  useEffect(() => { onExtendedRef.current     = onExtended     }, [onExtended])
  useEffect(() => { onConnectedRef.current    = onConnected    }, [onConnected])
  useEffect(() => { onDisconnectedRef.current = onDisconnected }, [onDisconnected])

  const clearRetryTimer = useCallback(() => {
    if (retryTimerRef.current) {
      clearTimeout(retryTimerRef.current)
      retryTimerRef.current = null
    }
  }, [])

  const disconnect = useCallback(() => {
    clearRetryTimer()
    retryCountRef.current = MAX_RETRIES // prevent auto-retry
    if (wsRef.current) {
      wsRef.current.close(1000, 'user disconnect')
      wsRef.current = null
    }
    setStatus('disconnected')
  }, [clearRetryTimer])

  const connect = useCallback(() => {
    if (!mountedRef.current) return
    if (!auctionId)          return

    // close existing connection cleanly
    if (wsRef.current) {
      wsRef.current.onclose = null // suppress auto-retry on manual close
      wsRef.current.close()
      wsRef.current = null
    }
    const token=getAuthToken()!;
const params = new URLSearchParams({
  auction_id: auctionId,
  token: token ?? "",
})
    const url = `${WS_BASE_URL}/ws/auction?${params.toString()},  `

    setStatus('connecting')

    const ws = new WebSocket(url)
    wsRef.current = ws

    ws.onopen = () => {
      if (!mountedRef.current) return
      retryCountRef.current = 0
      setStatus('connected')
      onConnectedRef.current?.()
    }

    ws.onmessage = (event) => {
      if (!mountedRef.current) return
      try {
        const data: AuctionEvent = JSON.parse(event.data as string)
        setLastEvent(data)

        switch (data.type) {
          case 'new_bid':
            onNewBidRef.current?.(data)
            break
          case 'auction_ended':
          case 'auction_won':
          case 'auction_cancelled':
          case 'auction_no_reserve':
            onAuctionEndedRef.current?.(data)
            break
          case 'price_dropped':
            onPriceDroppedRef.current?.(data)
            break
          case 'extended':
            onExtendedRef.current?.(data)
            break
          case 'pong':
          case 'connected':
            break // intentionally silent
          default:
            break
        }
      } catch {
        // malformed message — ignore
      }
    }

    ws.onerror = () => {
      if (!mountedRef.current) return
      setStatus('error')
    }

    ws.onclose = (e) => {
      if (!mountedRef.current) return
      wsRef.current = null
      setStatus('disconnected')
      onDisconnectedRef.current?.()

      // normal close (1000) or user-triggered — no retry
      if (e.code === 1000) return

      // exponential backoff retry
      if (retryCountRef.current < MAX_RETRIES) {
        const delay = INITIAL_RETRY_MS * 2 ** retryCountRef.current
        retryCountRef.current += 1
        retryTimerRef.current = setTimeout(connect, delay)
      }
    }
  }, [auctionId, userId]) // eslint-disable-line react-hooks/exhaustive-deps

  const reconnect = useCallback(() => {
    retryCountRef.current = 0
    clearRetryTimer()
    connect()
  }, [connect, clearRetryTimer])

  // connect/disconnect when enabled or auctionId changes
  useEffect(() => {
    mountedRef.current = true

    if (enabled && auctionId) {
      connect()
    }

    return () => {
      mountedRef.current = false
      clearRetryTimer()
      if (wsRef.current) {
        wsRef.current.onclose = null
        wsRef.current.close(1000, 'component unmount')
        wsRef.current = null
      }
    }
  }, [auctionId, enabled]) // eslint-disable-line react-hooks/exhaustive-deps

  return {
    status,
    lastEvent,
    isConnected: status === 'connected',
    disconnect,
    reconnect,
  }
}