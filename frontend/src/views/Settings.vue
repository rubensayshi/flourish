<template>
  <div class="max-w-lg mx-auto">
    <h2 class="text-xl font-bold mb-6">Settings</h2>

    <div v-for="(meta, key) in SETTINGS_META" :key="key" class="mb-6">
      <label class="block font-semibold mb-1">
        {{ meta.label }}
        <span class="ml-2 text-emerald-400 font-mono">{{ settings[key] }}</span>
      </label>
      <p class="text-sm text-slate-400 mb-2">{{ meta.description }}</p>
      <div class="flex items-center gap-3">
        <span class="text-xs text-slate-500">{{ meta.min }}</span>
        <input
          type="range"
          :min="meta.min"
          :max="meta.max"
          :step="meta.step"
          v-model.number="settings[key]"
          class="w-full accent-emerald-500"
        />
        <span class="text-xs text-slate-500">{{ meta.max }}</span>
      </div>
    </div>

    <p class="text-xs text-slate-500 mt-8">
      Settings are saved in your browser and applied to future analyses.
    </p>
  </div>
</template>

<script setup>
import { settings, SETTINGS_META } from '../composables/useSettings'
</script>
