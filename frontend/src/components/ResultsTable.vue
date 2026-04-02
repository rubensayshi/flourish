<template>
  <div>
    <div class="mb-4 text-slate-400">
      <span class="font-bold text-slate-100">{{ data.fight_name }}</span>
      &mdash;
      <span class="font-bold text-slate-100">{{ data.player_name }}</span>
      &mdash;
      Total healing: <span class="text-emerald-400 font-bold">{{ fmt(data.total_healing) }}</span>
      ({{ data.duration_sec }}s)
    </div>

    <table class="w-full text-sm">
      <thead>
        <tr class="text-left text-slate-400 border-b border-slate-700">
          <th class="py-2 pr-4">Talent</th>
          <th class="py-2 pr-4 text-right">Attributed</th>
          <th class="py-2 pr-4 text-right">% Total</th>
          <th class="py-2 text-right">HPS</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="(t, i) in data.talents"
          :key="t.name"
          :class="i % 2 === 0 ? 'bg-slate-800/50' : ''"
          class="border-b border-slate-800"
        >
          <td class="py-1.5 pr-4">{{ t.name }}</td>
          <td class="py-1.5 pr-4 text-right font-mono">{{ fmt(t.attributed) }}</td>
          <td class="py-1.5 pr-4 text-right font-mono" :class="pctColor(t.pct)">
            {{ t.pct.toFixed(1) }}%
          </td>
          <td class="py-1.5 text-right font-mono">{{ fmt(t.hps) }}</td>
        </tr>
      </tbody>
      <tfoot class="text-slate-500">
        <tr class="border-t border-slate-700">
          <td class="py-1.5 pr-4">Wasted (&gt;50% OH)</td>
          <td class="py-1.5 pr-4 text-right font-mono">{{ fmt(data.wasted) }}</td>
          <td class="py-1.5 pr-4 text-right">&mdash;</td>
          <td class="py-1.5 text-right">&mdash;</td>
        </tr>
        <tr>
          <td class="py-1.5 pr-4">Unattributed</td>
          <td class="py-1.5 pr-4 text-right font-mono">{{ fmt(data.unattributed) }}</td>
          <td class="py-1.5 pr-4 text-right">&mdash;</td>
          <td class="py-1.5 text-right">&mdash;</td>
        </tr>
      </tfoot>
    </table>

    <p v-if="totalAttributed > data.total_healing" class="mt-4 text-xs text-slate-500">
      Talents can overlap (multiple talents buff the same heal).
      Total attributed may exceed total healing.
    </p>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({ data: Object })

const totalAttributed = computed(() =>
  props.data.talents.reduce((sum, t) => sum + t.attributed, 0)
)

function fmt(n) {
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M'
  if (n >= 1_000) return (n / 1_000).toFixed(1) + 'k'
  return String(Math.round(n))
}

function pctColor(pct) {
  if (pct >= 5) return 'text-emerald-400'
  if (pct >= 2) return 'text-emerald-600'
  return 'text-slate-400'
}
</script>
