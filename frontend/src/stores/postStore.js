import { defineStore } from "pinia";
import { computed, ref, watch } from "vue";
import axios from "axios";

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

  // --- State ---
  /**
   * @type {import('vue').Ref<Array<Post>>}
   * @description posts array
   */
  const posts = ref([]);

  /** @type {import('vue').Ref<string|null>} */
  const currentPostId = ref(null);

  /** @type {import('vue').Ref<number|null>} */
  const lastFetched = ref(null);

  const cacheTime = 1000 * 60 * 60 * 24; // 24 hours

  // --- Local Storage Actions ---
  /**
   * load posts from localStorage
   * @returns {boolean} if loaded from localStorage, false otherwise
   */
  function loadPostsFromLocalStorage() {
    try {
      const cachedData = localStorage.getItem("postStore");

      if (!cachedData) return false;

      const { posts: cachedPosts, lastFetched: cachedTime } = JSON.parse(cachedData) || {};

      if (cachedPosts && Array.isArray(cachedPosts)) {
        posts.value = cachedPosts;
        lastFetched.value = cachedTime || null;
        return true;
      }
      return false;
    } catch (error) {
      console.error("Error loading posts from localStorage:", error);
      localStorage.removeItem("postStore");
      return false;
    }
  }

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

  /**
   * fetch posts from server and update the store
   * @returns {Promise<Array<Post>|null>} if fails, returns null
   */
  const fetchPostsFromServer = async () => {
    try {
      const { data } = await axios.get("https://blog.redeyes.dev/api/posts/");
      const newPosts = data;

      setPosts(newPosts);
      return newPosts;
    } catch (error) {
      console.error("Error fetching posts from server:", error);
      return null;
    }
  };

  // --- Core Actions ---

  /**
   * set posts and update localStorage
   * @param {Array<Post>} newPosts
   */
  function setPosts(newPosts) {
    posts.value = newPosts || [];
    lastFetched.value = Date.now();
    savePostsToLocalStorage();
  }

  /**
   * check if posts are in cache, if not fetch from server, if cache is stale(old) fetch in background
   * @returns {Promise<Array<Post>>} posts array, either from cache or freshly fetched
   */
  const initializePosts = async () => {
    const loadedFromCache = loadPostsFromLocalStorage();

    if (!loadedFromCache) {
      const freshPosts = await fetchPostsFromServer();
      return freshPosts || [];
    }

    if (lastFetched.value && Date.now() - lastFetched.value > cacheTime) {
      fetchPostsFromServer().then((data) => {
        if (data) console.log("Background posts update finished.");
      });
    }

    return posts.value;
  };

  function clearCache() {
    posts.value = [];
    lastFetched.value = null;
    localStorage.removeItem("postStore");
    console.log("Posts cache cleared");
  }

  /**
   * @param {string} postId
   */
  function setCurrentPostId(postId) {
    currentPostId.value = postId;
  }

  // --- Computed Properties ---
  const currentPostIndex = computed(() =>
    posts.value.findIndex((post) => post.folder === currentPostId.value),
  );

  const previousPost = computed(() =>
    currentPostIndex.value > 0 ? posts.value[currentPostIndex.value - 1] : null,
  );

  const nextPost = computed(() =>
    currentPostIndex.value < posts.value.length - 1
      ? posts.value[currentPostIndex.value + 1]
      : null,
  );

  watch(
    posts,
    () => {
      if (posts.value.length > 0) {
        savePostsToLocalStorage();
      }
    },
    { deep: true },
  );

  return {
    // State
    posts,
    currentPostId,
    lastFetched,

    // Getters
    currentPostIndex,
    previousPost,
    nextPost,

    // Actions
    loadPostsFromLocalStorage,
    fetchPostsFromServer,
    initializePosts,
    setPosts,
    setCurrentPostId,
    clearCache,
  };
});
