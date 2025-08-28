import { defineStore } from "pinia";
import { ref, computed } from "vue";
import axios from "axios";

export const useTagStore = defineStore("tags", () => {
  // State
  const tags = ref([]);
  const selectedTag = ref(null);
  const lastFetched = ref(null);
  const cacheExpiry = ref(24 * 60 * 60 * 1000); // 24시간 캐시 유효시간

  // Getters (Computed)
  const isCacheValid = computed(() => {
    return lastFetched.value && Date.now() - lastFetched.value < cacheExpiry.value;
  });

  const filteredTags = computed(() => {
    return tags.value.filter((tag) => tag && tag.trim() !== "");
  });

  // Actions
  const loadFromLocalStorage = () => {
    try {
      const cachedData = localStorage.getItem("tagStore");
      if (cachedData) {
        const { tags: cachedTags, lastFetched: cachedLastFetched } = JSON.parse(cachedData);
        tags.value = cachedTags || [];
        lastFetched.value = cachedLastFetched;

        //console.log("Tags loaded from localStorage:", tags.value.length, "tags");
        return true;
      }
    } catch (error) {
      console.error("Error loading tags from localStorage:", error);
      localStorage.removeItem("tagStore");
    }
    return false;
  };

  const saveToLocalStorage = () => {
    try {
      const dataToSave = {
        tags: tags.value,
        lastFetched: lastFetched.value,
      };
      localStorage.setItem("tagStore", JSON.stringify(dataToSave));
      //console.log("Tags saved to localStorage");
    } catch (error) {
      console.error("Error saving tags to localStorage:", error);
    }
  };

  const fetchTags = async (forceRefresh = false) => {
    // 캐시가 유효하고 강제 새로고침이 아닌 경우 서버 요청 건너뛰기
    if (!forceRefresh && isCacheValid.value && tags.value.length > 0) {
      //console.log("Using cached tags, skipping server request");
      return tags.value;
    }

    try {
      //console.log("Fetching tags from server...");
      const { data } = await axios.get("https://blog.redeyes.dev/api/tags/");

      tags.value = data || [];
      lastFetched.value = Date.now();

      // localStorage에 저장
      saveToLocalStorage();

      //console.log("Tags fetched and cached:", tags.value.length, "tags");
      return tags.value;
    } catch (error) {
      console.error("Error fetching tags:", error);
      // 오류 발생 시 캐시된 데이터 사용
      if (tags.value.length > 0) {
        //console.log("Using cached tags due to fetch error");
        return tags.value;
      }
      throw error;
    }
  };

  const initializeTags = async () => {
    selectedTag.value = null;

    // 먼저 localStorage에서 로드 시도
    const loadedFromCache = loadFromLocalStorage();

    // 캐시가 유효하면 서버 요청 건너뛰기
    if (loadedFromCache && isCacheValid.value) {
      //console.log("Using valid cached tags");
      return tags.value;
    }

    // 캐시가 없거나 만료된 경우 서버에서 가져오기
    return await fetchTags(true);
  };

  const toggleTag = (tag) => {
    selectedTag.value = selectedTag.value === tag ? null : tag;
  };

  const refreshTags = async () => {
    //console.log("Forcing tag refresh...");
    return await fetchTags(true);
  };

  const clearCache = () => {
    tags.value = [];
    selectedTag.value = null;
    lastFetched.value = null;
    localStorage.removeItem("tagStore");
    //console.log("Tag cache cleared");
  };

  return {
    // State
    tags,
    selectedTag,
    lastFetched,
    cacheExpiry,

    // Getters
    isCacheValid,
    filteredTags,

    // Actions
    loadFromLocalStorage,
    saveToLocalStorage,
    fetchTags,
    initializeTags,
    toggleTag,
    refreshTags,
    clearCache,
  };
});
