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
      setError('Invalid admin token.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-slate-50 flex items-center justify-center">
      <div className="bg-white rounded-xl shadow-sm border border-slate-200 p-8 w-full max-w-sm">
        <h1 className="text-xl font-semibold text-slate-900 mb-1">Angelix</h1>
        <p className="text-sm text-slate-500 mb-6">Sign in with your admin token.</p>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">
              Admin token
            </label>
            <input
              type="password"
              value={token}
              onChange={e => setToken(e.target.value)}
              required
              className="w-full border border-slate-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
              placeholder="••••••••"
            />
          </div>
          {error && <p className="text-sm text-red-600">{error}</p>}
          <button
            type="submit"
            disabled={loading || !token}
            className="w-full bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 text-white text-sm font-medium py-2 rounded-lg transition-colors"
          >
            {loading ? 'Verifying…' : 'Sign in'}
          </button>
        </form>
      </div>
    </div>
  )
}
