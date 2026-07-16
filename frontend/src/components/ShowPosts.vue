<template>
  <div class="posts-layout">
    <div class="posts-main-wrapper">
      <div v-if="posts.length > 0" class="posts-content-wrapper">
        <RouterLink
          class="postContainer"
          :to="`/posts/${post.folder}`"
          :key="post.folder"
          v-for="post in displayedPosts"
        >
          <div class="postImageBlock">
            <img
              v-if="post.coverImg"
              :src="getImageUrl(post.folder, post.coverImg)"
              alt="cover img"
              class="w-full h-full object-cover"
              loading="lazy"
            />
            <div v-else class="placeholder-default">
              <font-awesome-icon
                :icon="['fas', 'image']"
                size="2xl"
                style="color: var(--bg-primary)"
              />
            </div>
          </div>
          <div class="post-meta-block">
            <span class="post-title">{{ post.title }}</span>
            <span class="post-summary">
              {{ post.summary.slice(0, 115) }}
            </span>
            <div class="post-tags">
              <div class="tagBlock" v-for="tag in post.tag" :key="tag" v-show="tag">
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
          <!-- 이 요소가 뷰포트에 들어오면 자동으로 더 로드 -->
        </div>

        <!-- 로딩 인디케이터 -->
        <div v-if="isLoading" class="loading-indicator">
          <font-awesome-icon :icon="['fas', 'spinner']" spin size="2xl" />
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
        <font-awesome-icon :icon="['fas', 'arrow-up']" />
      </button>
    </Transition>
  </div>
</template>

<script setup>
import { onMounted, ref, watch, computed, onUnmounted, nextTick } from "vue";
import { RouterLink, useRoute } from "vue-router";
import axios from "axios";
import { usePostStore } from "@/stores/postStore";

/**
 * @typedef {Object} Post
 * @property {string} title
 * @property {string} summary
 * @property {string} date
 * @property {string} folder
 * @property {string} coverImg
 * @property {string[]} tag
 * @property {string} content
 */

const route = useRoute();
const postStore = usePostStore();

// 상태 관리
const posts = ref([]);
const displayLimit = ref(10); // 처음에 보여줄 포스트 수
const isLoading = ref(false);
const showScrollToTop = ref(false);
const autoLoadTrigger = ref(null);
let observer = null;

const displayedPosts = computed(() => {
  return posts.value.slice(0, displayLimit.value);
});

const hasMorePosts = computed(() => {
  return displayLimit.value < posts.value.length;
});

/**
 * return image url
 * @param {string} postFolder
 * @param {string} imageName
 */
const getImageUrl = (postFolder, imageName) => {
  return `https://blog.redeyes.dev/api/posts/images/${postFolder}/${imageName}`;
};

/** get Posts */
async function getPosts() {
  if (route.query.search || !postStore.posts.length) {
    const { data } = await axios.get("https://blog.redeyes.dev/api/posts/", {
      params: {
        search: route.query.search,
      },
    });

    if (route.query.search) {
      posts.value = data;
      return;
    } else {
      postStore.setPosts(data);
    }
  }

  posts.value = postStore.posts;
}

// 더 많은 포스트 로드
function loadMorePosts() {
  if (isLoading.value || !hasMorePosts.value) return;

  isLoading.value = true;

  console.log("Loading more posts...");

  // 로딩 시뮬레이션 (실제로는 API 호출)
  setTimeout(() => {
    displayLimit.value = Math.min(displayLimit.value + 10, posts.value.length);
    isLoading.value = false;
  }, 500);
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
        if (entry.isIntersecting && hasMorePosts.value && !isLoading.value) {
          console.log("Auto-loading more posts...");
          loadMorePosts();
        }
      });
    },
    {
      rootMargin: "200px", // 200px 전에 미리 로딩 (더 부드러운 경험)
      threshold: 0.1,
    },
  );
}

// Observer 시작
function startObserving() {
  if (autoLoadTrigger.value && observer) {
    observer.observe(autoLoadTrigger.value);
  }
}

// Observer 정리
function stopObserving() {
  if (observer) {
    observer.disconnect();
  }
}

onMounted(() => {
  getPosts();
  window.addEventListener("scroll", handleScroll);
  setupIntersectionObserver();

  // DOM이 업데이트된 후 observer 시작
  nextTick(() => {
    startObserving();
  });
});

onUnmounted(() => {
  window.removeEventListener("scroll", handleScroll);
  stopObserving();
});

// displayedPosts가 변경될 때마다 observer 재설정
watch(displayedPosts, () => {
  nextTick(() => {
    stopObserving();
    setupIntersectionObserver();
    startObserving();
  });
});

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
  background-color: var(--accent-color);
  position: relative;
  border-radius: 0.85rem;
  overflow: hidden;
  margin: 0.75rem;
  width: 9.6rem;
  flex-shrink: 0;
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
  border-width: 0.1rem;
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
