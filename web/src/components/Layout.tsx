import { NavLink, Outlet, useNavigate } from 'react-router-dom'
import { auth } from '../api/client'

const navItems = [
  { to: '/members',     label: 'MEMBERS',     idx: '01' },
  { to: '/credentials', label: 'CREDENTIALS', idx: '02' },
  { to: '/metrics',     label: 'METRICS',     idx: '03' },
]

export function Layout() {
  const navigate = useNavigate()

  function signOut() {
    auth.clearToken()
    navigate('/login')
  }

  return (
    <div className="flex h-screen" style={{ background: 'var(--c-bg)' }}>
      <aside
        className="w-52 shrink-0 flex flex-col"
        style={{
          background: 'var(--c-surface)',
          borderRight: '1px solid var(--c-border)',
          boxShadow: '4px 0 24px rgba(0,255,225,0.04)',
        }}
      >
        {/* Logo */}
        <div className="px-5 py-5" style={{ borderBottom: '1px solid var(--c-border)' }}>
          <div className="text-xs tracking-widest" style={{ color: 'var(--c-muted)' }}>// SYSTEM</div>
          <div
            className="text-xl tracking-widest uppercase mt-0.5"
            style={{ color: 'var(--c-cyan)', textShadow: '0 0 12px rgba(0,255,225,0.55)' }}
          >
            ANGELIX
          </div>
        </div>

        {/* Nav */}
        <nav className="flex-1 py-3">
          {navItems.map(({ to, label, idx }) => (
            <NavLink
              key={to}
              to={to}
              style={({ isActive }) => ({
                display: 'flex',
                alignItems: 'center',
                gap: '0.5rem',
                padding: '0.55rem 1.25rem',
                fontSize: '0.75rem',
                letterSpacing: '0.12em',
                textDecoration: 'none',
                color: isActive ? 'var(--c-cyan)' : 'var(--c-muted)',
                background: isActive ? 'rgba(0,255,225,0.07)' : 'transparent',
                borderLeft: isActive ? '2px solid var(--c-cyan)' : '2px solid transparent',
                transition: 'all 0.15s',
              })}
            >
              <span style={{ fontSize: '0.6rem', color: 'var(--c-muted)' }}>{idx}</span>
              {label}
            </NavLink>
          ))}
        </nav>

        {/* Exit */}
        <div className="px-4 py-4" style={{ borderTop: '1px solid var(--c-border)' }}>
          <button
            onClick={signOut}
            className="hud-btn-danger w-full"
            style={{ padding: '0.4rem 0.75rem' }}
          >
            [EXIT SESSION]
          </button>
        </div>
      </aside>

      <main className="flex-1 overflow-auto" style={{ background: 'var(--c-bg)' }}>
        <Outlet />
      </main>
    </div>
  )
}
