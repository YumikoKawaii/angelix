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
}

export interface CredentialItem {
  id: string
  name: string
  created_at: string
}

export interface CreateCredentialRequest {
  name: string
  access_token: string
}

export interface AssignCredentialRequest {
  credential_id: string
}

export interface ToolStat {
  name: string
  count: number
  avg_ms: number
  error_rate: number
  input_tokens: number
  output_tokens: number
  cache_read_tokens: number
  cache_creation_tokens: number
}

export interface MetricsSummaryResponse {
  member_id: string
  total_spans: number
  error_spans: number
  input_tokens: number
  output_tokens: number
  cache_read_tokens: number
  cache_creation_tokens: number
  top_tools: ToolStat[]
}
