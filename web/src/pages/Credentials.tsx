import { useEffect, useState } from 'react'
import { api } from '../api/client'
import type { CredentialItem, MemberResponse } from '../api/types'
import { Modal } from '../components/Modal'

export function Credentials() {
  const [creds, setCreds] = useState<CredentialItem[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const [showCreate, setShowCreate] = useState(false)
  const [form, setForm] = useState({ name: '', access_token: '', refresh_token: '' })
  const [creating, setCreating] = useState(false)
  const [createError, setCreateError] = useState('')

  const [managing, setManaging] = useState<CredentialItem | null>(null)

  async function load() {
    try {
      const res = await api.catalog.list()
      setCreds(res.credentials ?? [])
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
      const c = await api.catalog.create(form)
      setCreds(prev => [...prev, c])
      setShowCreate(false)
      setForm({ name: '', access_token: '', refresh_token: '' })
    } catch (e) {
      setCreateError(String(e))
    } finally {
      setCreating(false)
    }
  }

  async function handleDelete(id: string, name: string) {
    if (!confirm(`CONFIRM: Delete credential "${name}"? All assignments will be removed.`)) return
    await api.catalog.delete(id)
    setCreds(prev => prev.filter(c => c.id !== id))
  }

  return (
    <div className="p-8">
      <div className="flex items-end justify-between mb-6">
        <div>
          <div className="text-xs tracking-widest" style={{ color: 'var(--c-muted)' }}>// CATALOG</div>
          <h1 className="text-xl tracking-widest uppercase" style={{ color: 'var(--c-cyan)' }}>
            CREDENTIALS
          </h1>
        </div>
        <button onClick={() => setShowCreate(true)} className="hud-btn">
          &gt;_ ADD CREDENTIAL
        </button>
      </div>

      {loading && (
        <p className="text-xs tracking-widest" style={{ color: 'var(--c-muted)' }}>{'> LOADING...'}</p>
      )}
      {error && (
        <p className="text-xs" style={{ color: 'var(--c-magenta)' }}>ERR: {error}</p>
      )}

      {!loading && !error && (
        <div style={{
          background: 'var(--c-surface)',
          border: '1px solid var(--c-border)',
          boxShadow: '0 0 20px rgba(0,255,225,0.05)',
          overflow: 'hidden',
        }}>
          <table className="w-full text-sm">
            <thead>
              <tr style={{ borderBottom: '1px solid var(--c-border)', background: 'rgba(0,255,225,0.04)' }}>
                {['IDENT', 'NAME', 'CREATED', 'ASSIGNMENTS', 'OPS'].map(h => (
                  <th key={h} className="text-left px-4 py-3"
                    style={{ fontSize: '0.65rem', letterSpacing: '0.14em', color: 'var(--c-cyan)' }}>
                    {h}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {creds.length === 0 && (
                <tr>
                  <td colSpan={5} className="px-4 py-10 text-center text-xs tracking-widest"
                    style={{ color: 'var(--c-muted)' }}>
                    // NO CREDENTIALS — ADD ONE TO BEGIN
                  </td>
                </tr>
              )}
              {creds.map((c, i) => (
                <CredRow
                  key={c.id}
                  cred={c}
                  even={i % 2 === 0}
                  onDelete={() => handleDelete(c.id, c.name)}
                  onManage={() => setManaging(c)}
                />
              ))}
            </tbody>
          </table>
        </div>
      )}

      {showCreate && (
        <Modal title="ADD CREDENTIAL" onClose={() => setShowCreate(false)}>
          <form onSubmit={handleCreate} className="space-y-4">
            <div>
              <label className="hud-label">NAME</label>
              <input
                type="text"
                value={form.name}
                onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
                required
                className="hud-input"
                placeholder="e.g. team-prod"
              />
            </div>
            <div>
              <label className="hud-label">ACCESS TOKEN</label>
              <input
                type="password"
                value={form.access_token}
                onChange={e => setForm(f => ({ ...f, access_token: e.target.value }))}
                required
                className="hud-input"
                placeholder="OAuth access token"
              />
            </div>
            <div>
              <label className="hud-label">REFRESH TOKEN <span style={{ color: 'var(--c-muted)', fontSize: '0.6rem' }}>(OPTIONAL)</span></label>
              <input
                type="password"
                value={form.refresh_token}
                onChange={e => setForm(f => ({ ...f, refresh_token: e.target.value }))}
                className="hud-input"
                placeholder="OAuth refresh token"
              />
            </div>
            {createError && (
              <p className="text-xs" style={{ color: 'var(--c-magenta)' }}>ERR: {createError}</p>
            )}
            <div className="flex justify-end gap-3 pt-2">
              <button type="button" onClick={() => setShowCreate(false)}
                className="hud-btn-danger" style={{ padding: '0.35rem 0.9rem' }}>
                [ABORT]
              </button>
              <button type="submit" disabled={creating} className="hud-btn">
                {creating ? '> SAVING...' : '> CONFIRM'}
              </button>
            </div>
          </form>
        </Modal>
      )}

      {managing && (
        <AssignModal
          cred={managing}
          onClose={() => setManaging(null)}
        />
      )}
    </div>
  )
}

function CredRow({
  cred: c,
  even,
  onDelete,
  onManage,
}: {
  cred: CredentialItem
  even: boolean
  onDelete: () => void
  onManage: () => void
}) {
  const [hovered, setHovered] = useState(false)
  const [assignCount, setAssignCount] = useState<number | null>(null)
  const base = even ? 'transparent' : 'rgba(0,255,225,0.018)'

  useEffect(() => {
    api.members.listCredentials(c.id).catch(() => null)
    // count assignments via member list
    api.catalog.list().catch(() => null)
  }, [c.id])

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
      <td className="px-4 py-3 text-xs" style={{ color: 'var(--c-muted)', fontFamily: 'monospace' }}>
        {c.id.slice(0, 8)}
      </td>
      <td className="px-4 py-3" style={{ color: 'var(--c-text)' }}>
        <span style={{ color: 'var(--c-cyan)', marginRight: '0.4rem' }}>◆</span>
        {c.name}
      </td>
      <td className="px-4 py-3 text-xs" style={{ color: 'var(--c-muted)' }}>
        {new Date(c.created_at).toLocaleDateString()}
      </td>
      <td className="px-4 py-3">
        <button onClick={onManage} className="hud-btn" style={{ padding: '0.2rem 0.6rem', fontSize: '0.65rem' }}>
          [MANAGE]
        </button>
      </td>
      <td className="px-4 py-3 text-right">
        <button onClick={onDelete} className="hud-btn-danger"
          style={{ padding: '0.2rem 0.55rem', fontSize: '0.65rem' }}>
          [REMOVE]
        </button>
      </td>
    </tr>
  )
}

function AssignModal({ cred, onClose }: { cred: CredentialItem; onClose: () => void }) {
  const [allMembers, setAllMembers] = useState<MemberResponse[]>([])
  const [assigned, setAssigned] = useState<CredentialItem[]>([])
  const [selectedMember, setSelectedMember] = useState('')
  const [loading, setLoading] = useState(true)
  const [assigning, setAssigning] = useState(false)
  const [error, setError] = useState('')

  async function load() {
    setLoading(true)
    try {
      const [membersRes, assignedRes] = await Promise.all([
        api.members.list(),
        // we re-use listCredentials per member to find who has this cred;
        // simpler: load all members and check each — but that's N+1.
        // Instead, load assignments the other way: all members, then check per-member.
        // For now load from the cred's perspective via a dedicated endpoint isn't available,
        // so we load all members' cred lists in parallel.
        api.members.list(),
      ])
      const members = membersRes.members ?? []
      setAllMembers(members)

      // find which members have this credential assigned
      const checks = await Promise.all(
        members.map(m => api.members.listCredentials(m.id).then(r => ({
          member: m,
          has: (r.credentials ?? []).some(c => c.id === cred.id),
        })))
      )
      const assignedMembers = checks.filter(c => c.has).map(c => ({
        id: c.member.id,
        name: c.member.name,
        created_at: c.member.created_at,
      } as CredentialItem))
      setAssigned(assignedMembers)

      const assignedIds = new Set(assignedMembers.map(m => m.id))
      const available = members.filter(m => !assignedIds.has(m.id))
      if (available.length > 0) setSelectedMember(available[0].id)
    } catch (e) {
      setError(String(e))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [cred.id])

  async function handleAssign() {
    if (!selectedMember) return
    setAssigning(true)
    try {
      await api.members.assignCredential(selectedMember, { credential_id: cred.id })
      await load()
    } catch (e) {
      setError(String(e))
    } finally {
      setAssigning(false)
    }
  }

  async function handleUnassign(memberId: string) {
    try {
      await api.members.unassignCredential(memberId, cred.id)
      await load()
    } catch (e) {
      setError(String(e))
    }
  }

  const assignedIds = new Set(assigned.map(m => m.id))
  const available = allMembers.filter(m => !assignedIds.has(m.id))

  return (
    <Modal title={`ASSIGNMENTS — ${cred.name}`} onClose={onClose}>
      {loading ? (
        <p className="text-xs tracking-widest" style={{ color: 'var(--c-muted)' }}>{'> LOADING...'}</p>
      ) : (
        <div className="space-y-5">
          {error && <p className="text-xs" style={{ color: 'var(--c-magenta)' }}>ERR: {error}</p>}

          {/* Assigned members */}
          <div>
            <div className="hud-label mb-2">ASSIGNED MEMBERS</div>
            {assigned.length === 0 ? (
              <p className="text-xs tracking-widest" style={{ color: 'var(--c-muted)' }}>
                // NO MEMBERS ASSIGNED
              </p>
            ) : (
              <div style={{ border: '1px solid var(--c-border)' }}>
                {assigned.map((m, i) => (
                  <div
                    key={m.id}
                    className="flex items-center justify-between px-3 py-2"
                    style={{
                      borderBottom: i < assigned.length - 1 ? '1px solid rgba(0,255,225,0.07)' : 'none',
                    }}
                  >
                    <span className="text-xs" style={{ color: 'var(--c-text)' }}>
                      <span style={{ color: 'var(--c-cyan)', marginRight: '0.4rem' }}>◆</span>
                      {m.name}
                    </span>
                    <button
                      onClick={() => handleUnassign(m.id)}
                      className="hud-btn-danger"
                      style={{ padding: '0.15rem 0.5rem', fontSize: '0.6rem' }}
                    >
                      [REMOVE]
                    </button>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Assign new member */}
          {available.length > 0 && (
            <div>
              <div className="hud-label mb-2">ASSIGN MEMBER</div>
              <div className="flex gap-2">
                <select
                  value={selectedMember}
                  onChange={e => setSelectedMember(e.target.value)}
                  className="hud-input flex-1"
                  style={{ padding: '0.35rem 0.75rem' }}
                >
                  {available.map(m => (
                    <option key={m.id} value={m.id} style={{ background: 'var(--c-surface)' }}>
                      {m.name}
                    </option>
                  ))}
                </select>
                <button
                  onClick={handleAssign}
                  disabled={assigning || !selectedMember}
                  className="hud-btn shrink-0"
                >
                  {assigning ? '...' : '[ASSIGN]'}
                </button>
              </div>
            </div>
          )}

          {available.length === 0 && assigned.length > 0 && (
            <p className="text-xs tracking-widest" style={{ color: 'var(--c-muted)' }}>
              // ALL MEMBERS ASSIGNED
            </p>
          )}
        </div>
      )}
    </Modal>
  )
}
