<template>
  <div class="tagBox">
    <div class="header">
      <h1 class="text-2xl font-bold">Tag List</h1>
    </div>
    <ul class="list-disc gap-2 ml-2.5">
      <li v-for="tag in tagStore.filteredTags" :key="tag" v-show="tag">
        <button
          @click="handleTagClick(tag)"
          :class="{ active: tagStore.selectedTags.includes(tag) }"
        >
          {{ tag }}
        </button>
      </li>
    </ul>
  </div>
</template>

<script setup>
import { onMounted, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useTagStore } from "@/stores/tagStore";

const tagStore = useTagStore();
const router = useRouter();
const route = useRoute();

function handleTagClick(tag) {
  tagStore.toggleTag(tag);

  // 선택된 태그들을 쉼표로 구분하여 URL에 저장
  if (tagStore.selectedTags.length > 0) {
    const searchQuery = tagStore.selectedTags.map((t) => "#" + t).join(",");
    router.push({ path: "/", query: { search: searchQuery } });
  } else {
    router.push({ path: "/", query: {} });
  }
}

onMounted(async () => {
  try {
    await tagStore.initializeTags();

    // URL에서 초기 태그 상태 설정
    const searchQuery = route.query.search;
    if (searchQuery) {
      // 쉼표로 구분된 여러 태그 파싱
      const tagsFromUrl = searchQuery
        .split(",")
        .filter((t) => t.trim().startsWith("#"))
        .map((t) => t.trim().slice(1));
      tagStore.selectedTags.push(...tagsFromUrl);
    }
  } catch (error) {
    console.error("Failed to initialize tags:", error);
  }
});

watch(
  route,
  (newRoute) => {
    // URL 쿼리에서 선택된 태그 동기화
    const searchQuery = newRoute.query.search;
    if (searchQuery) {
      const tagsFromUrl = searchQuery
        .split(",")
        .filter((t) => t.trim().startsWith("#"))
        .map((t) => t.trim().slice(1));
      const urlTags = new Set(tagsFromUrl);
      const storeTags = new Set(tagStore.selectedTags);

      // 다른 경우에만 업데이트 (무한 루프 방지)
      if (urlTags.size !== storeTags.size || ![...urlTags].every((t) => storeTags.has(t))) {
        tagStore.selectedTags = tagsFromUrl;
      }
    } else {
      // 검색 쿼리가 없는 경우 태그 초기화
      if (tagStore.selectedTags.length > 0) {
        tagStore.selectedTags = [];
      }
    }
  },
  { immediate: true },
);
</script>

<style scoped>
.tagBox {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 1rem;
  padding: 1rem;
}

.header {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 0.35rem;
}

.header h1 {
  color: var(--text-primary);
  margin: 0;
  font-size: 1.1rem;
}

button {
  background: none;
  border: 1px solid transparent;
  border-radius: 999px;
  color: var(--text-secondary);
  cursor: pointer;
  font-size: 0.95rem;
  font-weight: 600;
  transition: all 0.3s ease;
  padding: 0.2rem 0.7rem;
  position: relative;
}

button:hover {
  color: var(--accent-color);
  background-color: var(--accent-soft);
  border-color: var(--border-color);
}

button::before {
  content: "# ";
}

button.active {
  color: var(--accent-color);
  background-color: var(--accent-soft);
  border-color: var(--accent-color);
}

/* 리스트 스타일 */
ul {
  list-style: none;
  padding: 0;
  margin: 0;
}

li {
  margin-bottom: 0.35rem;
}
</style>
