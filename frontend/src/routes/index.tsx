import { createFileRoute, Link } from '@tanstack/react-router'
import {
  BadgeDollarSign,
  Clock3,
  Gavel,
  ShieldCheck,
  Sparkles,
  TrendingUp,
} from 'lucide-react'
import { Button } from '#/components/ui/button'

export const Route = createFileRoute('/')({ component: Home })

const liveLots = [
  {
    title: 'Vintage Chronograph',
    bidder: 'N. Rai',
    bid: '$1,840',
    status: 'Live',
    statusClass: 'status-live',
  },
  {
    title: 'Studio Walnut Desk',
    bidder: 'M. Chen',
    bid: '$920',
    status: '12 bids',
    statusClass: 'status-count',
  },
  {
    title: 'Collector Camera Kit',
    bidder: 'A. Singh',
    bid: '$2,450',
    status: 'Closing',
    statusClass: 'status-closing',
  },
]

const highlights = [
  {
    icon: Clock3,
    title: 'Timed auctions',
    text: 'Schedule listings, extend active lots, and keep bidders focused until close.',
  },
  {
    icon: TrendingUp,
    title: 'Live bid tracking',
    text: 'Follow current price, bidder movement, and auction status from one workspace.',
  },
  {
    icon: ShieldCheck,
    title: 'Verified accounts',
    text: 'Protect buyer and seller activity with account verification and secure sessions.',
  },
]

