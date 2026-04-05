<template>
  <div v-if="visible" class="fixed inset-0 z-50 flex items-center justify-center p-4" @click.self="dismiss">
    <div class="absolute inset-0 bg-black/70 backdrop-blur-sm"></div>
    <div class="relative bg-slate-800 border border-slate-600 rounded-xl max-w-lg w-full p-6 shadow-2xl">
      <div class="text-center mb-4">
        <svg class="w-10 h-10 mx-auto mb-2 text-emerald-400" viewBox="0 0 24 24" fill="none">
          <path d="M12 2C6.5 6 3 11 3 16c0 3 2 5 4.5 5.5C8 19 9.5 16 12 14c2.5 2 4 5 4.5 7.5C19 21 21 19 21 16c0-5-3.5-10-9-14z" fill="currentColor" opacity="0.9"/>
          <path d="M12 8c-1.5 2.5-2.5 5-2.5 7.5 0 1.5.5 3 2.5 4 2-1 2.5-2.5 2.5-4 0-2.5-1-5-2.5-7.5z" fill="currentColor" opacity="0.4"/>
        </svg>
        <h2 class="text-xl font-bold text-emerald-400">Welcome to Flourish</h2>
      </div>

      <div class="space-y-3 text-sm text-slate-300 leading-relaxed">
        <p>
          Flourish analyzes your WarcraftLogs reports and breaks down how much healing
          each individual talent contributes. Paste a report link, pick your fight and
          player, and see exactly which talents are pulling their weight.
        </p>

        <div class="bg-amber-500/10 border border-amber-500/30 rounded-lg px-4 py-3">
          <p class="text-amber-400 font-medium mb-1">Experimental - grain of salt required</p>
          <p class="text-slate-400">
            This is very much a work in progress. Attribution logic can be tricky and some
            talents are harder to isolate than others. Numbers might be off - treat them as
            directional, not gospel.
          </p>
        </div>

        <div class="bg-slate-700/50 rounded-lg px-4 py-3">
          <p class="text-slate-300 mb-1">
            Real talk: this whole thing was vibe-coded with AI. We know the community has
            mixed feelings about AI-built tools, but building this properly by hand would've
            taken weeks (trust me, I've done similar projects before). AI let us ship something
            usable in a fraction of the time.
          </p>
        </div>

        <p>
          If you think a talent's attribution looks off compared to your theorycrafting -
          <span class="text-emerald-400 font-medium">please speak up!</span>
          That kind of feedback is exactly what makes this tool better. Hit me up as
          <span class="text-slate-200">saikodrakie</span> on Discord or in the
          <span class="text-slate-200">Dreamgrove</span> server.
        </p>
      </div>

      <button
        @click="dismiss"
        class="mt-5 w-full py-2.5 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white font-medium transition-colors cursor-pointer"
      >
        Got it, let me analyze some logs
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const STORAGE_KEY = 'flourish:welcome-dismissed'
const visible = ref(!localStorage.getItem(STORAGE_KEY))

function dismiss() {
  localStorage.setItem(STORAGE_KEY, '1')
  visible.value = false
}
</script>
