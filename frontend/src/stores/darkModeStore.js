import { defineStore } from "pinia";
import { ref, computed, watch } from "vue";

export const useDarkModeStore = defineStore("darkMode", () => {
  // State
  const isDarkMode = ref(false);

  // Getters (Computed)
  const currentTheme = computed(() => {
    return isDarkMode.value ? "dark" : "light";
  });

  const themeClass = computed(() => {
    return isDarkMode.value ? "dark-theme" : "light-theme";
  });

  // Actions
  const loadFromLocalStorage = () => {
    try {
      const savedTheme = localStorage.getItem("darkMode");
      if (savedTheme !== null) {
        isDarkMode.value = JSON.parse(savedTheme);
        return true;
      }
    } catch (error) {
      console.error("Error loading dark mode from localStorage:", error);
      localStorage.removeItem("darkMode");
    }
    return false;
  };

  const saveToLocalStorage = () => {
    try {
      localStorage.setItem("darkMode", JSON.stringify(isDarkMode.value));
    } catch (error) {
      console.error("Error saving dark mode to localStorage:", error);
    }
  };

  const toggleDarkMode = () => {
    isDarkMode.value = !isDarkMode.value;
    applyTheme();
    saveToLocalStorage();
  };

  /**
   * 설정된 다크 모드 값을 직접 설정합니다.
   * @param {boolean} value - 다크 모드 여부 (true: 다크 모드, false: 라이트 모드)
   * @description 이 함수는 다크 모드 상태를 직접 설정하며, 시스템 설정을 무시합니다.
   * @returns
   */
  const setDarkMode = (value) => {
    if (typeof value !== "boolean") {
      console.warn("setDarkMode expects a boolean value");
      return;
    }

    isDarkMode.value = value;
    applyTheme();
    saveToLocalStorage();
  };

  const applyTheme = () => {
    const htmlElement = document.documentElement;

    if (isDarkMode.value) {
      htmlElement.classList.add("dark");
      htmlElement.classList.remove("light");
    } else {
      htmlElement.classList.add("light");
      htmlElement.classList.remove("dark");
    }
  };

  const initializeDarkMode = () => {
    try {
      // localStorage에서 설정 로드
      const loadedFromStorage = loadFromLocalStorage();

      // localStorage에 설정이 없으면 시스템 설정 확인
      if (!loadedFromStorage) {
        const prefersDark =
          window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches;
        isDarkMode.value = prefersDark;
        saveToLocalStorage();
      }

      // 테마 적용
      applyTheme();

      // 시스템 다크 모드 변경 감지 (localStorage 설정이 없을 때만)
      if (window.matchMedia) {
        const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
        mediaQuery.addEventListener("change", (e) => {
          // localStorage에 설정이 없을 때만 시스템 설정 따라가기
          const savedTheme = localStorage.getItem("darkMode");
          if (savedTheme === null) {
            isDarkMode.value = e.matches;
            applyTheme();
          }
        });
      }
    } catch (error) {
      console.error("Error initializing dark mode:", error);
      isDarkMode.value = false;
      applyTheme();
    }
  };

  const clearSettings = () => {
    isDarkMode.value = false;
    localStorage.removeItem("darkMode");
    applyTheme();
  };

  // Watch for changes and apply theme
  watch(isDarkMode, () => {
    applyTheme();
  });

  return {
    // State
    isDarkMode,

    // Getters
    currentTheme,
    themeClass,

    // Actions
    loadFromLocalStorage,
    saveToLocalStorage,
    toggleDarkMode,
    setDarkMode,
    applyTheme,
    initializeDarkMode,
    clearSettings,
  };
});
