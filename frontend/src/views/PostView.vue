<template>
  <main>
    <!-- 구조화된 데이터 (JSON-LD) (SEO 용도) -->
    <div type="application/ld+json" class="hidden" v-if="post.title" v-html="structuredData"></div>

    <div class="post-wrap">
      <div class="post-shell">
        <div class="top-spacer"></div>

        <!-- title and extra -->
        <div class="title-section">
          <span class="post-title">{{ post.title }}</span>
          <div class="meta-row">
            <span>redeyes</span>
            <span>-</span>
            <span>{{ post.date }}</span>
          </div>
          <div class="tag-row">
            <RouterLink
              :to="{ path: '/', query: { search: '#' + tag } }"
              class="tagBlock"
              v-for="tag in post.tag"
              v-show="tag"
              :key="tag"
              >{{ tag }}</RouterLink
            >
          </div>
        </div>

        <HeadingList :headings="headings" />

        <!-- main content -->
        <ShowContent
          v-if="post.content"
          :content="post.content"
          :headings="headings"
          @update:headings="headings = $event"
        />
        <div v-else class="text-center py-10">
          <p class="loading">로딩 중</p>
        </div>

        <!-- other posts -->
        <div class="post-nav-row">
          <!-- Previous Post Button -->
          <button
            class="postButton"
            :class="{ 'disabled-button': !previousPost }"
            @click="navigateToPost(previousPost)"
            :disabled="!previousPost"
          >
            <ArrowLeft :size="48" />
            <div class="flex flex-col">
              <span>Previous Post</span>
              <span>{{ previousPost ? previousPost.title : "No Previous Post" }}</span>
            </div>
          </button>

          <!-- Next Post Button -->
          <button
            class="postButton justify-end"
            :class="{ 'disabled-button': !nextPost }"
            @click="navigateToPost(nextPost)"
            :disabled="!nextPost"
          >
            <div class="flex flex-col">
              <span>Next Post</span>
              <span>{{ nextPost ? nextPost.title : "No Next Post" }}</span>
            </div>
            <ArrowRight :size="48" />
          </button>
        </div>

        <!-- comments panel -->
        <!-- <div class="flex flex-col items-end gap-5">
          <input
            class="w-full bg-gray-400 rounded-lg p-4 outline-0"
            placeholder="login before type some comment"
          />
          <button class="bg-blue-500 p-3 rounded-lg">post</button>
        </div> -->
      </div>
    </div>
  </main>
</template>

<script setup>
import { onMounted, ref, watch, computed } from "vue";
import { useHead } from "@unhead/vue";
import { useRoute, RouterLink, useRouter } from "vue-router";
import axios from "axios";
import { usePostStore } from "@/stores/postStore";
import HeadingList from "@/components/HeadingList.vue";
import ShowContent from "@/components/showContent.vue";
import { ArrowLeft, ArrowRight } from "@lucide/vue";

const route = useRoute();
const router = useRouter();

const postStore = usePostStore();

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
 * @type {Post}
 */
const post = ref({
  folder: "",
  title: "",
  date: "",
  tags: [],
  content: "",
  coverImg: "",
});

const previousPost = ref(null);
const nextPost = ref(null);

const headings = ref([]);

