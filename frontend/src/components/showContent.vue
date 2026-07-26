<template>
  <div class="mb-20" v-html="renderedContent"></div>
</template>

<script setup>
import { computed, ref, watch } from "vue";

const props = defineProps({
  content: {
    type: String,
    required: true,
    default: "",
  },
  headings: {
    type: Array,
    required: false,
    default: () => [],
  },
});

const emit = defineEmits(["update:headings"]);

// 추출된 제목 정보를 저장할 반응형 변수 선언
const extractedHeadings = ref([]);

// HTML 문자열에서 제목을 추출하고 슬러그를 생성하는 함수
const extractHeadings = (htmlString) => {
  if (!htmlString) return [];

  const headingsArray = [];

  // DOMParser를 사용하여 문자열을 실제 DOM 객체로 변환
  const parser = new DOMParser();
  // 'text/html'로 파싱하여 전체 구조를 가져옵니다.
  const doc = parser.parseFromString(htmlString, "text/html");

  // 모든 heading 태그(h1~h6)를 선택합니다.
  const headingElements = doc.querySelectorAll("h1, h2, h3, h4, h5, h6");

  headingElements.forEach((el) => {
    const depth = parseInt(el.tagName.substring(1)); // 'h1' -> 1
    const headingText = el.textContent.trim();

    // ID를 위한 텍스트 정리 (기존 marked 로직 재현)
    const headingId = headingText
      .toLowerCase()
      .replace(/[^\w\s-]/g, "") // 특수문자 제거 (알파벳, 숫자, 공백, 하이픈만 남김)
      .replace(/\s+/g, "-") // 공백을 하이픈으로 치환
      .trim();

    headingsArray.push({
      id: headingId,
      depth: depth,
      text: headingText,
    });
  });

  return headingsArray;
};

// computed 속성: 이제 단순하게 content를 반환합니다 (v-html이 렌더링 담당)
const renderedContent = computed(() => {
  return props.content;
});

// Watcher를 사용하여 content가 변경될 때마다 제목을 추출합니다.
watch(
  () => props.content,
  (newContent) => {
    const extracted = extractHeadings(newContent);
    extractedHeadings.value = extracted;
    // 부모 컴포넌트에 제목 목록 업데이트 알림
    emit("update:headings", extracted);
  },
  { immediate: true },
);
</script>

<style scoped>
:deep(p) {
  margin-bottom: 1rem;
}

:deep(h1) {
  font-size: 2.25rem;
  font-weight: bold;
  margin-bottom: 1.5rem;
  margin-top: 2rem;
  border-bottom: 2px solid var(--border-color);
  padding-bottom: 0.5rem;
}

:deep(h2) {
  font-size: 1.875rem;
  font-weight: bold;
  margin-bottom: 1.25rem;
  margin-top: 1.75rem;
}

:deep(h3) {
  font-size: 1.5rem;
  font-weight: semibold;
  margin-bottom: 1rem;
  margin-top: 1.5rem;
}

:deep(h4) {
  font-size: 1.25rem;
  font-weight: semibold;
  margin-bottom: 0.75rem;
  margin-top: 1.25rem;
}

:deep(h5) {
  font-size: 1.125rem;
  font-weight: semibold;
  margin-bottom: 0.5rem;
  margin-top: 1rem;
}

:deep(h6) {
  font-size: 1rem;
  font-weight: semibold;
  margin-bottom: 0.5rem;
  margin-top: 1rem;
}

:deep(ul) {
  list-style-type: disc;
  margin-left: 1.5rem;
  margin-bottom: 1.5rem;
}

:deep(ol) {
  list-style-type: decimal;
  margin-left: 1.5rem;
  margin-bottom: 1.5rem;
}

:deep(li) {
  margin-bottom: 0.5rem;
}

:deep(blockquote) {
  border-left: 4px solid var(--accent-color);
  padding-left: 1rem;
  margin: 1.5rem 0;
  font-style: italic;
  background-color: var(--border-color);
  padding: 1rem;
}

:deep(blockquote p) {
  margin: 0;
}

:deep(blockquote code) {
  background-color: var(--accent-color);
  font-style: italic;
  color: var(--bg-primary);

  padding: 0.15rem 0.25rem 0.15rem 0.25rem;
  margin-right: 0.15rem;
}

:deep(code) {
  background-color: var(--border-color);
  padding: 0.25rem 0.35rem;
  margin-right: 0.15rem;
  border-radius: 0.25rem;
  font-size: 0.875rem;
}

:deep(pre) {
  background-color: var(--border-color);
  color: var(--text-primary);
  padding: 1rem;
  border-radius: 0.5rem;
  overflow-x: auto;
  margin: 1.5rem 0;
}

:deep(pre code) {
  background-color: transparent;
  padding: 0;
  color: inherit;
}

:deep(table) {
  width: 100%;
  border-collapse: collapse;
  margin: 1.5rem 0;
}

:deep(th),
:deep(td) {
  border: 1px solid #d1d5db;
  padding: 0.75rem;
  text-align: left;
}

:deep(th) {
  background-color: #f9fafb;
  font-weight: semibold;
}

:deep(img) {
  max-width: 100%;
  height: auto;
  border-radius: 0.25rem;
  margin: 1rem 0;
}

:deep(a) {
  color: #3b82f6;
  text-decoration: underline;
}

:deep(a:hover) {
  color: #1d4ed8;
}

:deep(hr) {
  border: none;
  border-top: 1px solid #e5e7eb;
  margin: 2rem 0;
}
</style>
