<template>
  <div class="navbar-fixed">
    <div class="left-group">
      <RouterLink class="brand" to="/">Redeyes dev</RouterLink>
      <form @submit.prevent="search">
        <label for="default-search" class="sr-only">Search</label>
        <input
          type="search"
          id="default-search"
          placeholder="찾고 싶은 글..."
          v-model="searchForm"
        />
        <button type="submit" class="search-btn">
          <font-awesome-icon :icon="['fas', 'magnifying-glass']" />
        </button>
      </form>
    </div>

    <div class="right-group">
      <RouterLink to="/" class="buttons">Blog</RouterLink>
      <RouterLink to="/about" class="buttons">About</RouterLink>
      <button class="theme-btn" @click="darkModeStore.toggleDarkMode" aria-label="theme toggle">
        <font-awesome-icon
          v-if="darkModeStore.isDarkMode"
          :icon="['fas', 'sun']"
          class="theme-icon"
        />
        <font-awesome-icon v-else :icon="['fas', 'moon']" class="theme-icon" />
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref } from "vue";
import { RouterLink, useRouter } from "vue-router";
import { useDarkModeStore } from "@/stores/darkModeStore";

const router = useRouter();
const darkModeStore = useDarkModeStore();

const searchForm = ref("");

function search() {
  router.push({ path: "/", query: { search: searchForm.value } });
}
</script>

<style scoped>
.navbar-fixed {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 20;

  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  min-height: var(--navbar-height);
  padding: 0.75rem 1.5rem;

  background-color: var(--bg-primary);
  backdrop-filter: blur(8px);
  border-bottom: 1px solid var(--border-color);
  box-shadow: 0 8px 24px var(--shadow);

  transition: all 0.3s ease;
}

.left-group,
.right-group {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.left-group {
  min-width: 0;
}

.brand {
  font-size: 1.5rem;
  font-weight: 700;
  letter-spacing: 0.01em;

  transition: all 0.3s ease;
}

.brand:hover {
  color: #dc143c;
}

.buttons {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-secondary);
  transition: all 0.2s ease;
}

.buttons:hover {
  color: var(--accent-color);
}

form {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 999px;
  padding: 0.4rem 0.4rem 0.4rem 0.8rem;
}

input {
  width: min(32vw, 260px);
  border: none;
  background: transparent;
  color: var(--text-primary);
  font-size: 0.95rem;
  outline: none;
}

.search-btn {
  width: 2rem;
  height: 2rem;
  aspect-ratio: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-secondary);
  border-radius: 50%;
  border: none;
  background: transparent;
  cursor: pointer;
  transition: all 0.3s ease;
}

.search-btn:hover {
  background-color: var(--accent-soft);
  color: var(--accent-color);
}

.theme-btn {
  width: 2.25rem;
  height: 2.25rem;
  display: grid;
  place-items: center;
  border-radius: 50%;
  border: 1px solid var(--border-color);
  background-color: var(--bg-secondary);
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.3s ease;
}

.theme-btn:hover {
  color: var(--accent-color);
  transform: translateY(-1px);
}

.theme-icon {
  font-size: 1rem;
}

.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

@media (max-width: 768px) {
  .navbar-fixed {
    padding: 0.75rem 1rem;
  }

  .brand {
    font-size: 1.25rem;
  }

  .buttons {
    display: none;
  }

  input {
    width: 36vw;
  }
}
</style>
