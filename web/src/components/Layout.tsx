import { NavLink, Outlet, useNavigate } from 'react-router-dom'
import { auth } from '../api/client'

const navItems = [
  { to: '/members', label: 'Members' },
  { to: '/metrics', label: 'Metrics' },
]

export function Layout() {
  const navigate = useNavigate()

  function signOut() {
    auth.clearToken()
    navigate('/login')
  }

  return (
    <div className="flex h-screen bg-slate-50">
      {/* Sidebar */}
      <aside className="w-56 shrink-0 bg-slate-900 flex flex-col">
        <div className="px-5 py-5 border-b border-slate-800">
          <span className="text-white font-semibold tracking-wide">Angelix</span>
        </div>
        <nav className="flex-1 px-3 py-4 space-y-1">
          {navItems.map(({ to, label }) => (
            <NavLink
              key={to}
              to={to}
              className={({ isActive }) =>
                `block px-3 py-2 rounded-lg text-sm font-medium transition-colors ${
                  isActive
                    ? 'bg-indigo-600 text-white'
                    : 'text-slate-400 hover:bg-slate-800 hover:text-white'
                }`
              }
            >
              {label}
            </NavLink>
          ))}
        </nav>
        <div className="px-3 py-4 border-t border-slate-800">
          <button
            onClick={signOut}
            className="w-full text-left px-3 py-2 text-sm text-slate-400 hover:text-white hover:bg-slate-800 rounded-lg transition-colors"
          >
            Sign out
          </button>
        </div>
      </aside>

      {/* Main */}
      <main className="flex-1 overflow-auto">
        <Outlet />
      </main>
    </div>
  )
}
