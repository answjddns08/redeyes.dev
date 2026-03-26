import { defineStore } from "pinia";
import { computed, ref, watch } from "vue";

export const usePostStore = defineStore("postStore", () => {
  /**
   * post type
   * @typedef {Object} Post
   * @property {string} folder - post folder name
   * @property {string} title - post title
   * @property {string} date - post date
   * @property {Array<string>} tag - post tags
   * @property {string} content - post content
   * @property {string} coverImg - post cover image
   */

  /**
   * @type {Array<Post>}
   * @description posts array
   */
  const posts = ref([]);
  const currentPostId = ref(null);
  const lastFetched = ref(null);
  const cacheExpiry = 24 * 60 * 60 * 1000; // 24시간 캐시 유효시간

  /** 캐시가 유효한지 확인 */
  const isCacheValid = computed(() => {
    return lastFetched.value && Date.now() - lastFetched.value < cacheExpiry;
  });

  /** LocalStorage에서 posts 불러오기 */
  function loadPostsFromLocalStorage() {
    try {
      const cachedData = localStorage.getItem("postStore");
      if (cachedData) {
        const { posts: savedPosts, lastFetched: savedLastFetched } = JSON.parse(cachedData);
        posts.value = savedPosts || [];
        lastFetched.value = savedLastFetched;

        return true;
      }
    } catch (error) {
      console.error("Error loading posts from localStorage:", error);
      localStorage.removeItem("postStore");
    }
    return false;
  }

  /** LocalStorage에 posts 저장 */
  function savePostsToLocalStorage() {
    try {
      const dataToSave = {
        posts: posts.value,
        lastFetched: lastFetched.value,
      };
      localStorage.setItem("postStore", JSON.stringify(dataToSave));
    } catch (error) {
      console.error("Error saving posts to localStorage:", error);
    }
  }

  /** posts를 업데이트하는 함수 */
  function setPosts(newPosts) {
    posts.value = newPosts;
    lastFetched.value = Date.now();
    savePostsToLocalStorage();
  }

  /** 캐시가 유효한지 확인하는 함수 */
  function checkCacheValidity() {
    return isCacheValid.value && posts.value.length > 0;
  }

  /** 캐시 클리어 함수 */
  function clearCache() {
    posts.value = [];
    lastFetched.value = null;
    localStorage.removeItem("postStore");
    console.log("Posts cache cleared");
  }

  /** 현재 포스트 ID 설정 */
  function setCurrentPostId(postId) {
    currentPostId.value = postId;
  }

  /** 현재 포스트의 인덱스 계산 */
  const currentPostIndex = computed(() =>
    posts.value.findIndex((post) => post.folder === currentPostId.value)
  );

  /** 이전 포스트 계산 */
  const previousPost = computed(() =>
    currentPostIndex.value > 0 ? posts.value[currentPostIndex.value - 1] : null
  );

  /** 다음 포스트 계산 */
  const nextPost = computed(() =>
    currentPostIndex.value < posts.value.length - 1 ? posts.value[currentPostIndex.value + 1] : null
  );

  // LocalStorage와 동기화 (lastFetched는 자동 저장하지 않음)
  watch(
    posts,
    () => {
      if (posts.value.length > 0) {
        savePostsToLocalStorage();
      }
    },
    { deep: true }
  );

  // 초기화
  loadPostsFromLocalStorage();

  return {
    posts,
    currentPostId,
    lastFetched,
    isCacheValid,
    setPosts,
    setCurrentPostId,
    loadPostsFromLocalStorage,
    checkCacheValidity,
    clearCache,
    previousPost,
    nextPost,
  };
});
