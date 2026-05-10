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

const MONO = "'Share Tech Mono', 'JetBrains Mono', monospace"

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

  function refresh() {
    const id = selectedId
    setSelectedId('')
    setTimeout(() => setSelectedId(id), 0)
  }

  const errorRate =
    metrics && metrics.total_spans > 0
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

  const tooltipStyle = {
    background: '#0d1117',
    border: '1px solid rgba(0,255,225,0.28)',
    color: '#c9d1d9',
    fontFamily: MONO,
    fontSize: 11,
    borderRadius: 0,
  }

  return (
    <div className="p-8">
      {/* Header */}
      <div className="flex items-end justify-between mb-6">
        <div>
          <div className="text-xs tracking-widest" style={{ color: 'var(--c-muted)' }}>// TELEMETRY</div>
          <h1 className="text-xl tracking-widest uppercase" style={{ color: 'var(--c-cyan)' }}>
            METRICS CONSOLE
          </h1>
        </div>
        <div className="flex items-center gap-3">
          <select
            value={selectedId}
            onChange={e => setSelectedId(e.target.value)}
            className="hud-input"
            style={{ width: 'auto', padding: '0.35rem 0.75rem' }}
          >
            {members.map(m => (
              <option key={m.id} value={m.id} style={{ background: 'var(--c-surface)' }}>
                {m.name}
              </option>
            ))}
          </select>
          <button onClick={refresh} className="hud-btn" style={{ padding: '0.35rem 0.75rem' }}>
            [REFRESH]
          </button>
        </div>
      </div>

      {loading && (
        <p className="text-xs tracking-widest" style={{ color: 'var(--c-muted)' }}>
          {'> LOADING TELEMETRY...'}
        </p>
      )}
      {error && (
        <p className="text-xs" style={{ color: 'var(--c-magenta)' }}>
          ERR: {error}
        </p>
      )}

      {metrics && !loading && (
        <div className="space-y-5">
          {/* Span stats */}
          <div className="grid grid-cols-3 gap-4">
            <StatCard label="TOTAL SPANS"  value={metrics.total_spans.toLocaleString()} />
            <StatCard label="ERROR SPANS"  value={metrics.error_spans.toLocaleString()} danger />
            <StatCard
              label="ERROR RATE"
              value={`${errorRate}%`}
              danger={parseFloat(errorRate) > 5}
            />
          </div>

          {/* Token stats */}
          <div className="grid grid-cols-4 gap-4">
            <TokenCard label="INPUT TOKENS"  value={metrics.input_tokens} />
            <TokenCard label="OUTPUT TOKENS" value={metrics.output_tokens} />
            <TokenCard label="CACHE READ"    value={metrics.cache_read_tokens} />
            <TokenCard label="CACHE WRITE"   value={metrics.cache_creation_tokens} />
          </div>

          {metrics.top_tools.length === 0 && (
            <p className="text-xs tracking-widest" style={{ color: 'var(--c-muted)' }}>
              // NO SPANS RECORDED
            </p>
          )}

          {metrics.top_tools.length > 0 && (
            <div className="grid grid-cols-2 gap-5">
              <ChartCard title="CALLS PER TOOL">
                <ResponsiveContainer width="100%" height={260}>
                  <BarChart data={countData} layout="vertical" margin={{ left: 8, right: 16, top: 4, bottom: 4 }}>
                    <CartesianGrid strokeDasharray="2 6" horizontal={false} stroke="rgba(0,255,225,0.08)" />
                    <XAxis
                      type="number"
                      tick={{ fontSize: 10, fill: '#3d4451', fontFamily: MONO }}
                      axisLine={{ stroke: 'rgba(0,255,225,0.15)' }}
                      tickLine={false}
                    />
                    <YAxis
                      type="category"
                      dataKey="name"
                      tick={{ fontSize: 10, fill: '#3d4451', fontFamily: MONO }}
                      axisLine={false}
                      tickLine={false}
                      width={110}
                    />
                    <Tooltip contentStyle={tooltipStyle} cursor={{ fill: 'rgba(0,255,225,0.04)' }} />
                    <Bar dataKey="count" fill="#00ffe1" radius={[0, 2, 2, 0]} />
                  </BarChart>
                </ResponsiveContainer>
              </ChartCard>

              <ChartCard title="AVG LATENCY / TOOL (ms)">
                <ResponsiveContainer width="100%" height={260}>
                  <BarChart data={latencyData} layout="vertical" margin={{ left: 8, right: 16, top: 4, bottom: 4 }}>
                    <CartesianGrid strokeDasharray="2 6" horizontal={false} stroke="rgba(0,255,225,0.08)" />
                    <XAxis
                      type="number"
                      tick={{ fontSize: 10, fill: '#3d4451', fontFamily: MONO }}
                      axisLine={{ stroke: 'rgba(0,255,225,0.15)' }}
                      tickLine={false}
                    />
                    <YAxis
                      type="category"
                      dataKey="name"
                      tick={{ fontSize: 10, fill: '#3d4451', fontFamily: MONO }}
                      axisLine={false}
                      tickLine={false}
                      width={110}
                    />
                    <Tooltip
                      contentStyle={tooltipStyle}
                      formatter={(v: number) => [`${v} ms`, 'avg latency']}
                      cursor={{ fill: 'rgba(0,255,225,0.04)' }}
                    />
                    <Bar dataKey="avg_ms" fill="#ff2d78" radius={[0, 2, 2, 0]} />
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

function StatCard({
  label,
  value,
  danger = false,
}: {
  label: string
  value: string
  danger?: boolean
}) {
  const accent = danger ? 'var(--c-magenta)' : 'var(--c-cyan)'
  return (
    <div
      style={{
        background: 'var(--c-surface)',
        border: '1px solid var(--c-border)',
        borderTop: `2px solid ${accent}`,
        boxShadow: danger
          ? '0 0 14px rgba(255,45,120,0.08)'
          : '0 0 14px rgba(0,255,225,0.06)',
        padding: '1rem 1.25rem',
      }}
    >
      <p className="text-xs tracking-widest mb-2" style={{ color: 'var(--c-muted)' }}>{label}</p>
      <p className="text-2xl tracking-wide" style={{ color: accent }}>{value}</p>
    </div>
  )
}

function TokenCard({ label, value }: { label: string; value: number }) {
  return (
    <div
      style={{
        background: 'var(--c-surface)',
        border: '1px solid var(--c-border)',
        borderTop: '2px solid rgba(0,255,225,0.35)',
        padding: '0.85rem 1rem',
      }}
    >
      <p className="text-xs tracking-widest mb-1" style={{ color: 'var(--c-muted)' }}>{label}</p>
      <p className="text-lg" style={{ color: 'var(--c-text)' }}>{value.toLocaleString()}</p>
    </div>
  )
}

function ChartCard({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div
      style={{
        background: 'var(--c-surface)',
        border: '1px solid var(--c-border)',
        boxShadow: '0 0 16px rgba(0,255,225,0.05)',
        padding: '1.25rem',
      }}
    >
      <h3 className="text-xs tracking-widest mb-4" style={{ color: 'var(--c-cyan)' }}>
        // {title}
      </h3>
      {children}
    </div>
  )
}

function shortName(name: string) {
  return name.includes('.') ? name.split('.').pop()! : name
}
