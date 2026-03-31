<template>
  <main>
    <!-- 구조화된 데이터 (JSON-LD) -->
    <div type="application/ld+json" class="hidden" v-if="post.title" v-html="structuredData"></div>

    <div class="flex w-full justify-center">
      <div class="flex flex-col w-1/2">
        <!-- margin space -->
        <div class="mb-10"></div>

        <!-- title and extra -->
        <div class="flex flex-col w-full gap-4 mb-10">
          <span class="text-5xl font-bold">{{ post.title }}</span>
          <div class="flex gap-2" style="color: var(--text-secondary)">
            <span>redeyes</span>
            <span>-</span>
            <span>{{ post.date }}</span>
          </div>
          <div class="flex gap-3">
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

        <HeadingList :headings="{ headings }" />

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
        <div class="flex justify-between w-full py-3 gap-3 mb-5">
          <!-- Previous Post Button -->
          <button
            class="postButton"
            :class="{ 'disabled-button': !previousPost }"
            @click="navigateToPost(previousPost)"
            :disabled="!previousPost"
          >
            <font-awesome-icon :icon="['fas', 'arrow-left']" size="2xl" />
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
            <font-awesome-icon :icon="['fas', 'arrow-right']" size="2xl" />
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
import { useRoute, RouterLink, useRouter } from "vue-router";
import { useHead } from "@vueuse/head";
import axios from "axios";
import { usePostStore } from "@/stores/postStore";
import HeadingList from "@/components/HeadingList.vue";
import ShowContent from "@/components/showContent.vue";

axios.defaults.withCredentials = true;

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
  tag: [],
  content: "",
  coverImg: "",
});

const previousPost = ref(null);
const nextPost = ref(null);

const headings = ref([]);

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

// SEO를 위한 동적 메타태그 설정
useHead({
  title: () => (post.value.title ? `${post.value.title} - Kellog Blog` : "Kellog Blog"),
  meta: [
    {
      name: "description",
      content: () =>
        post.value.content
          ? post.value.content.substring(0, 160).replace(/[#*`]/g, "").trim() + "..."
          : "개인 블로그 플랫폼 - 기술, 개발, 일상을 공유하는 공간",
    },
    {
      name: "keywords",
      content: () =>
        post.value.tag && post.value.tag.length > 0
          ? post.value.tag.join(", ") + ", 블로그, 개발, 기술"
          : "블로그, 개발, 기술, 프로그래밍",
    },
    // Open Graph
    {
      property: "og:title",
      content: () => post.value.title || "Kellog Blog",
    },
    {
      property: "og:description",
      content: () =>
        post.value.content
          ? post.value.content.substring(0, 160).replace(/[#*`]/g, "").trim() + "..."
          : "개인 블로그 플랫폼 - 기술, 개발, 일상을 공유하는 공간",
    },
    {
      property: "og:type",
      content: "article",
    },
    {
      property: "og:url",
      content: () => `https://blog.redeyes.dev/posts/${route.params.folder}`,
    },
    {
      property: "og:image",
      content: () =>
        post.value.coverImg
          ? `https://blog.redeyes.dev/api/posts/${post.value.folder}/${post.value.coverImg}`
          : "https://blog.redeyes.dev/eye.png",
    },
    {
      property: "article:author",
      content: "redeyes",
    },
    {
      property: "article:published_time",
      content: () => post.value.date,
    },
    {
      property: "article:tag",
      content: () => (post.value.tag ? post.value.tag.join(", ") : ""),
    },
    // Twitter Card
    {
      name: "twitter:card",
      content: "summary_large_image",
    },
    {
      name: "twitter:title",
      content: () => post.value.title || "Kellog Blog",
    },
    {
      name: "twitter:description",
      content: () =>
        post.value.content
          ? post.value.content.substring(0, 160).replace(/[#*`]/g, "").trim() + "..."
          : "개인 블로그 플랫폼 - 기술, 개발, 일상을 공유하는 공간",
    },
    {
      name: "twitter:image",
      content: () =>
        post.value.coverImg
          ? `https://blog.redeyes.dev/api/posts/${post.value.folder}/${post.value.coverImg}`
          : "https://blog.redeyes.dev/eye.png",
    },
  ],
  link: [
    {
      rel: "canonical",
      href: () => `https://blog.redeyes.dev/posts/${route.params.folder}`,
    },
  ],
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
.tagBlock {
  border-width: 0.1rem;

  padding-inline: 0.35rem;

  font-weight: 500;

  font-size: 1rem;

  border-radius: 0.5rem;

  transition: ease-out 0.25s;
}

.tagBlock::before {
  content: "#";
  padding-right: 0.5rem;
}

.tagBlock:hover {
  background-color: var(--shadow);
  color: var(--accent-color);
}

.postButton {
  display: flex;

  flex: 1;

  align-items: center;

  gap: 0.5rem;
  padding: 0.5rem;

  border-width: 0.15rem;
  border-radius: 0.5rem;

  transition: ease-out 0.25s;
}

.postButton:hover {
  background-color: var(--bg-secondary);
  color: var(--text-primary);
}

.postButton:disabled {
  pointer-events: none;
}

.disabled-button {
  background-color: var(--bg-secondary);
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
</style>
