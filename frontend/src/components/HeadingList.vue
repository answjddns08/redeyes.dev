<template>
  <div class="relative">
    <div v-if="headings.headings.length" class="block">
      <button 
        v-for="(heading, index) in headings.headings" 
        :class="[`h${heading.depth}`, { active: activeHeadingIndex === index }]" 
        :key="index"
        @click="scrollToHeading(heading.tokens[0].text)"
      >
        {{ heading.tokens[0].text }}
      </button>
    </div>
  </div>
</template>

<script setup>
import { watch, ref, onMounted, onUnmounted, nextTick } from "vue";

const props = defineProps({
  headings: {
    Type: Array(Object),
    default: {},
  },
});

const activeHeadingIndex = ref(0);
let observer = null;

// 헤딩으로 스크롤하는 함수
const scrollToHeading = (headingText) => {
  // 헤딩 텍스트를 기반으로 해당 요소를 찾아서 스크롤
  const headingElements = document.querySelectorAll('h1, h2, h3, h4, h5, h6');
  const targetHeading = Array.from(headingElements).find(el => 
    el.textContent.trim() === headingText.trim()
  );
  
  if (targetHeading) {
    targetHeading.scrollIntoView({ 
      behavior: 'smooth', 
      block: 'start',
      inline: 'nearest'
    });
  }
};

// Intersection Observer 설정
const setupIntersectionObserver = () => {
  // 기존 observer가 있다면 정리
  if (observer) {
    observer.disconnect();
  }
  
  const headingElements = document.querySelectorAll('h1, h2, h3, h4, h5, h6');
  
  if (headingElements.length === 0) return;

  observer = new IntersectionObserver(
    (entries) => {
      // 현재 화면에 보이는 헤딩들 중 가장 위에 있는 것을 찾기
      let visibleHeadings = entries
        .filter(entry => entry.isIntersecting)
        .sort((a, b) => a.boundingClientRect.top - b.boundingClientRect.top);

      if (visibleHeadings.length > 0) {
        const activeHeading = visibleHeadings[0].target;
        const activeText = activeHeading.textContent.trim();
        
        // headings 배열에서 해당 헤딩의 인덱스 찾기
        const index = props.headings.headings?.findIndex(heading => 
          heading.tokens[0].text.trim() === activeText
        );
        
        if (index !== -1) {
          activeHeadingIndex.value = index;
        }
      }
    },
    {
      rootMargin: '-10% 0px -80% 0px', // 상단 10%, 하단 80% 여백으로 더 정확한 감지
      threshold: 0.1
    }
  );

  headingElements.forEach(heading => {
    observer.observe(heading);
  });
};

onMounted(() => {
  // DOM이 완전히 로드된 후 observer 설정
  nextTick(() => {
    setTimeout(() => {
      setupIntersectionObserver();
    }, 500); // 콘텐츠가 렌더링될 시간을 충분히 줌
  });
});

onUnmounted(() => {
  if (observer) {
    observer.disconnect();
  }
});

watch(() => props.headings, () => {
  // headings가 변경되면 observer를 다시 설정
  nextTick(() => {
    setTimeout(() => {
      setupIntersectionObserver();
    }, 300);
  });
}, { deep: true });
</script>

<style scoped>
.block {
  position: fixed;

  top: 25%;
  transform: translateY(-50%);

  width: 15rem;

  right: 8.125rem;

  padding: 0.25rem 0.75rem;

  line-height: 1.5;

  border-left: 3px solid var(--border-color);

  max-height: 70vh;

  z-index: 10;
}

button {
  display: block;

  width: 100%;

  text-align: left;

  font-weight: 600;

  font-size: 1rem;

  color: var(--text-color);

  background-color: transparent;

  border: none;

  cursor: pointer;

  transition: all 0.3s ease-in-out;
}

button:hover {
  color: var(--accent-color);
  transform: translateX(0.25rem);
}

button.active {
  color: var(--accent-color);
  background-color: var(--shadow);
  border-radius: 0.25rem;
  transform: translateX(0.5rem);
  font-weight: 700;
  border-left: 3px solid var(--accent-color);
  padding-left: 0.5rem;
}

.h2 {
  margin-left: 0.75rem;
}

.h3 {
  margin-left: 1.5rem;
}

.h4 {
  margin-left: 2.25rem;
}

.h5 {
  margin-left: 3rem;
}

.h6 {
  margin-left: 3.75rem;
}
</style>
