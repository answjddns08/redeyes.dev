<template>
  <main class="home-layout">
    <!-- 왼쪽: 태그 목록 -->
    <div class="tags-wrapper">
      <showTags />
    </div>

    <!-- 가운데: 포스트 목록 -->
    <ShowPosts />

    <!-- 오른쪽: 작가 정보 -->
    <div class="author-wrapper">
      <AuthorField />
    </div>
  </main>
</template>

<script setup>
import showTags from "@/components/showTags.vue";
import { useHead } from "@vueuse/head";
import ShowPosts from "@/components/ShowPosts.vue";
import AuthorField from "@/components/authorField.vue";

// 홈페이지 SEO 메타태그 설정
useHead({
  title: "Kellog Blog - 개발 블로그",
  meta: [
    {
      name: "description",
      content:
        "개발자 redeyes의 개인 블로그입니다. 웹 개발, 프로그래밍, 기술 트렌드에 대한 글을 공유합니다.",
    },
    {
      name: "keywords",
      content: "개발 블로그, 웹개발, 프로그래밍, Vue.js, JavaScript, Node.js, 기술 블로그",
    },
    // Open Graph
    {
      property: "og:title",
      content: "Kellog Blog - 개발 블로그",
    },
    {
      property: "og:description",
      content:
        "개발자 redeyes의 개인 블로그입니다. 웹 개발, 프로그래밍, 기술 트렌드에 대한 글을 공유합니다.",
    },
    {
      property: "og:type",
      content: "website",
    },
    {
      property: "og:url",
      content: "https://blog.redeyes.dev",
    },
    // Twitter Card
    {
      name: "twitter:card",
      content: "summary",
    },
    {
      name: "twitter:title",
      content: "Kellog Blog - 개발 블로그",
    },
    {
      name: "twitter:description",
      content:
        "개발자 redeyes의 개인 블로그입니다. 웹 개발, 프로그래밍, 기술 트렌드에 대한 글을 공유합니다.",
    },
  ],
});
</script>

<style scoped>
.home-layout {
  display: grid;
  grid-template-columns: 240px minmax(0, 1fr) 320px;
  gap: 2.5rem;
  width: 100%;
  max-width: 1200px;
  margin: 0 auto;
  min-height: 100vh;
}

.home-hero {
  margin: 1rem auto 1.75rem;
  padding: 1.25rem 1.5rem;
  border: 1px solid var(--border-color);
  border-radius: 1rem;
  background: linear-gradient(135deg, var(--bg-secondary) 0%, var(--accent-soft) 100%);
}

.eyebrow {
  margin: 0;
  font-size: 0.8rem;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--text-secondary);
}

h1 {
  margin: 0.5rem 0 0.35rem;
  font-size: clamp(1.3rem, 3vw, 1.9rem);
  font-weight: 700;
  line-height: 1.35;
}

.description {
  margin: 0;
  color: var(--text-secondary);
}

.tags-wrapper {
  width: 100%;
  flex-shrink: 0;
  position: sticky;
  top: calc(var(--navbar-height) + 1rem);
  margin-top: 0.5rem;
  padding-top: 1rem;
  height: fit-content;
}

@media (max-width: 1180px) {
  .posts-layout {
    grid-template-columns: 220px minmax(0, 1fr);
    gap: 1rem;
  }

  .author-wrapper {
    display: none;
  }

  .tags-wrapper {
    position: sticky;
    top: calc(var(--navbar-height) + 0.75rem);
    margin-top: 0.5rem;
    padding-top: 0;
  }

  .posts-main-wrapper,
  .posts-content-wrapper {
    max-width: 100%;
  }
}

@media (max-width: 768px) {
  .tags-wrapper,
  .author-wrapper {
    display: none;
  }

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

.author-wrapper {
  width: 100%;
  flex-shrink: 0;
  position: sticky;
  top: calc(var(--navbar-height) + 1rem);
  margin-top: 0.5rem;
  padding-top: 1rem;
  height: fit-content;
}
</style>
