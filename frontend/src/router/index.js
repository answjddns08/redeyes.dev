import { createRouter, createWebHistory } from "vue-router";

// 코드 스플리팅을 위한 동적 import
const HomeView = () => import("../views/HomeView.vue");
const PostView = () => import("@/views/PostView.vue");
const AboutView = () => import("@/views/aboutView.vue");
const NotFoundView = () => import("@/views/notFoundView.vue");

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: "/",
      name: "home",
      component: HomeView,
    },
    {
      path: "/posts/:folder",
      name: "post",
      component: PostView,
    },
    {
      path: "/about",
      name: "about",
      component: AboutView,
    },
    {
      path: "/404",
      name: "notFound",
      component: NotFoundView,
    },
    {
      path: "/:patchMatch(.*)*",
      redirect: "/404",
    },
  ],
});

export default router;
