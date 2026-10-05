<template>
  <div class="posts-layout">
    <div class="posts-main-wrapper">
      <div v-if="posts.length > 0" class="posts-content-wrapper">
        <RouterLink
          class="postContainer"
          :to="`/posts/${post.folder}`"
          :key="post.folder"
          v-for="post in posts"
        >
          <div class="postImageBlock" :class="{ enablePaint: !post.coverImg }">
            <img
              v-if="post.coverImg"
              :src="getImageUrl(post.folder, post.coverImg)"
              alt="cover img"
              class="w-full h-full object-cover"
            />
            <div v-else class="placeholder-default">
              <Image :size="48" style="color: var(--bg-primary)" />
            </div>
          </div>
          <div class="post-meta-block">
            <span class="post-title">{{ post.title }}</span>
            <span class="post-summary">
              {{ post.summary.slice(0, 115) }}
            </span>
            <div class="post-tags">
              <div class="tagBlock" v-for="tag in post.tags" :key="tag" v-show="tag">
                {{ tag }}
              </div>
            </div>
            <div class="post-foot">
              <span>redeyes · {{ post.date }}</span>
            </div>
          </div>
        </RouterLink>

        <!-- 자동 로딩을 위한 트리거 요소 -->
        <div v-if="hasMorePosts && !isLoading" ref="autoLoadTrigger" class="auto-load-trigger">
          <!-- 이 요소가 뷰포트에 들어오면 자동으로 로드 -->
        </div>

        <!-- 로딩 인디케이터 -->
        <div v-if="isLoading" class="loading-indicator">
          <Loader2 :size="48" class="animate-spin" />
          <span>포스트를 불러오는 중...</span>
        </div>

        <!-- 모든 포스트 로드 완료 메시지 -->
        <div v-if="!hasMorePosts && posts.length > 0" class="end-message">
          <p>모든 포스트를 확인했습니다!</p>
        </div>
      </div>

      <!-- 포스트가 없을 때 메시지 -->
      <div v-if="posts.length == 0" class="empty-state posts-content-wrapper">
        <p>흠.. 포스트가 없나 보네요 ¯\_(ツ)_/¯</p>
      </div>
    </div>

    <!-- 맨 위로 가기 버튼 -->
    <Transition name="scroll-btn" appear>
      <button v-show="showScrollToTop" class="scroll-to-top-btn" @click="scrollToTop">
        <ArrowUp :size="24" />
      </button>
    </Transition>
  </div>
</template>

<script setup>
import { onMounted, ref, watch, computed, onUnmounted } from "vue";
import { RouterLink, useRoute } from "vue-router";
import { ArrowUp, Image, Loader2 } from "@lucide/vue";
import { usePostStore } from "@/stores/postStore";

/**
 * @typedef {Object} Post
 * @property {string} title
 * @property {string} summary
 * @property {string} date
 * @property {string} folder
 * @property {string} coverImg
 * @property {string[]} tags
 * @property {string} content
 */

const route = useRoute();
const postStore = usePostStore();

/** 페이지당 로드할 포스트 수 (서버 limit과 동일) */
const PAGE_SIZE = 10;

// 상태 관리
const isLoading = ref(false);
const showScrollToTop = ref(false);
const autoLoadTrigger = ref(null);
let observer = null;

// 검색 모드 전용 상태 (검색 결과는 store에 캐시하지 않음)
const searchPosts = ref([]);
const searchTotal = ref(0);

const isSearch = computed(() => !!route.query.search);

// 일반 모드: store의 로드된 포스트 / 검색 모드: 로컬 결과
const posts = computed(() => (isSearch.value ? searchPosts.value : postStore.posts));
const total = computed(() => (isSearch.value ? searchTotal.value : postStore.total));

const hasMorePosts = computed(() => posts.value.length < total.value);

/**
 * return image url
 * @param {string} postFolder
 * @param {string} imageName
 */
const getImageUrl = (postFolder, imageName) => {
  return `https://blog.redeyes.dev/api/posts/images/${postFolder}_01_${imageName}`;
};

/** get Posts: 첫 페이지 로드 */
async function getPosts() {
  if (isSearch.value) {
    const page = await postStore.fetchPostsPage({
      offset: 0,
      limit: PAGE_SIZE,
      search: route.query.search,
    });

    if (page) {
      searchPosts.value = page.posts;
      searchTotal.value = page.total;
    } else {
      searchPosts.value = [];
      searchTotal.value = 0;
    }
    return;
  }

  await postStore.initializePosts();
}

/** 중복 없이 목록에 새 페이지를 append한다 (folder 기준) */
function appendUnique(list, newPosts) {
  const seen = new Set(list.map((p) => p.folder));
  const merged = [...list];
  for (const p of newPosts) {
    if (!seen.has(p.folder)) {
      seen.add(p.folder);
      merged.push(p);
    }
  }
  return merged;
}

// 더 많은 포스트 로드 (서버에서 다음 페이지 fetch)
async function loadMorePosts() {
  if (isLoading.value || !hasMorePosts.value) return;

  isLoading.value = true;

  const searchAtRequest = route.query.search || null;

  try {
    const page = await postStore.fetchPostsPage({
      offset: posts.value.length,
      limit: PAGE_SIZE,
      search: route.query.search,
    });

    if (!page) return;

    // 요청 중 검색어가 바뀌었으면 오래된 응답이므로 버린다
    if (searchAtRequest !== (route.query.search || null)) return;

    if (isSearch.value) {
      searchPosts.value = appendUnique(searchPosts.value, page.posts);
      searchTotal.value = page.total;
    } else {
      postStore.appendPosts(page.posts, page.total);
    }
  } finally {
    isLoading.value = false;
  }
}

