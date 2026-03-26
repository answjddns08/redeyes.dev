<template>
  <footer class="simple-footer">
    <div class="footer-content">
      <p>© 2025 redeyes. All rights reserved.</p>
      <p class="footer-subtitle">Built with ❤️ using Vue.js</p>
    </div>
  </footer>
</template>

<!-- <template>
  <footer class="footer">
    <div class="footer-content">
      <p class="footer-description">
        개발과 기술에 대한 이야기를 담은 redeyes의 블로그
      </p>
      <div class="footer-stats">
        <span>{{ totalPosts }} Posts</span>
        <span>•</span>
        <span>Built with Vue.js</span>
      </div>
    </div>
    
    <div class="footer-bottom">
      <p>© 2025 redeyes</p>
    </div>
  </footer>
</template> -->

<script setup>
// 컴포넌트 이름 정의
defineOptions({
  name: "SiteFooter",
});
</script>

<style scoped>
.simple-footer {
  background-color: var(--bg-secondary);
  border-top: 1px solid var(--border-color);
  margin-top: 4rem;
  padding: 2rem 0;
}

.footer-content {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 2rem;
  text-align: center;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.footer-content p {
  color: var(--text-secondary);
  font-size: 0.875rem;
}

.footer-subtitle {
  font-size: 0.75rem !important;
  opacity: 0.8;
}

@media (max-width: 768px) {
  .footer-content {
    padding: 0 1rem;
  }
}
</style>
