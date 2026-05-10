import { useEffect, useState } from 'react'
import { api } from '../api/client'
import type { MemberResponse } from '../api/types'
import { Modal } from '../components/Modal'

export function Members() {
  const [members, setMembers] = useState<MemberResponse[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const [showCreate, setShowCreate] = useState(false)
  const [form, setForm] = useState({ name: '', email: '', api_key: '' })
  const [creating, setCreating] = useState(false)
  const [createError, setCreateError] = useState('')

  const [newToken, setNewToken] = useState<string | null>(null)
  const [copied, setCopied] = useState(false)

  async function load() {
    try {
      const res = await api.members.list()
      setMembers(res.members ?? [])
    } catch (e) {
      setError(String(e))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [])

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    setCreateError('')
    setCreating(true)
    try {
      const m = await api.members.create(form)
      setMembers(prev => [m, ...prev])
      setShowCreate(false)
      setForm({ name: '', email: '', api_key: '' })
      setNewToken(m.token)
    } catch (e) {
      setCreateError(String(e))
    } finally {
      setCreating(false)
    }
  }

  async function handleDelete(id: string, name: string) {
    if (!confirm(`Delete member "${name}"? This cannot be undone.`)) return
    await api.members.delete(id)
    setMembers(prev => prev.filter(m => m.id !== id))
  }

  function copyToken() {
    if (!newToken) return
    navigator.clipboard.writeText(newToken)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="p-8">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-lg font-semibold text-slate-900">Members</h1>
        <button
          onClick={() => setShowCreate(true)}
          className="bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-medium px-4 py-2 rounded-lg transition-colors"
        >
          Add member
        </button>
      </div>

      {loading && <p className="text-sm text-slate-500">Loading…</p>}
      {error && <p className="text-sm text-red-600">{error}</p>}

      {!loading && !error && (
        <div className="bg-white rounded-xl border border-slate-200 overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-slate-50 border-b border-slate-200">
              <tr>
                {['Name', 'Email', 'Token', 'Created', ''].map(h => (
                  <th key={h} className="text-left px-4 py-3 font-medium text-slate-600 text-xs uppercase tracking-wide">
                    {h}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {members.length === 0 && (
                <tr>
                  <td colSpan={5} className="px-4 py-8 text-center text-slate-400">
                    No members yet. Add one to get started.
                  </td>
                </tr>
              )}
              {members.map(m => (
                <tr key={m.id} className="hover:bg-slate-50">
                  <td className="px-4 py-3 font-medium text-slate-900">{m.name}</td>
                  <td className="px-4 py-3 text-slate-600">{m.email}</td>
                  <td className="px-4 py-3">
                    <TokenCell token={m.token} />
                  </td>
                  <td className="px-4 py-3 text-slate-500">
                    {new Date(m.created_at).toLocaleDateString()}
                  </td>
                  <td className="px-4 py-3 text-right">
                    <button
                      onClick={() => handleDelete(m.id, m.name)}
                      className="text-xs text-red-600 hover:text-red-800 font-medium"
                    >
                      Delete
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Create modal */}
      {showCreate && (
        <Modal title="Add member" onClose={() => setShowCreate(false)}>
          <form onSubmit={handleCreate} className="space-y-4">
            {(['name', 'email', 'api_key'] as const).map(field => (
              <div key={field}>
                <label className="block text-sm font-medium text-slate-700 mb-1 capitalize">
                  {field === 'api_key' ? 'Anthropic API key' : field}
                </label>
                <input
                  type={field === 'api_key' ? 'password' : 'text'}
                  value={form[field]}
                  onChange={e => setForm(f => ({ ...f, [field]: e.target.value }))}
                  required
                  className="w-full border border-slate-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                />
              </div>
            ))}
            {createError && <p className="text-sm text-red-600">{createError}</p>}
            <div className="flex justify-end gap-2 pt-1">
              <button
                type="button"
                onClick={() => setShowCreate(false)}
                className="px-4 py-2 text-sm text-slate-600 hover:text-slate-900"
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={creating}
                className="bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 text-white text-sm font-medium px-4 py-2 rounded-lg transition-colors"
              >
                {creating ? 'Creating…' : 'Create'}
              </button>
            </div>
          </form>
        </Modal>
      )}

      {/* Token reveal modal */}
      {newToken && (
        <Modal title="Member created" onClose={() => setNewToken(null)}>
          <p className="text-sm text-slate-600 mb-3">
            Save this token — it won't be shown again.
          </p>
          <div className="flex items-center gap-2 bg-slate-50 border border-slate-200 rounded-lg px-3 py-2 mb-4">
            <code className="flex-1 text-xs text-slate-800 break-all">{newToken}</code>
            <button
              onClick={copyToken}
              className="shrink-0 text-xs text-indigo-600 hover:text-indigo-800 font-medium"
            >
              {copied ? 'Copied!' : 'Copy'}
            </button>
          </div>
          <button
            onClick={() => setNewToken(null)}
            className="w-full bg-slate-900 hover:bg-slate-800 text-white text-sm font-medium py-2 rounded-lg transition-colors"
          >
            Done
          </button>
        </Modal>
      )}
    </div>
  )
}

function TokenCell({ token }: { token: string }) {
  const [visible, setVisible] = useState(false)
  const masked = token.slice(0, 8) + '••••••••'
  return (
    <span className="font-mono text-xs text-slate-500">
      {visible ? token : masked}
      <button
        onClick={() => setVisible(v => !v)}
        className="ml-2 text-indigo-500 hover:text-indigo-700"
      >
        {visible ? 'hide' : 'show'}
      </button>
    </span>
  )
}
