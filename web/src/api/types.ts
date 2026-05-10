export interface MemberResponse {
  id: string
  name: string
  email: string
  token: string
  created_at: string
}

export interface CreateMemberRequest {
  name: string
  email: string
  api_key: string
}

export interface ToolStat {
  name: string
  count: number
  avg_ms: number
  error_rate: number
}

export interface MetricsSummaryResponse {
  member_id: string
  total_spans: number
  error_spans: number
  top_tools: ToolStat[]
}
