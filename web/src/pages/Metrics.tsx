import { useEffect, useState } from 'react'
import {
  Bar,
  BarChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'
import { api } from '../api/client'
import type { MemberResponse, MetricsSummaryResponse } from '../api/types'

export function Metrics() {
  const [members, setMembers] = useState<MemberResponse[]>([])
  const [selectedId, setSelectedId] = useState('')
  const [metrics, setMetrics] = useState<MetricsSummaryResponse | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    api.members.list().then(r => {
      setMembers(r.members ?? [])
      if (r.members?.length) setSelectedId(r.members[0].id)
    })
  }, [])

  useEffect(() => {
    if (!selectedId) return
    setLoading(true)
    setError('')
    api.members
      .metrics(selectedId)
      .then(setMetrics)
      .catch(e => setError(String(e)))
      .finally(() => setLoading(false))
  }, [selectedId])

  const errorRate = metrics && metrics.total_spans > 0
    ? ((metrics.error_spans / metrics.total_spans) * 100).toFixed(1)
    : '0.0'

  const countData = (metrics?.top_tools ?? []).map(t => ({
    name: shortName(t.name),
    count: t.count,
  }))

  const latencyData = (metrics?.top_tools ?? []).map(t => ({
    name: shortName(t.name),
    avg_ms: Math.round(t.avg_ms),
  }))

  return (
    <div className="p-8">
      <div className="flex items-center gap-4 mb-6">
        <h1 className="text-lg font-semibold text-slate-900">Metrics</h1>
        <select
          value={selectedId}
          onChange={e => setSelectedId(e.target.value)}
          className="border border-slate-300 rounded-lg px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
        >
          {members.map(m => (
            <option key={m.id} value={m.id}>{m.name}</option>
          ))}
        </select>
        <button
          onClick={() => setSelectedId(id => { const s = id; setSelectedId(''); setTimeout(() => setSelectedId(s), 0); return id })}
          className="text-sm text-indigo-600 hover:text-indigo-800"
        >
          Refresh
        </button>
      </div>

      {loading && <p className="text-sm text-slate-500">Loading…</p>}
      {error && <p className="text-sm text-red-600">{error}</p>}

      {metrics && !loading && (
        <div className="space-y-6">
          {/* Stats */}
          <div className="grid grid-cols-3 gap-4">
            <StatCard label="Total spans" value={metrics.total_spans.toLocaleString()} />
            <StatCard label="Error spans" value={metrics.error_spans.toLocaleString()} color="red" />
            <StatCard label="Error rate" value={`${errorRate}%`} color={parseFloat(errorRate) > 5 ? 'red' : 'green'} />
          </div>

          {metrics.top_tools.length === 0 && (
            <p className="text-sm text-slate-400">No spans recorded yet.</p>
          )}

          {metrics.top_tools.length > 0 && (
            <div className="grid grid-cols-2 gap-6">
              <ChartCard title="Calls per tool">
                <ResponsiveContainer width="100%" height={260}>
                  <BarChart data={countData} layout="vertical" margin={{ left: 16, right: 16 }}>
                    <CartesianGrid strokeDasharray="3 3" horizontal={false} />
                    <XAxis type="number" tick={{ fontSize: 11 }} />
                    <YAxis type="category" dataKey="name" tick={{ fontSize: 11 }} width={100} />
                    <Tooltip />
                    <Bar dataKey="count" fill="#6366f1" radius={[0, 4, 4, 0]} />
                  </BarChart>
                </ResponsiveContainer>
              </ChartCard>

              <ChartCard title="Avg latency per tool (ms)">
                <ResponsiveContainer width="100%" height={260}>
                  <BarChart data={latencyData} layout="vertical" margin={{ left: 16, right: 16 }}>
                    <CartesianGrid strokeDasharray="3 3" horizontal={false} />
                    <XAxis type="number" tick={{ fontSize: 11 }} />
                    <YAxis type="category" dataKey="name" tick={{ fontSize: 11 }} width={100} />
                    <Tooltip formatter={(v: number) => [`${v} ms`, 'avg latency']} />
                    <Bar dataKey="avg_ms" fill="#f59e0b" radius={[0, 4, 4, 0]} />
                  </BarChart>
                </ResponsiveContainer>
              </ChartCard>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

function StatCard({ label, value, color = 'default' }: { label: string; value: string; color?: 'default' | 'red' | 'green' }) {
  const valueColor = color === 'red' ? 'text-red-600' : color === 'green' ? 'text-emerald-600' : 'text-slate-900'
  return (
    <div className="bg-white rounded-xl border border-slate-200 px-5 py-4">
      <p className="text-xs font-medium text-slate-500 uppercase tracking-wide mb-1">{label}</p>
      <p className={`text-2xl font-semibold ${valueColor}`}>{value}</p>
    </div>
  )
}

function ChartCard({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="bg-white rounded-xl border border-slate-200 p-5">
      <h3 className="text-sm font-medium text-slate-700 mb-4">{title}</h3>
      {children}
    </div>
  )
}

function shortName(name: string) {
  return name.includes('.') ? name.split('.').pop()! : name
}
