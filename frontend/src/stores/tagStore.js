import { defineStore } from "pinia";
import { ref, computed } from "vue";
import axios from "axios";

export const useTagStore = defineStore("tags", () => {
  // State
  /** @type {import('vue').Ref<string[]>} */
  const tags = ref([]);

  /** @type {import('vue').Ref<string[]>} */
  const selectedTags = ref([]);

  /** @type {import('vue').Ref<number|null>} */
  const lastFetched = ref(null);
  const cacheTime = 1000 * 60 * 60 * 24; // 24 hours

  // Getters
  const filteredTags = computed(() => {
    return tags.value.filter((tag) => tag && tag.trim() !== "");
  });

  // Actions
  /**
   * load tags from localStorage
   * @returns true if loaded from localStorage, false otherwise
   */
  const loadFromLocalStorage = () => {
    try {
      const cachedData = localStorage.getItem("tagStore");

      if (!cachedData) return false;

      const { tags: cachedTags, lastFetched: cachedTime } = JSON.parse(cachedData) || {};

      if (cachedTags && Array.isArray(cachedTags)) {
        tags.value = cachedTags;
        lastFetched.value = cachedTime || null;
        return true;
      }

      return false;
    } catch (error) {
      console.error("Error loading tags from localStorage:", error);
      localStorage.removeItem("tagStore");
      return false;
    }
  };

  const saveToLocalStorage = () => {
    try {
      lastFetched.value = Date.now();
      const dataToSave = {
        tags: tags.value,
        lastFetched: lastFetched.value,
      };
      localStorage.setItem("tagStore", JSON.stringify(dataToSave));
    } catch {
      console.error("Error saving tags to localStorage");
    }
  };

  /**
   * fetch tags from server and update localStorage
   * @returns {Promise<string[]>} tags from server or cached tags if server fails
   */
  const fetchTagsFromServer = async () => {
    try {
      /** @type {import('axios').AxiosResponse<string[]>} */
      const { data } = await axios.get("https://blog.redeyes.dev/api/tags/");
      tags.value = data || [];
      saveToLocalStorage();
      return tags.value;
    } catch (error) {
      console.error("Error fetching tags from server:", error);
      return tags.value;
    }
  };

  /**
   *  initialize tags from localStorage or server, depending on cache validity
   * @returns {Promise<string[]>} tags from cache or server
   */
  const initializeTags = async () => {
    selectedTags.value = [];

    const iscacheLoaded = loadFromLocalStorage();

    if (!iscacheLoaded) {
      const serverData = await fetchTagsFromServer();
      return serverData && serverData.length > 0 ? serverData : [];
    }

    if (lastFetched.value && Date.now() - lastFetched.value > cacheTime) {
      fetchTagsFromServer().then(() => {
        console.log("Background cache update finished.");
      });
    }

    return tags.value;
  };

  const toggleTag = (tag) => {
    const index = selectedTags.value.indexOf(tag);
    if (index > -1) {
      selectedTags.value.splice(index, 1);
    } else {
      selectedTags.value.push(tag);
    }
  };

  const clearCache = () => {
    tags.value = [];
    selectedTags.value = [];
    lastFetched.value = null;
    localStorage.removeItem("tagStore");
  };

  return {
    tags,
    selectedTags,
    filteredTags,
    loadFromLocalStorage,
    saveToLocalStorage,
    fetchTagsFromServer,
    initializeTags,
    toggleTag,
    clearCache,
  };
});
