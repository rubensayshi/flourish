<template>
  <div class="min-h-screen bg-slate-900 text-slate-100">
    <header class="border-b border-slate-700 px-6 py-4">
      <router-link to="/" class="text-xl font-bold text-emerald-400 hover:text-emerald-300">
        Resto Druid Talent Analyzer
      </router-link>
    </header>
    <main class="mx-auto max-w-5xl px-4 py-8">
      <router-view />
    </main>
  </div>
</template>
