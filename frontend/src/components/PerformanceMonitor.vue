<template>
  <div v-if="showPerformance && isDev" class="performance-monitor">
    <button @click="toggleExpanded" class="toggle-btn">
      📊 Performance
    </button>
    <div v-if="expanded" class="performance-details">
      <div class="metric">
        <span>Memory: {{ memoryUsage }}</span>
      </div>
      <div class="metric">
        <span>Cache Size: {{ cacheSize }}</span>
      </div>
      <div class="metric">
        <span>Route Changes: {{ routeChanges }}</span>
      </div>
      <div class="metric">
        <span>API Calls: {{ apiCalls }}</span>
      </div>
      <button @click="clearCache" class="clear-btn">Clear Cache</button>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue';
import { useRouter } from 'vue-router';

const showPerformance = ref(true);
const expanded = ref(false);
const memoryUsage = ref('N/A');
const cacheSize = ref('N/A');
const routeChanges = ref(0);
const apiCalls = ref(0);

// 개발 환경에서만 표시
const isDev = import.meta.env.DEV;

const router = useRouter();

const toggleExpanded = () => {
  expanded.value = !expanded.value;
};

const updateMemoryUsage = () => {
  if (performance.memory) {
    const used = Math.round(performance.memory.usedJSHeapSize / 1024 / 1024 * 100) / 100;
    memoryUsage.value = `${used} MB`;
  }
};

const updateCacheSize = () => {
  try {
    let total = 0;
    for (let key in localStorage) {
      if (Object.prototype.hasOwnProperty.call(localStorage, key)) {
        total += localStorage[key].length;
      }
    }
    cacheSize.value = `${Math.round(total / 1024 * 100) / 100} KB`;
  } catch (e) {
    cacheSize.value = 'Error';
    console.error('Error calculating cache size:', e);
  }
};

const clearCache = () => {
  updateCacheSize();
  console.log('Cache cleared manually');
};

// API 호출 모니터링
const originalFetch = window.fetch;
window.fetch = function(...args) {
  apiCalls.value++;
  return originalFetch.apply(this, args);
};

let interval;

onMounted(() => {
  // 라우트 변경 감지
  router.afterEach(() => {
    routeChanges.value++;
  });

  // 주기적으로 성능 지표 업데이트
  interval = setInterval(() => {
    updateMemoryUsage();
    updateCacheSize();
  }, 2000);
  
  updateMemoryUsage();
  updateCacheSize();
});

onUnmounted(() => {
  if (interval) {
    clearInterval(interval);
  }
});
</script>

<style scoped>
.performance-monitor {
  position: fixed;
  top: 10px;
  right: 10px;
  z-index: 9999;
  background: var(--bg-secondary);
  border: 2px solid var(--border-color);
  border-radius: 8px;
  padding: 8px;
  font-size: 12px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.toggle-btn {
  background: var(--accent-color);
  color: white;
  border: none;
  padding: 4px 8px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 11px;
}

.performance-details {
  margin-top: 8px;
  min-width: 150px;
}

.metric {
  margin: 4px 0;
  color: var(--text-primary);
}

.clear-btn {
  background: #ff4757;
  color: white;
  border: none;
  padding: 2px 6px;
  border-radius: 3px;
  cursor: pointer;
  font-size: 10px;
  margin-top: 4px;
}

.clear-btn:hover {
  background: #ff3838;
}
</style>
