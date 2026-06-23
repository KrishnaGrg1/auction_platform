import { Link, useLocation, useNavigate } from '@tanstack/react-router'
import {
  Gavel,
  LayoutDashboard,
  ListChecks,
  PlusCircle,
  LogOut,
} from 'lucide-react'
import { Button } from '#/components/ui/button'

const navItems = [
  { to: '/dashboard', icon: LayoutDashboard, label: 'Overview' },
  { to: '/dashboard/auction', icon: ListChecks, label: 'All auctions' },
  {
    to: '/dashboard/auction/create',
    icon: PlusCircle,
    label: 'Create auction',
  },
]

export function DashboardLayout({ children }: { children: React.ReactNode }) {
  const pathname = useLocation().pathname
  const navigate = useNavigate()

  const handleLogout = () => {
    localStorage.removeItem('auth_token')
    navigate({ to: '/login' })
  }

  return (
    <div className="flex min-h-screen bg-muted/30">
      {/* ── Sidebar ── */}
      <aside className="hidden w-64 shrink-0 border-r bg-card md:flex flex-col">
        {/* Logo */}
        <div className="flex h-14 items-center gap-2.5 border-b px-5">
          <span
            className="flex size-8 items-center justify-center rounded-lg"
            style={{ background: 'var(--ink)' }}
          >
            <Gavel
              className="size-3.5"
              style={{ color: 'var(--amber-light)' }}
            />
          </span>
          <span
            className="text-sm font-semibold"
            style={{ fontFamily: 'var(--font-heading)' }}
          >
            Auction Platform
          </span>
        </div>

        {/* Nav */}
        <nav className="flex-1 space-y-1 p-3">
          {navItems.map((item) => {
            const Icon = item.icon
            const isActive =
              pathname === item.to || pathname.startsWith(item.to + '/')
            return (
              <Link
                key={item.to}
                to={item.to}
                className={`flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors ${
                  isActive
                    ? 'bg-amber/10 text-amber-foreground'
                    : 'text-muted-foreground hover:bg-muted hover:text-foreground'
                }`}
              >
                <Icon className="size-4" />
                {item.label}
              </Link>
            )
          })}
        </nav>

        {/* Sign out */}
        <div className="border-t p-3">
          <Button
            variant="ghost"
            size="sm"
            className="w-full justify-start gap-2 text-muted-foreground"
            onClick={handleLogout}
          >
            <LogOut className="size-4" />
            Sign out
          </Button>
        </div>
      </aside>

      {/* ── Content area ── */}
      <div className="flex flex-1 flex-col">
        {/* Mobile header */}
        <header className="flex h-14 items-center gap-3 border-b bg-card px-4 md:hidden">
          <span
            className="flex size-7 items-center justify-center rounded-md"
            style={{ background: 'var(--ink)' }}
          >
            <Gavel className="size-3" style={{ color: 'var(--amber-light)' }} />
          </span>
          <span className="text-sm font-semibold flex-1">Auction Platform</span>
          <Button
            variant="ghost"
            size="icon"
            className="size-8"
            onClick={handleLogout}
          >
            <LogOut className="size-4" />
          </Button>
        </header>

        <main className="flex-1 p-4 md:p-6 lg:p-8">{children}</main>
      </div>
    </div>
  )
}
