import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { auth, api } from '../api/client'

export function Login() {
  const [token, setToken] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    setLoading(true)
    auth.setToken(token)
    try {
      await api.members.list()
      navigate('/members')
    } catch {
      auth.clearToken()
      setError('ERR: AUTHENTICATION FAILED — INVALID TOKEN')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div
      className="min-h-screen flex flex-col items-center justify-center"
      style={{ background: 'var(--c-bg)' }}
    >
      <div className="w-full max-w-sm mx-4">
        {/* Title */}
        <div className="text-center mb-8">
          <div className="text-xs tracking-widest mb-1" style={{ color: 'var(--c-muted)' }}>
            // SECURE ACCESS TERMINAL
          </div>
          <h1
            className="text-4xl tracking-widest uppercase"
            style={{ color: 'var(--c-cyan)', textShadow: '0 0 20px rgba(0,255,225,0.6)' }}
          >
            ANGELIX
          </h1>
        </div>

        {/* Panel */}
        <div
          style={{
            background: 'var(--c-surface)',
            border: '1px solid var(--c-border)',
            boxShadow: '0 0 32px rgba(0,255,225,0.08)',
            padding: '2rem',
            position: 'relative',
          }}
        >
          {/* Corner brackets */}
          <span className="absolute top-0 left-0 text-sm leading-none select-none" style={{ color: 'var(--c-cyan)', padding: '4px' }}>┌</span>
          <span className="absolute top-0 right-0 text-sm leading-none select-none" style={{ color: 'var(--c-cyan)', padding: '4px' }}>┐</span>
          <span className="absolute bottom-0 left-0 text-sm leading-none select-none" style={{ color: 'var(--c-cyan)', padding: '4px' }}>└</span>
          <span className="absolute bottom-0 right-0 text-sm leading-none select-none" style={{ color: 'var(--c-cyan)', padding: '4px' }}>┘</span>

          <div className="text-xs tracking-widest mb-6" style={{ color: 'var(--c-cyan)' }}>
            SYSTEM AUTHENTICATION
          </div>

          <form onSubmit={handleSubmit} className="space-y-5">
            <div>
              <label className="hud-label">Admin Token</label>
              <input
                type="password"
                value={token}
                onChange={e => setToken(e.target.value)}
                required
                className="hud-input"
                placeholder="••••••••••••"
              />
            </div>

            {error && (
              <p className="text-xs tracking-wide" style={{ color: 'var(--c-magenta)' }}>
                {error}
              </p>
            )}

            <button
              type="submit"
              disabled={loading || !token}
              className="hud-btn w-full"
            >
              {loading ? '> VERIFYING...' : '> AUTHENTICATE'}
            </button>
          </form>
        </div>

        <div className="mt-4 text-center text-xs tracking-widest" style={{ color: 'var(--c-muted)' }}>
          AUTHORIZED ACCESS ONLY
        </div>
      </div>
    </div>
  )
}
