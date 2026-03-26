<template>
  <main class="flex flex-col items-center mt-10 gap-2">
    <p class="text-6xl mt-4 font-bold" style="color: var(--text-primary);">Construction Zone</p>

    <p class="mt-2" style="color: var(--text-secondary);">This page is currently under development.</p>
    <p class="my-2" style="color: var(--text-secondary);">You can go back to the home page.</p>

    <button @click="$router.push('/')">
      Go to Home
    </button>
  </main>
</template>

<script setup></script>

<style scoped>
.svg-spinner {
  width: 5rem;
  height: 5rem;

  margin-top: 1rem;

  animation: rotate 2s linear infinite;
}

.path {
  stroke: var(--text-primary);
  stroke-linecap: round;
  animation: dash 1.5s ease-in-out infinite;
}

@keyframes rotate {
  100% {
    transform: rotate(360deg);
  }
}

@keyframes dash {
  0% {
    stroke-dasharray: 1, 300;
    stroke-dashoffset: 0;
  }
  50% {
    stroke-dasharray: 200, 300;
    stroke-dashoffset: -35;
  }
  100% {
    stroke-dasharray: 1, 300;
    stroke-dashoffset: -240;
  }
}

button {
  cursor: pointer;

  font-size: 1rem;
  font-weight: bold;

  border: none;
  border-radius: 0.375rem;
  padding: 0.5rem 1rem;

  color: white;
  background-color: #3b82f6;

  transition: background-color 0.3s ease;
}

button:hover {
  background-color: #2563eb;
}
</style>