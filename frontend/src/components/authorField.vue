<template>
  <div class="author-field">
    <div class="bg-image-container"><img src="/background.webp" class="object-cover" /></div>
    <div class="author-image-container">
      <img src="/eye.png" class="object-cover" />
    </div>

    <div class="flex flex-col gap-3 items-center justify-center">
      <p class="author">redeyes</p>
      <p class="description">눈(snow)이 너무 빨감</p>
      <div class="flex gap-5 justify-between">
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
}

.bg-image-container {
  width: 25rem;
  height: 15rem;

  border-radius: 0.75rem;
  background-color: #aaaaaa;

  overflow: hidden;
}

.author-image-container {
  width: 6rem;
  height: 6rem;

  margin-top: -3rem;

  border-radius: 50%;

  background-color: #ed0000;

  overflow: hidden;

  z-index: 1;
}

.author {
  color: var(--text-primary);
  font-weight: 600;
  font-size: 1.25rem;
}

.description {
  color: var(--text-secondary);
  
  font-weight: 500;
  font-size: 1rem;

  margin-bottom: .5rem;
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
