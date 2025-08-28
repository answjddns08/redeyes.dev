<template>
  <main>
    <ShowPosts />
  </main>
</template>

<script setup>
import { useHead } from "@vueuse/head";
import ShowPosts from "@/components/ShowPosts.vue";

// 홈페이지 SEO 메타태그 설정
useHead({
  title: 'Kellog Blog - 개발 블로그',
  meta: [
    {
      name: 'description',
      content: '개발자 redeyes의 개인 블로그입니다. 웹 개발, 프로그래밍, 기술 트렌드에 대한 글을 공유합니다.'
    },
    {
      name: 'keywords',
      content: '개발 블로그, 웹개발, 프로그래밍, Vue.js, JavaScript, Node.js, 기술 블로그'
    },
    // Open Graph
    {
      property: 'og:title',
      content: 'Kellog Blog - 개발 블로그'
    },
    {
      property: 'og:description',
      content: '개발자 redeyes의 개인 블로그입니다. 웹 개발, 프로그래밍, 기술 트렌드에 대한 글을 공유합니다.'
    },
    {
      property: 'og:type',
      content: 'website'
    },
    {
      property: 'og:url',
      content: 'https://blog.redeyes.dev'
    },
    // Twitter Card
    {
      name: 'twitter:card',
      content: 'summary'
    },
    {
      name: 'twitter:title',
      content: 'Kellog Blog - 개발 블로그'
    },
    {
      name: 'twitter:description',
      content: '개발자 redeyes의 개인 블로그입니다. 웹 개발, 프로그래밍, 기술 트렌드에 대한 글을 공유합니다.'
    }
  ]
});
</script>

<style scoped>
.sortMenu {
  position: absolute;

  top: 1.5rem;
  left: 45rem;

  padding: 0.5rem;

  color: #727272;
  font-weight: 700;
  font-size: 0.875rem;

  background-color: #b3b3b3;
  border-radius: 0.5rem;

  white-space: nowrap;
  /* 줄바꿈 끄기 */

  transition: ease-out 0.25s;
}

.sortMenu:hover {
  background-color: #858585;
}
</style>
