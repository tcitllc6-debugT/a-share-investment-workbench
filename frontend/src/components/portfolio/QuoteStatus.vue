<script setup>
import { computed } from 'vue'
import { formatTime, readValue } from './portfolio-api.mjs'

const props = defineProps({
  quote: { type: [Object, String], default: null },
  compact: { type: Boolean, default: false },
})

const state = computed(() => {
  if (typeof props.quote === 'string') return props.quote
  const provider = String(readValue(props.quote, 'quoteProvider', 'QuoteProvider', 'provider', 'Provider', 'source', 'Source') || '').toLowerCase()
  if (provider === 'manual') return 'manual'
  if (readValue(props.quote, 'quoteStale', 'QuoteStale', 'isStale', 'IsStale')) return 'stale'
  if (provider) return 'fresh'
  return String(readValue(props.quote, 'status', 'Status', 'quoteStatus', 'QuoteStatus') || 'unknown').toLowerCase()
})

const presentation = computed(() => {
  if (['fresh', 'ok', 'success', 'realtime'].includes(state.value)) return { label: '实时', tone: 'fresh' }
  if (['stale', 'expired', 'cached'].includes(state.value)) return { label: '缓存', tone: 'stale' }
  if (['manual', 'fixed'].includes(state.value)) return { label: '手动', tone: 'manual' }
  if (['loading', 'refreshing', 'pending'].includes(state.value)) return { label: '更新中', tone: 'loading' }
  if (['error', 'failed', 'unavailable'].includes(state.value)) return { label: '失败', tone: 'error' }
  return { label: '待报价', tone: 'unknown' }
})

const updatedAt = computed(() => readValue(props.quote, 'quoteAsOf', 'QuoteAsOf', 'updatedAt', 'UpdatedAt', 'quotedAt', 'QuotedAt', 'fetchedAt', 'FetchedAt'))
const provider = computed(() => readValue(props.quote, 'quoteProvider', 'QuoteProvider', 'provider', 'Provider', 'source', 'Source'))
const title = computed(() => [provider.value, updatedAt.value ? formatTime(updatedAt.value) : ''].filter(Boolean).join(' · ') || presentation.value.label)
</script>

<template>
  <span class="quote-status" :class="[`quote-status--${presentation.tone}`, { 'quote-status--compact': compact }]" :title="title">
    <i aria-hidden="true" />
    <span>{{ presentation.label }}</span>
    <small v-if="!compact && updatedAt">{{ formatTime(updatedAt) }}</small>
  </span>
</template>

<style scoped>
.quote-status { display:inline-flex;align-items:center;gap:6px;min-height:24px;padding:3px 7px;border:1px solid var(--workspace-border);border-radius:999px;color:var(--workspace-text-secondary);background:var(--workspace-surface-muted);font-size:10px;white-space:nowrap; }
.quote-status i { width:6px;height:6px;border-radius:50%;background:var(--workspace-text-faint);box-shadow:0 0 0 3px color-mix(in srgb,var(--workspace-text-faint) 14%,transparent); }
.quote-status small { color:var(--workspace-text-muted);font-size:9px; }
.quote-status--compact { min-height:20px;padding:2px 6px; }
.quote-status--fresh { color:var(--workspace-negative);background:color-mix(in srgb,var(--workspace-negative) 8%,var(--workspace-surface-solid));border-color:color-mix(in srgb,var(--workspace-negative) 25%,var(--workspace-border)); }
.quote-status--fresh i { background:var(--workspace-negative);box-shadow:0 0 0 3px color-mix(in srgb,var(--workspace-negative) 14%,transparent); }
.quote-status--stale,.quote-status--loading { color:var(--workspace-warning);background:color-mix(in srgb,var(--workspace-warning) 8%,var(--workspace-surface-solid));border-color:color-mix(in srgb,var(--workspace-warning) 25%,var(--workspace-border)); }
.quote-status--stale i,.quote-status--loading i { background:var(--workspace-warning);box-shadow:0 0 0 3px color-mix(in srgb,var(--workspace-warning) 14%,transparent); }
.quote-status--error { color:var(--workspace-positive);background:color-mix(in srgb,var(--workspace-positive) 8%,var(--workspace-surface-solid));border-color:color-mix(in srgb,var(--workspace-positive) 25%,var(--workspace-border)); }
.quote-status--error i { background:var(--workspace-positive);box-shadow:0 0 0 3px color-mix(in srgb,var(--workspace-positive) 14%,transparent); }
.quote-status--manual { color:var(--workspace-accent);background:var(--workspace-accent-soft);border-color:color-mix(in srgb,var(--workspace-accent) 24%,var(--workspace-border)); }
.quote-status--manual i { background:var(--workspace-accent);box-shadow:0 0 0 3px var(--workspace-accent-muted); }
@media (prefers-reduced-motion:no-preference) { .quote-status--loading i { animation:quote-pulse 1.3s ease-in-out infinite; } }
@keyframes quote-pulse { 50% { opacity:.35; } }
</style>