// 맨 위로 스크롤
function scrollToTop() {
  window.scrollTo({ top: 0, behavior: "smooth" });
}

// 스크롤 이벤트 핸들러
function handleScroll() {
  showScrollToTop.value = window.scrollY > 300;
}

// Intersection Observer 설정
function setupIntersectionObserver() {
  observer = new IntersectionObserver(
    (entries) => {
      entries.forEach((entry) => {
        if (entry.isIntersecting) {
          loadMorePosts();
        }
      });
    },
    {
      rootMargin: "200px",
      threshold: 0.1,
    },
  );
}

// Observer 정리
function stopObserving() {
  if (observer) {
    observer.disconnect();
  }
}

// 트리거 요소가 v-if로 생기면 observe, 사라지면 정리
watch(autoLoadTrigger, (el) => {
  stopObserving();
  if (el) {
    setupIntersectionObserver();
    observer.observe(el);
  }
});

onMounted(() => {
  window.addEventListener("scroll", handleScroll);
});

onUnmounted(() => {
  window.removeEventListener("scroll", handleScroll);
  stopObserving();
});

// 검색어/라우트 변경 시 첫 페이지 다시 로드 (immediate로 마운트 시에도 실행)
watch(route, getPosts, { immediate: true });
</script>

<style scoped>
p {
  font-family: "Noto Sans KR", sans-serif;
  font-optical-sizing: auto;
  font-weight: 500;
  font-style: normal;
  font-size: 1.45rem;
  line-height: 1.55;
}

.posts-layout {
  gap: 1rem;
  width: 100%;
  max-width: 1200px;
  margin: 0 auto;
}

.posts-main-wrapper {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-top: 1.5rem;
  min-width: 0;
  max-width: 47rem;
}

.posts-content-wrapper {
  width: 100%;
  max-width: 47rem;
}

.postContainer {
  display: flex;
  align-items: stretch;
  width: 100%;
  max-width: 47rem;
  min-height: 11.5rem;
  background-color: var(--bg-secondary);
  border-radius: 1rem;
  border: 1px solid var(--border-color);
  margin-bottom: 1.5rem;
  transition: all 0.3s ease;
  position: relative;
}

.postContainer:hover {
  border-color: var(--accent-color);
  transform: translateY(-2px);
  box-shadow: 0 10px 28px var(--shadow);
}

.postImageBlock {
  position: relative;
  border-radius: 0.85rem;
  overflow: hidden;
  margin: 0.75rem;
  width: 9.6rem;
  flex-shrink: 0;
}

.enablePaint {
  background-color: var(--accent-color);
}

.post-meta-block {
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 0.4rem;
  padding: 0.85rem 0.85rem 0.85rem 0;
}

.post-title {
  font-size: 1.25rem;
  font-weight: 700;
  line-height: 1.35;
}

.post-summary {
  color: var(--text-secondary);
  line-height: 1.5;
}

.post-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
}

.post-foot {
  color: var(--text-secondary);
  font-size: 0.92rem;
}

.placeholder-default {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
}

.tagBlock {
  border-width: 0.01rem;
  border-color: var(--text-secondary);
  color: var(--text-secondary);
  font-size: 0.84rem;
  padding: 0.1rem 0.45rem;
  border-radius: 999px;
}

.tagBlock::before {
  content: "# ";
}

.loading-indicator {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
  margin: 2rem 0;
  color: var(--text-secondary);
}

.loading-indicator span {
  font-size: 1rem;
  font-weight: 500;
}

.end-message {
  text-align: center;
  margin: 3rem 0;
  padding: 1.25rem;
  background-color: var(--bg-secondary);
  border-radius: 1rem;
  border: 1px solid var(--border-color);
}

.end-message p {
  color: var(--text-secondary);
  font-size: 1.125rem;
  font-weight: 500;
  margin: 0;
}

.auto-load-trigger {
  height: 50px;
  width: 100%;
  margin: 2rem 0;
}

.scroll-to-top-btn {
  position: fixed;
  bottom: 2rem;
  right: 2rem;
  width: 3rem;
  height: 3rem;
  background-color: var(--accent-color);
  color: var(--bg-secondary);
  border: none;
  border-radius: 50%;
  cursor: pointer;
  box-shadow: 0 4px 15px var(--shadow);
  transition: all 0.3s ease-in-out;
  z-index: 100;
}

.scroll-to-top-btn:hover {
  background-color: var(--text-primary);
  color: var(--bg-primary);
  transform: translateY(-2px);
  box-shadow: 0 6px 20px var(--shadow);
}

.empty-state {
  display: flex;
  justify-content: center;
  margin-top: 4rem;
}

.scroll-btn-enter-active,
.scroll-btn-leave-active {
  transition: all 0.4s cubic-bezier(0.25, 0.46, 0.45, 0.94);
}

.scroll-btn-enter-from {
  opacity: 0;
  transform: translateY(30px) scale(0.8);
}

.scroll-btn-enter-to {
  opacity: 1;
  transform: translateY(0) scale(1);
}

.scroll-btn-leave-from {
  opacity: 1;
  transform: translateY(0) scale(1);
}

.scroll-btn-leave-to {
  opacity: 0;
  transform: translateY(30px) scale(0.8);
}

@media (max-width: 1180px) {
  .posts-main-wrapper,
  .posts-content-wrapper {
    max-width: 100%;
  }
}

@media (max-width: 768px) {
  .posts-main-wrapper,
  .posts-content-wrapper {
    max-width: 100%;
  }

  .postContainer {
    flex-direction: column;
    min-height: auto;
  }

  .postImageBlock {
    width: calc(100% - 1.5rem);
    height: 10.5rem;
  }

  .post-meta-block {
    padding: 0 0.85rem 0.95rem;
  }
}
</style>
