import "./assets/style.css"; //tailwind css
import "./assets/scrollbar.css"; //custom scrollbar styles

import { createApp } from "vue";
import { createPinia } from "pinia";
import { createGtag } from "vue-gtag";

import App from "./App.vue";
import router from "./router";
import { createHead } from "@unhead/vue/client";

const app = createApp(App);
const head = createHead();

app.use(createPinia());
app.use(router);
app.use(head);

// for Google Analytics
if (import.meta.env.PROD) {
  app.use(
    createGtag({
      tagId: "G-XMP7KKHEZN",
      pageTracker: {
        router,
      },
    }),
  );
}

app.mount("#app");
