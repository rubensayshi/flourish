<template>
  <div>
    <div class="mb-6 text-slate-400">
      <span class="font-bold text-slate-100">{{ data.fight_name }}</span>
      &mdash;
      <span class="font-bold text-slate-100">{{ data.player_name }}</span>
      &mdash;
      Total healing: <span class="text-emerald-400 font-bold">{{ fmt(data.total_healing) }}</span>
      ({{ data.duration_sec }}s)
    </div>

    <div class="flex gap-6">
      <!-- Class talents -->
      <div class="flex-1 min-w-0">
        <h3 class="text-sm font-semibold text-slate-400 uppercase tracking-wide mb-2">Class &amp; Spec Talents</h3>
        <table class="w-full text-sm">
          <thead>
            <tr class="text-left text-slate-400 border-b border-slate-700">
              <th class="py-2 pr-3">Talent</th>
              <th class="py-2 pr-3 text-right">Attributed</th>
              <th class="py-2 pr-3 text-right">%</th>
              <th class="py-2 text-right">HPS</th>
            </tr>
          </thead>
          <tbody>
            <template v-for="(t, i) in data.talents" :key="t.name">
              <tr
                :class="i % 2 === 0 ? 'bg-slate-800/50' : ''"
                class="border-b border-slate-800"
              >
                <td class="py-1.5 pr-3">
                  <span class="inline-flex items-center gap-1.5">
                    <a v-if="wowheadUrl(t.name)" :href="wowheadUrl(t.name)" target="_blank" class="text-slate-100 hover:text-emerald-400 no-underline">{{ t.name }}</a>
                    <span v-else>{{ t.name }}</span>
                    <button
                      v-if="hasExplanation(t.name)"
                      @click="toggleExplanation(t.name)"
                      class="text-slate-500 hover:text-emerald-400 transition-colors flex-shrink-0"
                      :class="{ 'text-emerald-400': expandedTalents.has(t.name) }"
                      title="How is this calculated?"
                    >
                      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-3.5 h-3.5">
                        <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a.75.75 0 000 1.5h.253a.25.25 0 01.244.304l-.459 2.066A1.75 1.75 0 0010.747 15H11a.75.75 0 000-1.5h-.253a.25.25 0 01-.244-.304l.459-2.066A1.75 1.75 0 009.253 9H9z" clip-rule="evenodd" />
                      </svg>
                    </button>
                  </span>
                </td>
                <td class="py-1.5 pr-3 text-right font-mono">{{ fmt(t.attributed) }}</td>
                <td class="py-1.5 pr-3 text-right font-mono" :class="pctColor(t.pct)">
                  {{ t.pct.toFixed(1) }}%
                </td>
                <td class="py-1.5 text-right font-mono">{{ fmt(t.hps) }}</td>
              </tr>
              <tr v-if="expandedTalents.has(t.name)" :key="t.name + '-info'">
                <td colspan="4" class="px-3 py-3 bg-slate-800/80 border-b border-slate-700">
                  <div class="text-xs text-slate-300 leading-relaxed [&_code]:text-emerald-400 [&_code]:bg-slate-700/50 [&_code]:px-1 [&_code]:py-0.5 [&_code]:rounded" v-html="getExplanation(t.name)"></div>
                </td>
              </tr>
            </template>
          </tbody>
          <tfoot class="text-slate-500">
            <tr class="border-t border-slate-700">
              <td class="py-1.5 pr-3">
                <span class="inline-flex items-center gap-1.5">
                  Wasted (&gt;50% OH)
                  <button
                    @click="showWastedExplanation = !showWastedExplanation"
                    class="text-slate-500 hover:text-emerald-400 transition-colors flex-shrink-0"
                    :class="{ 'text-emerald-400': showWastedExplanation }"
                    title="What does this mean?"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-3.5 h-3.5">
                      <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a.75.75 0 000 1.5h.253a.25.25 0 01.244.304l-.459 2.066A1.75 1.75 0 0010.747 15H11a.75.75 0 000-1.5h-.253a.25.25 0 01-.244-.304l.459-2.066A1.75 1.75 0 009.253 9H9z" clip-rule="evenodd" />
                    </svg>
                  </button>
                </span>
              </td>
              <td class="py-1.5 pr-3 text-right font-mono">{{ fmt(data.wasted) }}</td>
              <td class="py-1.5 pr-3 text-right">&mdash;</td>
              <td class="py-1.5 text-right">&mdash;</td>
            </tr>
            <tr v-if="showWastedExplanation">
              <td colspan="4" class="px-3 py-3 bg-slate-800/80 border-b border-slate-700">
                <div class="text-xs text-slate-300 leading-relaxed">
                  Healing events where more than 50% was overheal are excluded from talent attribution, since the healing was largely wasted and doesn't meaningfully reflect talent value.
                </div>
              </td>
            </tr>
            <tr>
              <td class="py-1.5 pr-3">
                <span class="inline-flex items-center gap-1.5">
                  Unattributed
                  <button
                    @click="showUnattributedExplanation = !showUnattributedExplanation"
                    class="text-slate-500 hover:text-emerald-400 transition-colors flex-shrink-0"
                    :class="{ 'text-emerald-400': showUnattributedExplanation }"
                    title="What does this mean?"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-3.5 h-3.5">
                      <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a.75.75 0 000 1.5h.253a.25.25 0 01.244.304l-.459 2.066A1.75 1.75 0 0010.747 15H11a.75.75 0 000-1.5h-.253a.25.25 0 01-.244-.304l.459-2.066A1.75 1.75 0 009.253 9H9z" clip-rule="evenodd" />
                    </svg>
                  </button>
                </span>
              </td>
              <td class="py-1.5 pr-3 text-right font-mono">{{ fmt(data.unattributed) }}</td>
              <td class="py-1.5 pr-3 text-right">&mdash;</td>
              <td class="py-1.5 text-right">&mdash;</td>
            </tr>
            <tr v-if="showUnattributedExplanation">
              <td colspan="4" class="px-3 py-3 bg-slate-800/80 border-b border-slate-700">
                <div class="text-xs text-slate-300 leading-relaxed">
                  Healing that could not be attributed to any specific talent. This includes baseline healing from spells that aren't modified by any selected talent, or healing from interactions not yet tracked by the analyzer.
                </div>
              </td>
            </tr>
          </tfoot>
        </table>
      </div>

      <!-- Hero tree talents -->
      <div v-if="data.hero_trees?.length" class="flex-1 min-w-0">
        <div v-for="tree in data.hero_trees" :key="tree.name" class="mb-6 last:mb-0">
          <h3 class="text-sm font-semibold text-emerald-400 uppercase tracking-wide mb-2">{{ tree.name }}</h3>
          <table class="w-full text-sm">
            <thead>
              <tr class="text-left text-slate-400 border-b border-slate-700">
                <th class="py-2 pr-3">Talent</th>
                <th class="py-2 pr-3 text-right">Attributed</th>
                <th class="py-2 pr-3 text-right">%</th>
                <th class="py-2 text-right">HPS</th>
              </tr>
            </thead>
            <tbody>
              <template v-for="(t, i) in tree.talents" :key="t.name">
                <tr
                  :class="i % 2 === 0 ? 'bg-slate-800/50' : ''"
                  class="border-b border-slate-800"
                >
                  <td class="py-1.5 pr-3">
                    <span class="inline-flex items-center gap-1.5">
                      <a v-if="wowheadUrl(t.name)" :href="wowheadUrl(t.name)" target="_blank" class="text-slate-100 hover:text-emerald-400 no-underline">{{ t.name }}</a>
                      <span v-else>{{ t.name }}</span>
                      <button
                        v-if="hasExplanation(t.name)"
                        @click="toggleExplanation(t.name)"
                        class="text-slate-500 hover:text-emerald-400 transition-colors flex-shrink-0"
                        :class="{ 'text-emerald-400': expandedTalents.has(t.name) }"
                        title="How is this calculated?"
                      >
                        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-3.5 h-3.5">
                          <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a.75.75 0 000 1.5h.253a.25.25 0 01.244.304l-.459 2.066A1.75 1.75 0 0010.747 15H11a.75.75 0 000-1.5h-.253a.25.25 0 01-.244-.304l.459-2.066A1.75 1.75 0 009.253 9H9z" clip-rule="evenodd" />
                        </svg>
                      </button>
                    </span>
                  </td>
                  <td class="py-1.5 pr-3 text-right font-mono">{{ fmt(t.attributed) }}</td>
                  <td class="py-1.5 pr-3 text-right font-mono" :class="pctColor(t.pct)">
                    {{ t.pct.toFixed(1) }}%
                  </td>
                  <td class="py-1.5 text-right font-mono">{{ fmt(t.hps) }}</td>
                </tr>
                <tr v-if="expandedTalents.has(t.name)" :key="t.name + '-info'">
                  <td colspan="4" class="px-3 py-3 bg-slate-800/80 border-b border-slate-700">
                    <div class="text-xs text-slate-300 leading-relaxed [&_code]:text-emerald-400 [&_code]:bg-slate-700/50 [&_code]:px-1 [&_code]:py-0.5 [&_code]:rounded" v-html="getExplanation(t.name)"></div>
                  </td>
                </tr>
              </template>
            </tbody>
            <tfoot>
              <tr class="border-t border-slate-700 font-bold">
                <td class="py-1.5 pr-3">Total</td>
                <td class="py-1.5 pr-3 text-right font-mono">{{ fmt(tree.attributed) }}</td>
                <td class="py-1.5 pr-3 text-right font-mono" :class="pctColor(tree.pct)">
                  {{ tree.pct.toFixed(1) }}%
                </td>
                <td class="py-1.5 text-right font-mono">{{ fmt(tree.hps) }}</td>
              </tr>
            </tfoot>
          </table>
        </div>
      </div>
    </div>

    <p v-if="totalAttributed > data.total_healing" class="mt-4 text-xs text-slate-500">
      Talents can overlap (multiple talents buff the same heal).
      Total attributed may exceed total healing.
    </p>
  </div>
</template>

<script setup>
import { computed, ref, onMounted, onUpdated, nextTick } from 'vue'
import { wowheadUrl } from '../wowheadSpells.js'
import { hasExplanation, getExplanation } from '../talentExplanations.js'

const props = defineProps({ data: Object })

const expandedTalents = ref(new Set())
const showWastedExplanation = ref(false)
const showUnattributedExplanation = ref(false)

function toggleExplanation(name) {
  if (expandedTalents.value.has(name)) {
    expandedTalents.value.delete(name)
  } else {
    expandedTalents.value.add(name)
  }
}

function refreshTooltips() {
  nextTick(() => { window.$WowheadPower?.refreshLinks() })
}
onMounted(refreshTooltips)
onUpdated(refreshTooltips)

const totalAttributed = computed(() => {
  const talentSum = props.data.talents.reduce((sum, t) => sum + t.attributed, 0)
  const heroSum = (props.data.hero_trees || []).reduce((sum, tree) => sum + tree.attributed, 0)
  return talentSum + heroSum
})

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
