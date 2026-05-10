import { useEffect, useState } from 'react'
import { api } from '../api/client'
import type { MemberResponse } from '../api/types'
import { Modal } from '../components/Modal'

export function Members() {
  const [members, setMembers] = useState<MemberResponse[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const [showCreate, setShowCreate] = useState(false)
  const [form, setForm] = useState({ name: '', email: '' })
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
      setForm({ name: '', email: '' })
      setNewToken(m.token)
    } catch (e) {
      setCreateError(String(e))
    } finally {
      setCreating(false)
    }
  }

  async function handleDelete(id: string, name: string) {
    if (!confirm(`CONFIRM: Remove member "${name}" from registry?`)) return
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
      {/* Header */}
      <div className="flex items-end justify-between mb-6">
        <div>
          <div className="text-xs tracking-widest" style={{ color: 'var(--c-muted)' }}>// DATABASE</div>
          <h1 className="text-xl tracking-widest uppercase" style={{ color: 'var(--c-cyan)' }}>
            MEMBER REGISTRY
          </h1>
        </div>
        <button onClick={() => setShowCreate(true)} className="hud-btn">
          &gt;_ ENLIST MEMBER
        </button>
      </div>

      {loading && (
        <p className="text-sm tracking-widest" style={{ color: 'var(--c-muted)' }}>
          {'> LOADING...'}
        </p>
      )}
      {error && (
        <p className="text-sm" style={{ color: 'var(--c-magenta)' }}>
          ERR: {error}
        </p>
      )}

      {!loading && !error && (
        <div
          style={{
            background: 'var(--c-surface)',
            border: '1px solid var(--c-border)',
            boxShadow: '0 0 20px rgba(0,255,225,0.05)',
            overflow: 'hidden',
          }}
        >
          <table className="w-full text-sm">
            <thead>
              <tr
                style={{
                  borderBottom: '1px solid var(--c-border)',
                  background: 'rgba(0,255,225,0.04)',
                }}
              >
                {['IDENT', 'EMAIL', 'TOKEN', 'ENROLLED', 'OPS'].map(h => (
                  <th
                    key={h}
                    className="text-left px-4 py-3"
                    style={{ fontSize: '0.65rem', letterSpacing: '0.14em', color: 'var(--c-cyan)' }}
                  >
                    {h}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {members.length === 0 && (
                <tr>
                  <td
                    colSpan={5}
                    className="px-4 py-10 text-center text-xs tracking-widest"
                    style={{ color: 'var(--c-muted)' }}
                  >
                    // NO MEMBERS REGISTERED — ENLIST TO BEGIN
                  </td>
                </tr>
              )}
              {members.map((m, i) => (
                <MemberRow
                  key={m.id}
                  member={m}
                  even={i % 2 === 0}
                  onDelete={() => handleDelete(m.id, m.name)}
                />
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Create modal */}
      {showCreate && (
        <Modal title="ENLIST MEMBER" onClose={() => setShowCreate(false)}>
          <form onSubmit={handleCreate} className="space-y-4">
            {([
              { field: 'name',  label: 'CALLSIGN',     type: 'text' },
              { field: 'email', label: 'COMM ADDRESS', type: 'text' },
            ] as const).map(({ field, label, type }) => (
              <div key={field}>
                <label className="hud-label">{label}</label>
                <input
                  type={type}
                  value={form[field]}
                  onChange={e => setForm(f => ({ ...f, [field]: e.target.value }))}
                  required
                  className="hud-input"
                />
              </div>
            ))}
            {createError && (
              <p className="text-xs" style={{ color: 'var(--c-magenta)' }}>
                ERR: {createError}
              </p>
            )}
            <div className="flex justify-end gap-3 pt-2">
              <button
                type="button"
                onClick={() => setShowCreate(false)}
                className="hud-btn-danger"
                style={{ padding: '0.35rem 0.9rem' }}
              >
                [ABORT]
              </button>
              <button type="submit" disabled={creating} className="hud-btn">
                {creating ? '> PROCESSING...' : '> CONFIRM'}
              </button>
            </div>
          </form>
        </Modal>
      )}

      {/* Token reveal modal */}
      {newToken && (
        <Modal title="ACCESS TOKEN ISSUED" onClose={() => setNewToken(null)}>
          <p className="text-xs tracking-wide mb-4" style={{ color: 'var(--c-muted)' }}>
            // STORE SECURELY — TOKEN WILL NOT BE SHOWN AGAIN
          </p>
          <div
            className="flex items-center gap-3 mb-5 px-3 py-2"
            style={{
              background: 'rgba(0,255,225,0.04)',
              border: '1px solid var(--c-border)',
            }}
          >
            <code className="flex-1 text-xs break-all" style={{ color: 'var(--c-cyan)' }}>
              {newToken}
            </code>
            <button
              onClick={copyToken}
              className="hud-btn shrink-0"
              style={{ padding: '0.2rem 0.6rem' }}
            >
              {copied ? '[COPIED]' : '[COPY]'}
            </button>
          </div>
          <button onClick={() => setNewToken(null)} className="hud-btn w-full">
            &gt; CLOSE
          </button>
        </Modal>
      )}
    </div>
  )
}

function MemberRow({
  member: m,
  even,
  onDelete,
}: {
  member: MemberResponse
  even: boolean
  onDelete: () => void
}) {
  const [hovered, setHovered] = useState(false)
  const base = even ? 'transparent' : 'rgba(0,255,225,0.018)'

  return (
    <tr
      style={{
        borderBottom: '1px solid rgba(0,255,225,0.07)',
        background: hovered ? 'rgba(0,255,225,0.05)' : base,
        transition: 'background 0.12s',
      }}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
    >
      <td className="px-4 py-3" style={{ color: 'var(--c-text)' }}>
        <span style={{ color: 'var(--c-cyan)', marginRight: '0.4rem' }}>◆</span>
        {m.name}
      </td>
      <td className="px-4 py-3 text-xs" style={{ color: 'var(--c-muted)' }}>{m.email}</td>
      <td className="px-4 py-3">
        <TokenCell token={m.token} />
      </td>
      <td className="px-4 py-3 text-xs" style={{ color: 'var(--c-muted)' }}>
        {new Date(m.created_at).toLocaleDateString()}
      </td>
      <td className="px-4 py-3 text-right">
        <button
          onClick={onDelete}
          className="hud-btn-danger"
          style={{ padding: '0.2rem 0.55rem', fontSize: '0.65rem' }}
        >
          [REMOVE]
        </button>
      </td>
    </tr>
  )
}

function TokenCell({ token }: { token: string }) {
  const [visible, setVisible] = useState(false)
  const masked = token.slice(0, 8) + '••••••••'
  return (
    <span className="text-xs" style={{ color: 'var(--c-muted)' }}>
      <span style={{ color: 'var(--c-text)' }}>{visible ? token : masked}</span>
      <button
        onClick={() => setVisible(v => !v)}
        className="ml-2 text-xs"
        style={{
          color: 'var(--c-cyan)',
          background: 'none',
          border: 'none',
          cursor: 'pointer',
          padding: 0,
        }}
      >
        [{visible ? 'HIDE' : 'SHOW'}]
      </button>
    </span>
  )
}
