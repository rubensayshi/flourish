<template>
  <div class="flex items-center gap-3 text-emerald-400">
    <div class="leaf-spinner">
      <svg class="h-6 w-6" viewBox="0 0 24 24" fill="none">
        <path d="M12 2C6.5 6 3 11 3 16c0 3 2 5 4.5 5.5C8 19 9.5 16 12 14c2.5 2 4 5 4.5 7.5C19 21 21 19 21 16c0-5-3.5-10-9-14z" fill="currentColor" class="leaf-outer"/>
        <path d="M12 8c-1.5 2.5-2.5 5-2.5 7.5 0 1.5.5 3 2.5 4 2-1 2.5-2.5 2.5-4 0-2.5-1-5-2.5-7.5z" fill="currentColor" class="leaf-vein"/>
      </svg>
    </div>
    <span class="text-slate-400"><slot>Loading...</slot></span>
  </div>
</template>

<style scoped>
.leaf-spinner {
  animation: leaf-pulse 1.8s ease-in-out infinite;
}

.leaf-outer {
  opacity: 0.9;
}

.leaf-vein {
  opacity: 0.3;
  animation: vein-glow 1.8s ease-in-out infinite;
}

@keyframes leaf-pulse {
  0%, 100% {
    transform: scale(0.85) rotate(-8deg);
    filter: brightness(0.7);
  }
  50% {
    transform: scale(1.1) rotate(8deg);
    filter: brightness(1.3) drop-shadow(0 0 6px rgba(52, 211, 153, 0.5));
  }
}

@keyframes vein-glow {
  0%, 100% {
    opacity: 0.2;
  }
  50% {
    opacity: 0.6;
  }
}
</style>
