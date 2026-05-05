<template>
  <div class="author-field">
    <div class="bg-image-container"><img src="/background.webp" class="cover-img" /></div>
    <div class="author-image-container">
      <img src="/eye.png" class="profile-img" />
    </div>

    <div class="author-info">
      <p class="author">redeyes</p>
      <p class="description">눈(snow 아님)이 너무 빨감</p>
      <div class="social-links">
        <a href="https://github.com/answjddns08" target="_blank">
          <font-awesome-icon :icon="['fab', 'github']" class="icon" size="2xl" />
        </a>
      </div>
    </div>
  </div>
</template>

<script setup></script>

<style scoped>
.author-field {
  display: flex;
  flex-direction: column;
  align-items: center;
  border: 1px solid var(--border-color);
  border-radius: 1rem;
  background-color: var(--bg-secondary);
  padding-bottom: 1rem;
}

.bg-image-container {
  width: 100%;
  height: 11.5rem;
  border-radius: 0.75rem;
  background-color: #aaaaaa;
  overflow: hidden;
}

.cover-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.author-image-container {
  width: 6rem;
  height: 6rem;
  margin-top: -2.75rem;
  border-radius: 50%;
  border: 3px solid var(--bg-secondary);
  background-color: #ed0000;
  overflow: hidden;
  z-index: 1;
}

.profile-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.author-info {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  align-items: center;
  justify-content: center;
  margin-top: 0.35rem;
}

.author {
  color: var(--text-primary);
  font-weight: 600;
  font-size: 1.25rem;
}

.description {
  color: var(--text-secondary);
  font-weight: 500;
  font-size: 0.95rem;
  margin-bottom: 0.3rem;
}

.social-links {
  display: flex;
  gap: 0.75rem;
  justify-content: center;
}

.icon {
  color: var(--text-secondary);
  padding: 0.5rem;
  transition: all 0.3s ease;
}

.icon:hover {
  color: var(--text-primary);

  scale: 1.2;
}
</style>
