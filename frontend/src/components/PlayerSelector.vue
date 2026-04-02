<template>
  <div>
    <label class="block text-sm text-slate-400 mb-1">Player</label>
    <select
      :value="modelValue"
      @change="$emit('update:modelValue', $event.target.value)"
      class="w-full rounded-lg bg-slate-800 border border-slate-600 px-4 py-2.5
             text-slate-100 focus:outline-none focus:border-emerald-500"
    >
      <option value="" disabled>Select a druid...</option>
      <option v-for="d in druids" :key="d.id" :value="d.name">
        {{ d.name }} ({{ d.server }})
      </option>
    </select>
  </div>
</template>

<script setup>
defineProps({
  druids: Array,
  modelValue: String,
})
defineEmits(['update:modelValue'])
</script>
