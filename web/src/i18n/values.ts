import type { Locale } from './types'

// The endonyms are what each language calls itself.
//
// A switcher that translates the names of the languages is a switcher that
// a reader who cannot read the current language cannot use, so each option
// is named in its own language.
export const endonyms: Record<Locale, string> = {
  'zh-CN': '中文',
  en: 'EN',
}

// valueLabels maps a value the server publishes to the label the interface
// shows for it.
//
// The server speaks in closed vocabularies — a run status, a coverage
// state, a termination reason — and those vocabularies are part of the
// platform's contract, so they stay in their canonical form on the wire and
// are only translated where they are shown.
//
// A value that is not listed here is shown as the server published it. An
// unknown value is a fact about the server, and hiding it behind a
// placeholder would lose the very thing the reader needs to see.
export const valueLabels: Record<Locale, Record<string, string>> = {
  en: {
    running: 'running',
    completed: 'completed',
    failed: 'failed',
    passed: 'passed',
    covered: 'covered',
    'no coverage': 'no coverage',
    ok: 'ok',
    missing: 'missing',
    macro: 'macro',
    micro: 'micro',
    ratio: 'ratio',
    domain: 'domain',
    capability: 'capability',
    difficulty: 'difficulty',
    form: 'form',
    declared_infeasible: 'declared_infeasible',
    step_limit: 'step_limit',
    turn_limit: 'turn_limit',
    budget_exhausted: 'budget_exhausted',
    protocol_violation: 'protocol_violation',
    agent_error: 'agent_error',
    environment_error: 'environment_error',
    platform_error: 'platform_error',
    cancelled: 'cancelled',
    model_failure: 'model_failure',
    platform_failure: 'platform_failure',
  },
  'zh-CN': {
    running: '运行中',
    completed: '已完成',
    failed: '失败',
    passed: '通过',
    covered: '已覆盖',
    'no coverage': '无覆盖',
    ok: '正常',
    missing: '缺失',
    macro: '宏平均',
    micro: '微平均',
    ratio: '比率',
    domain: '领域',
    capability: '能力维度',
    difficulty: '难度',
    form: '形式',
    declared_infeasible: '声明不可行',
    step_limit: '步数上限',
    turn_limit: '轮次上限',
    budget_exhausted: '预算耗尽',
    protocol_violation: '协议违规',
    agent_error: '智能体错误',
    environment_error: '环境错误',
    platform_error: '平台错误',
    cancelled: '已取消',
    model_failure: '模型失败',
    platform_failure: '平台故障',
  },
}
