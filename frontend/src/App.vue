<script setup>
import { RouterView } from "vue-router";
import { onMounted } from "vue";
import NavBar from "./components/NavBar.vue";
import SiteFooter from "./components/Footer.vue";
import { useDarkModeStore } from "./stores/darkModeStore";

const darkModeStore = useDarkModeStore();

// 앱 시작 시 다크 모드 초기화
onMounted(() => {
  darkModeStore.initializeDarkMode();
});
</script>

<template>
  <div class="app-layout">
    <NavBar />
    <main class="main-content">
      <RouterView />
    </main>
    <SiteFooter />
  </div>
</template>

<style scoped>
.app-layout {
  min-height: 100vh;
  width: 100%;
  display: flex;
  flex-direction: column;
}

.main-content {
  flex: 1;
  padding-top: var(--navbar-height, 6rem);
  transition: padding-top 0.3s ease;
}
</style>
