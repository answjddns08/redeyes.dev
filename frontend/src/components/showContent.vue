<template>
  <div class="mb-20" v-html="renderedContent"></div>
</template>

<script setup>
import { computed } from 'vue';
import { marked } from 'marked';

const props = defineProps({
  content: {
    type: String,
    required: true,
    default: ''
  },
  headings: {
    type: Array,
    required: false,
    default: () => []
  },
});

const emit = defineEmits(['update:headings']);

const renderedContent = computed(() => {
  // content가 없으면 빈 문자열 반환
  if (!props.content) {
    return '';
  }

  const headingsArray = [];

  const renderer = new marked.Renderer();

  renderer.heading = function ({ tokens, depth }) {
    headingsArray.push({ tokens, depth });

    const headingText = this.parser.parseInline(tokens);
    // ID를 위한 텍스트 정리 (공백을 하이픈으로, 특수문자 제거)
    const headingId = headingText
      .toLowerCase()
      .replace(/[^\w\s-]/g, '') // 특수문자 제거
      .replace(/\s+/g, '-') // 공백을 하이픈으로
      .trim();

    return `<h${depth} id="${headingId}">${headingText}</h${depth}>`;
  };

  const rendered = marked(props.content, {
    renderer,
    gfm: true,
    breaks: true,
  });

  // headings 업데이트 emit
  emit('update:headings', headingsArray);

  return rendered;
});
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
  border-radius: 0.5rem;
}

:deep(code) {
  background-color: var(--border-color);
  padding: 0.25rem 0.5rem;
  border-radius: 0.25rem;
  font-family: "Courier New", monospace;
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
  border-radius: 0.5rem;
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
