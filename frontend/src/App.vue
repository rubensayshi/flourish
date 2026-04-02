<template>
  <div class="min-h-screen bg-slate-900 text-slate-100">
    <header class="border-b border-slate-700 px-6 py-4 flex items-center justify-between">
      <router-link to="/" class="text-xl font-bold text-emerald-400 hover:text-emerald-300 flex items-center gap-2">
        <svg class="w-6 h-6" viewBox="0 0 24 24" fill="none">
          <path d="M12 2C6.5 6 3 11 3 16c0 3 2 5 4.5 5.5C8 19 9.5 16 12 14c2.5 2 4 5 4.5 7.5C19 21 21 19 21 16c0-5-3.5-10-9-14z" fill="currentColor" opacity="0.9"/>
          <path d="M12 8c-1.5 2.5-2.5 5-2.5 7.5 0 1.5.5 3 2.5 4 2-1 2.5-2.5 2.5-4 0-2.5-1-5-2.5-7.5z" fill="currentColor" opacity="0.4"/>
        </svg>
        Flourish
      </router-link>
      <router-link to="/settings" class="text-sm text-slate-400 hover:text-emerald-400">
        Settings
      </router-link>
    </header>
    <main class="px-4 py-8">
      <router-view />
    </main>
  </div>
</template>