// 페이지별 메타 태그 (제목/설명/OG/canonical) - 글 로드 시 반응형 업데이트
useHead(() => {
  const siteName = "redeyes dev";
  const title = post.value.title ? `${post.value.title} | ${siteName}` : siteName;
  const description = (post.value.content || "").substring(0, 160).replace(/[#*`]/g, "").trim();
  const url = `https://blog.redeyes.dev/posts/${route.params.folder}`;
  const image = post.value.coverImg
    ? `https://blog.redeyes.dev/api/posts/${post.value.folder}/${post.value.coverImg}`
    : "https://blog.redeyes.dev/eye.png";

  return {
    title,
    meta: [
      { name: "description", content: description },
      { property: "og:title", content: post.value.title || siteName },
      { property: "og:description", content: description },
      { property: "og:type", content: "article" },
      { property: "og:url", content: url },
      { property: "og:image", content: image },
      { name: "twitter:card", content: "summary_large_image" },
    ],
    link: [{ rel: "canonical", href: url }],
  };
});

// 구조화된 데이터 생성
const structuredData = computed(() => {
  if (!post.value.title) return "";

  return JSON.stringify({
    "@context": "https://schema.org",
    "@type": "BlogPosting",
    headline: post.value.title,
    description: post.value.content
      ? post.value.content.substring(0, 160).replace(/[#*`]/g, "").trim()
      : "",
    author: {
      "@type": "Person",
      name: "redeyes",
      url: "https://blog.redeyes.dev",
    },
    publisher: {
      "@type": "Organization",
      name: "Kellog Blog",
      logo: {
        "@type": "ImageObject",
        url: "https://blog.redeyes.dev/eye.png",
      },
    },
    datePublished: post.value.date,
    dateModified: post.value.date,
    image: post.value.coverImg
      ? `https://blog.redeyes.dev/api/posts/${post.value.folder}/${post.value.coverImg}`
      : "https://blog.redeyes.dev/eye.png",
    url: `https://blog.redeyes.dev/posts/${route.params.folder}`,
    mainEntityOfPage: {
      "@type": "WebPage",
      "@id": `https://blog.redeyes.dev/posts/${route.params.folder}`,
    },
    keywords: post.value.tag ? post.value.tag.join(", ") : "",
    articleSection: "Technology",
  });
});

function navigateToPost(post) {
  if (!post) return;

  router.push({ path: `/posts/${post.folder}` });
}

async function getPostData(folder) {
  const { data } = await axios.get(`https://blog.redeyes.dev/api/posts/${folder}`);

  headings.value = [];

  post.value = data;

  postStore.setCurrentPostId(folder);

  previousPost.value = postStore.previousPost;
  nextPost.value = postStore.nextPost;
}

onMounted(async () => {
  await getPostData(route.params.folder);
});

watch(
  () => route.params.folder,
  async (newFolder) => {
    window.scrollTo(0, 0);
    await getPostData(newFolder);
  },
);
</script>

<style scoped>
.post-wrap {
  width: min(840px, 62vw);
  margin: 0 auto;
}

.post-shell {
  display: flex;
  flex-direction: column;
  width: 100%;
}

.top-spacer {
  margin-bottom: 1.25rem;
}

.title-section {
  display: flex;
  flex-direction: column;
  width: 100%;
  gap: 0.75rem;
  margin-bottom: 2rem;
  padding: 1.35rem;
  border-radius: 1rem;
  border: 1px solid var(--border-color);
  background-color: var(--bg-secondary);
}

.post-title {
  font-size: clamp(1.8rem, 3.8vw, 2.7rem);
  font-weight: 800;
  line-height: 1.3;
}

.meta-row {
  display: flex;
  gap: 0.45rem;
  color: var(--text-secondary);
}

.tag-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.45rem;
}

.tagBlock {
  border-width: 0.1rem;
  border-color: var(--border-color);
  color: var(--text-secondary);
  padding: 0.2rem 0.65rem;
  font-weight: 500;
  font-size: 0.9rem;
  border-radius: 999px;

  transition: ease-out 0.25s;
}

.tagBlock::before {
  content: "#";
  padding-right: 0.5rem;
}

.tagBlock:hover {
  background-color: var(--accent-soft);
  color: var(--accent-color);
}

.post-nav-row {
  display: flex;
  justify-content: space-between;
  width: 100%;
  padding: 1rem 0;
  gap: 0.8rem;
  margin-bottom: 1.25rem;
}

.postButton {
  display: flex;
  flex: 1;
  align-items: center;
  gap: 0.5rem;
  padding: 0.8rem;
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 0.75rem;
  color: var(--text-primary);
  transition: ease-out 0.25s;
}

.postButton:hover {
  border-color: var(--accent-color);
  transform: translateY(-1px);
}

.postButton:disabled {
  pointer-events: none;
}

.disabled-button {
  background-color: var(--bg-primary);
  color: var(--text-secondary);
  cursor: not-allowed;
}

.loading {
  color: var(--text-secondary);
  font-size: 1.5rem;
  font-weight: 500;
}

.loading::after {
  content: ".";
  animation: dotting 2s steps(3, end) infinite;
}

@keyframes dotting {
  0% {
    content: ".";
  }
  33% {
    content: "..";
  }
  66% {
    content: "...";
  }
}

@media (max-width: 768px) {
  .post-nav-row {
    flex-direction: column;
  }

  .postButton {
    justify-content: space-between;
  }
}
</style>
