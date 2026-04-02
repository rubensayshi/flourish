<template>
  <div>
    <label class="block text-sm text-slate-400 mb-1">Fight</label>
    <select
      :value="modelValue"
      @change="$emit('update:modelValue', Number($event.target.value))"
      class="w-full rounded-lg bg-slate-800 border border-slate-600 px-4 py-2.5
             text-slate-100 focus:outline-none focus:border-emerald-500"
    >
      <option :value="0" disabled>Select a fight...</option>
      <option v-for="f in fights" :key="f.id" :value="f.id">
        {{ f.name }} &mdash; {{ f.kill ? 'Kill' : 'Wipe' }} ({{ f.duration }}s)
      </option>
    </select>
  </div>
</template>

<script setup>
defineProps({
  fights: Array,
  modelValue: Number,
})
defineEmits(['update:modelValue'])
</script>