function Home() {
  return (
    <main className="min-h-screen" style={{ background: 'var(--surface)' }}>
      {/* ── Header ── */}
      <header
        className="page-wrap flex items-center justify-between py-4"
        style={{ borderBottom: '0.5px solid var(--line)' }}
      >
        <Link to="/" className="flex items-center gap-3 no-underline">
          <span
            className="flex size-9 items-center justify-center rounded-lg"
            style={{ background: 'var(--ink)' }}
          >
            <Gavel className="size-4" style={{ color: 'var(--amber-light)' }} />
          </span>
          <span
            style={{
              fontFamily: 'var(--font-heading)',
              fontWeight: 600,
              fontSize: '0.9375rem',
              color: 'var(--ink)',
            }}
          >
            Auction Platform
          </span>
        </Link>

        <nav className="hidden items-center gap-6 md:flex">
          <a className="nav-link" href="#live">
            Live lots
          </a>
          <a className="nav-link" href="#tools">
            Tools
          </a>
          <Link className="nav-link" to="/login">
            Sign in
          </Link>
        </nav>

        <Button
          asChild
          size="sm"
          style={{
            background: 'var(--ink)',
            color: 'var(--amber-light)',
            fontFamily: 'var(--font-heading)',
          }}
          className="hover:opacity-90 transition-opacity"
        >
          <Link to="/register">Create account</Link>
        </Button>
      </header>

      {/* ── Hero ── */}
      <section className="page-wrap grid items-start gap-8 pt-10 pb-16 lg:grid-cols-[1.1fr_0.9fr] lg:pt-14 lg:pb-20">
        {/* Left */}
        <div className="rise-in flex flex-col">
          <p className="island-kicker mb-3">Buyer and seller marketplace</p>

          <h1
            className="display-title"
            style={{
              fontSize: 'clamp(2.2rem, 4.5vw, 3.5rem)',
              maxWidth: '680px',
            }}
          >
            Run trusted auctions from listing to final bid.
          </h1>

          <p
            className="mt-5 max-w-[520px] leading-7"
            style={{
              color: 'var(--ink-soft)',
              fontFamily: 'var(--font-heading)',
              fontSize: '1rem',
            }}
          >
            Launch timed lots, accept competitive bids, and give every seller a
            clear view of auction activity.
          </p>

          {/* Metrics — ABOVE CTAs for visual weight */}
          <div className="mt-7 grid max-w-[460px] grid-cols-3 gap-2.5">
            <Metric value="2.4k" label="active bids" />
            <Metric value="98%" label="verified users" />
            <Metric value="24/7" label="lot tracking" />
          </div>

          {/* CTAs */}
          <div className="mt-6 flex flex-col gap-3 sm:flex-row">
            <Button
              asChild
              size="lg"
              style={{
                background: 'var(--ink)',
                color: 'var(--amber-light)',
                fontFamily: 'var(--font-heading)',
              }}
              className="hover:opacity-90 transition-opacity"
            >
              <Link to="/register">Start auctioning</Link>
            </Button>
            <Button
              asChild
              variant="outline"
              size="lg"
              style={{
                borderColor: 'var(--line-strong)',
                color: 'var(--ink)',
                fontFamily: 'var(--font-heading)',
              }}
              className="transition-colors hover:bg-[var(--amber-bg)]"
            >
              <Link to="/login">View dashboard</Link>
            </Button>
          </div>
        </div>

        {/* Right — Live auction card */}
        <div
          id="live"
          className="rise-in rounded-xl"
          style={{
            background: 'var(--card-bg)',
            border: '0.5px solid var(--line)',
            boxShadow: 'var(--shadow-lg)',
            animationDelay: '80ms',
          }}
        >
          {/* Card header */}
          <div
            className="flex items-center justify-between gap-4 px-5 py-4"
            style={{ borderBottom: '0.5px solid var(--line)' }}
          >
            <div>
              <p className="island-kicker mb-0.5">Live auction desk</p>
              <h2
                className="text-lg font-bold"
                style={{
                  fontFamily: 'var(--font-heading)',
                  color: 'var(--ink)',
                  lineHeight: 1.2,
                }}
              >
                Current bidding
              </h2>
            </div>
            <span className="badge-amber badge-live">Open</span>
          </div>

          {/* Lot rows */}
          <div className="px-5 py-3 flex flex-col gap-0">
            {liveLots.map((lot, i) => (
              <div
                key={lot.title}
                className="grid grid-cols-[1fr_auto] items-center gap-4 py-3.5"
                style={{
                  borderBottom:
                    i < liveLots.length - 1
                      ? '0.5px solid var(--line)'
                      : 'none',
                }}
              >
                <div className="min-w-0">
                  <p
                    className="truncate text-sm font-semibold"
                    style={{
                      color: 'var(--ink)',
                      fontFamily: 'var(--font-heading)',
                    }}
                  >
                    {lot.title}
                  </p>
                  <p
                    className="mt-0.5 text-xs"
                    style={{
                      color: 'var(--ink-muted)',
                      fontFamily: 'var(--font-heading)',
                    }}
                  >
                    Top bidder: {lot.bidder}
                  </p>
                </div>
                <div className="text-right shrink-0">
                  <p
                    className="text-base font-extrabold"
                    style={{
                      color: 'var(--ink)',
                      fontFamily: 'var(--font-heading)',
                    }}
                  >
                    {lot.bid}
                  </p>
                  <p className={`text-xs font-semibold ${lot.statusClass}`}>
                    {lot.status}
                  </p>
                </div>
              </div>
            ))}
          </div>

          {/* Stats footer */}
          <div
            className="grid grid-cols-2 gap-0"
            style={{ borderTop: '0.5px solid var(--line)' }}
          >
            <div
              className="flex flex-col gap-1 px-5 py-4"
              style={{
                background: 'var(--ink)',
                borderBottomLeftRadius: '0.75rem',
              }}
            >
              <BadgeDollarSign
                className="size-4 mb-1"
                style={{ color: 'var(--amber)' }}
              />
              <p
                className="text-xl font-extrabold"
                style={{
                  color: 'var(--amber-light)',
                  fontFamily: 'var(--font-heading)',
                }}
              >
                $18.6k
              </p>
              <p
                className="text-xs"
                style={{
                  color: 'rgba(250,248,244,0.45)',
                  fontFamily: 'var(--font-heading)',
                }}
              >
                volume today
              </p>
            </div>
            <div
              className="flex flex-col gap-1 px-5 py-4"
              style={{
                background: 'var(--amber-bg)',
                borderLeft: '0.5px solid var(--amber-border)',
                borderBottomRightRadius: '0.75rem',
              }}
            >
              <Sparkles
                className="size-4 mb-1"
                style={{ color: 'var(--amber)' }}
              />
              <p
                className="text-xl font-extrabold"
                style={{
                  color: 'var(--ink)',
                  fontFamily: 'var(--font-heading)',
                }}
              >
                36
              </p>
              <p
                className="text-xs"
                style={{
                  color: 'var(--ink-soft)',
                  fontFamily: 'var(--font-heading)',
                }}
              >
                lots closing
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* ── Divider ── */}
      <div
        className="page-wrap"
        style={{ borderTop: '0.5px solid var(--line)' }}
      />

      {/* ── Features ── */}
      <section id="tools" className="page-wrap py-16 pb-24">
        <div className="mb-10 flex flex-col justify-between gap-4 md:flex-row md:items-end">
          <div>
            <p className="island-kicker mb-3">Marketplace controls</p>
            <h2
              className="display-title"
              style={{ fontSize: 'clamp(1.6rem, 2.8vw, 2.25rem)' }}
            >
              Built for repeated auction work.
            </h2>
          </div>
          <p
            className="max-w-[400px] text-sm leading-6"
            style={{
              color: 'var(--ink-soft)',
              fontFamily: 'var(--font-heading)',
            }}
          >
            Keep buyer flow, seller operations, and lot status readable without
            hiding important activity.
          </p>
        </div>

        <div className="grid gap-4 md:grid-cols-3">
          {highlights.map((item) => {
            const Icon = item.icon
            return (
              <article
                key={item.title}
                className="feature-card rounded-xl p-6"
                style={{ background: 'var(--card-bg)' }}
              >
                <div
                  className="mb-4 flex size-10 items-center justify-center rounded-lg"
                  style={{ background: 'var(--amber-bg)' }}
                >
                  <Icon className="size-5" style={{ color: 'var(--amber)' }} />
                </div>
                <h3
                  className="text-base font-bold"
                  style={{
                    fontFamily: 'var(--font-heading)',
                    color: 'var(--ink)',
                  }}
                >
                  {item.title}
                </h3>
                <p
                  className="mt-2.5 text-sm leading-6"
                  style={{
                    color: 'var(--ink-soft)',
                    fontFamily: 'var(--font-heading)',
                  }}
                >
                  {item.text}
                </p>
              </article>
            )
          })}
        </div>
      </section>
    </main>
  )
}

function Metric({ value, label }: { value: string; label: string }) {
  return (
    <div className="metric-card">
      <p className="metric-value">{value}</p>
      <p className="metric-label">{label}</p>
    </div>
  )
}
