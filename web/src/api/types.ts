// The shapes the API returns.
//
// They mirror the Go structures in internal/domain, internal/report and
// internal/metrics. The field names match the JSON tags exactly, so a
// rename on the server is a compile error here rather than a silent blank.

export interface Interval {
  lower: number
  upper: number
  method: string
  requested: string
  basis: string
  level: number
}

export interface MetricRow {
  id: string
  name: string
  description: string
  builtin: boolean
  unit: string
  version: string
  formula: string
  numerator: string
  denominator: string
  aggregation: string
  value: number
  has_value: boolean
  coverage: string
  numerator_value: number
  denominator_value: number
  sample_size: number
  ci?: Interval
}

export interface Confidence {
  platform_failure_rate: number
  judgement_failure_rate: number
  valid_sample_size: number
  planned_sample_size: number
  coverage: number
  threshold: number
  insufficient_confidence: boolean
  reason?: string
}

export interface MainScore {
  aggregation: string
  total_score: number
  has_value: boolean
  weights: Record<string, number>
  coverage: number
  ci?: Interval
  pass_at_k: number
  pass_at_k_has_value: boolean
  pass_power_k: number
  pass_power_k_has_value: boolean
  continuous_mean: number
  primary_criterion: string
}

export interface BreakdownCell {
  key: string
  sample_size: number
  success_rate: number
  has_value: boolean
  coverage: string
  continuous_mean: number
  ci?: Interval
}

export interface Breakdown {
  dimension: string
  cells: BreakdownCell[]
}

export interface ScorerBreakdown {
  scorer: string
  pass: number
  fail: number
  unscored: number
  share: Record<string, number>
}

export interface FailedTrial {
  trial_id: string
  task_id: string
  domain: string
  difficulty: string
  termination_reason: string
  failure_class: string
  detail?: string
}

export interface FailureAnalysis {
  termination_reasons: Record<string, number>
  failure_classes: Record<string, number>
  scorers: ScorerBreakdown[]
  failed_trials: FailedTrial[]
}

export interface TaskResult {
  task_id: string
  display_name: string
  form: string
  domain: string
  difficulty: string
  trials: number
  passed: number
  success_rate: number
  has_value: boolean
  mean_score: number
}

export interface PlanInfo {
  plan_id: string
  task_set_id: string
  task_set_version: string
  agent_id: string
  repetitions: number
  aggregation: string
  ci_method: string
  fingerprint: string
}

// RunPlan is the plan as the run list publishes it.
//
// The report's plan carries the fingerprint of the scoring configuration and
// the run list's plan does not, so the two shapes are declared separately: a
// screen that read a fingerprint off a run list would get nothing and fail
// on it.
export interface RunPlan {
  plan_id: string
  task_set_id: string
  task_set_version: string
  agent_id: string
  repetitions: number
  aggregation: string
  ci_method: string
}

export interface Report {
  run_id: string
  generated_at: string
  plan: PlanInfo
  confidence: Confidence
  main_score: MainScore
  metrics: MetricRow[]
  breakdown: Breakdown[]
  failures: FailureAnalysis
  tasks: TaskResult[]
}

export interface Task {
  task_id: string
  form: string
  display_name: string
  benchmark: string
  task_set_version: string
  capability: {
    level1_dimensions: string[] | null
    level2_labels: string[] | null
    difficulty: string
    domains: string[] | null
  }
  description: {
    instruction: string
    evaluator_note?: string
  }
  budget: {
    steps: number
    tokens: number
    duration_seconds: number
    cost_usd: number
    errors: number
  }
  interaction_contract?: {
    action_space?: string[] | null
    observation_shape?: string
    termination_conditions?: string[] | null
    multi_turn?: boolean
  }
  judgement: {
    combination: string
    primary: number
    reference_answer?: string
    scorers: {
      type: string
      params?: Record<string, unknown>
      weight: number
      primary: boolean
    }[]
  }
  meta: {
    author: string
    created_at: string
    updated_at: string
  }
}

export interface TaskSet {
  task_set_id: string
  version: string
  benchmark: string
  description?: string
  created_at: string
  frozen: boolean
  tasks: Task[]
}

export interface Agent {
  agent_id: string
  name: string
  version: string
  shape: string
  endpoint?: string
  mode?: string
  options?: Record<string, unknown>
  self_check?: {
    status: string
    tasks: number
    action_format_error_rate: number
    contract_violations: number
    usage_reporting_ok: boolean
    invalid_reasons?: string[]
    checked_at: string
  }
  created_at: string
}

export interface Run {
  run_id: string
  plan: RunPlan
  status: string
  started_at: string
  finished_at?: string
  trials: unknown[]
  trajectories: unknown[]
  verdicts: unknown[]
}

export interface MetricDef {
  id: string
  name: string
  description: string
  builtin: boolean
  unit: string
  aggregations: string[]
  default_aggregation: string
  numerator: string
  denominator: string
  formula: string
  ci_method: string
  version: string
}

export interface CustomKind {
  name: string
  description: string
  required_params: string[]
}

export interface MetricsResponse {
  metrics: MetricDef[]
  custom_kinds: CustomKind[]
}
