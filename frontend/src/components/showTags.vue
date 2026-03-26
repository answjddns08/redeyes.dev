<template>
  <div class="tagBox">
    <div class="header">
        <h1 class="text-2xl font-bold">Tag List</h1>
      </div>
      <ul class="list-disc gap-2 ml-2.5">
        <li v-for="tag in tagStore.filteredTags" :key="tag" v-show="tag">
          <button @click="handleTagClick(tag)" :class="{ active: tagStore.selectedTag === tag }">
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
const route = useRoute()

function handleTagClick(tag) {
  tagStore.toggleTag(tag);

  // toggle 후의 상태로 라우팅 결정
  if (tagStore.selectedTag) {
    router.push({ path: "/", query: { search: "#" + tagStore.selectedTag } });
  } else {
    router.push({ path: "/", query: {} });
  }
}

onMounted(async () => {
  try {
    await tagStore.initializeTags();

    // URL에서 초기 태그 상태 설정
    const searchQuery = route.query.search;
    if (searchQuery && searchQuery.startsWith('#')) {
      const tagFromUrl = searchQuery.slice(1);
      tagStore.selectedTag = tagFromUrl;
    }
  } catch (error) {
    console.error("Failed to initialize tags:", error);
  }
});

watch(route, (newRoute) => {
  
  // URL 쿼리에서 선택된 태그 동기화
  const searchQuery = newRoute.query.search;
  if (searchQuery && searchQuery.startsWith('#')) {
    const tagFromUrl = searchQuery.slice(1); // '#' 제거
    if (tagStore.selectedTag !== tagFromUrl) {
      tagStore.selectedTag = tagFromUrl;
    }
  } else {
    // 검색 쿼리가 없거나 태그가 아닌 경우 선택 해제
    if (tagStore.selectedTag) {
      tagStore.selectedTag = null;
    }
  }
}, { immediate: true })
</script>

<style scoped>
.tagBox {
  display: flex;
  flex-direction: column;
}

.header {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 0.75rem;
}

.header h1 {
  color: var(--text-primary);
  margin: 0;
}

button {
  background: none;

  border: none;
  border-radius: 4px;

  color: var(--text-primary);

  cursor: pointer;

  font-size: larger;
  font-weight: 550;

  transition: all 0.3s ease;

  padding: 0rem 0.5rem;

  position: relative;
}

button:hover {
  color: var(--accent-color);
  background-color: var(--bg-primary);

  text-decoration: underline;

  transform: translateY(-2px);
}

button::before {
  content: "# ";
}

button.active {
  color: var(--accent-color);
  background-color: var(--bg-primary);
}

/* 리스트 스타일 */
ul {
  list-style: none;
  padding: 0;
  margin: 0;
}

li {
  margin-bottom: 0.25rem;
}
</style>
