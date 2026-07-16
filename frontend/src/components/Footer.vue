<template>
  <footer class="simple-footer">
    <div class="footer-content">
      <p>© {{ new Date().getFullYear() }} redeyes. All rights reserved.</p>
    </div>
  </footer>
</template>

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
