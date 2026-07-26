<template>
  <div class="about-page">
    <section>
      <img src="/eye.png" class="profile-img" />
      <h2>Redeyes</h2>
      <!-- <p class="font-semibold italic text-xl mb-6" style="color: var(--text-secondary)"> -->
      <!--   Higher and higher! -->
      <!-- </p> -->
      <div class="flex gap-3">
        <a href="https://github.com/answjddns08" target="_blank">
          <font-awesome-icon :icon="['fab', 'github']" class="icon" size="2xl" />
        </a>
        <a href="mailto:ansgjddns0808@gmail.com">
          <font-awesome-icon icon="fa-solid fa-envelope" class="icon" size="2xl" />
        </a>
      </div>
    </section>
    <section>
      <h2>About</h2>
      <p class="mb-3">
        안녕하세요! 학생 개발자 redeyes입니다. 초등학생 때 블록 코딩을 접한 걸 시작으로 코딩에
        관심을 가지게 되었고, 중학생 때 게임 개발에 관심을 가져 유니티를 만지다가, 고등학생 때
        웹개발에 관심을 가지게 되어 웹 개발을 시작하게 되었습니다. 지금도 웹 개발을 주로 하고 있고
        간단한 블로그 사이트부터 웹 게임까지 다양한 프로젝트를 진행하고 있습니다.
      </p>

      <p>
        라즈베리파이로 본격적인 홈 서버 구축을 시작하면서 리눅스, 서버 관리에도 관심을 가지게 되었고
        공부도 하고 있습니다. 리소스가 제한적인 컴퓨터라 개발할 때 최적화에 대한 고민을 많이 하게
        되었고, 그 과정에서 성능 최적화에 대한 관심도 가지게 되었습니다. 개발할 때 제가 불편한
        것들이나 친구들의 불편한 것들을 확인하고 그걸 해결할 수 있는 서비스나 프로그램을 만드는 걸
        좋아합니다.
      </p>
    </section>
    <section>
      <h2>Project</h2>
      <div class="projects">
        <section class="project-card">
          <img src="/aboutblog.webp" alt="project-card" />
          <div>
            <h3>블로그 개발</h3>
            <p>
              프론트엔드와 백엔드를 모두 직접 개발한 블로그 사이트입니다. 프론트엔드는 Vue.js와
              Tailwind CSS를 사용하여 개발하였고, 백엔드는 Golang과 SQLite를 사용하여
              개발하였습니다.
            </p>
            <ul>
              <li>Vue.js</li>
              <li>Javascript</li>
              <li>Tailwind CSS</li>
              <li>SQLite</li>
              <li>Golang</li>
            </ul>
          </div>
        </section>
        <section class="project-card">
          <img src="/aboutBot.png" alt="project-card" />
          <div>
            <h3>디스코드 봇 개발</h3>
            <p>
              discord.js를 활용하여 디스코드 봇을 만들어 라즈베리파이에 있는 서버를 켜고 끌 수
              있도록 만든 프로젝트입니다.
            </p>
            <ul>
              <li>Javascript</li>
              <li>Discord.js</li>
              <li>Raspberry Pi</li>
              <li>Docker</li>
            </ul>
          </div>
        </section>
        <section class="project-card">
          <img src="/aboutherb.webp" alt="project-card" />
          <div>
            <h3>웹 게임 허브 제작</h3>
            <p>
              디스코드 활동에 넣을 수 있는 웹 게임 허브를 제작한 프로젝트입니다. 일반적인 html, css,
              js로 만든 게임, phaser.js로 만든 2D 게임, three.js로 만든 3D 게임들을 담은 허브를
              만들었습니다.
            </p>
            <ul>
              <li>Vue.js</li>
              <li>Tailwind CSS</li>
              <li>Javascript</li>
              <li>Phaser.js</li>
              <li>Three.js</li>
              <li>Discord</li>
            </ul>
          </div>
        </section>
      </div>
    </section>
  </div>
</template>

<script setup>
//TODO:이미지 넣기

//TODO:UI조금 개선해야 할 듯?
</script>

<style scoped>
.about-page {
  display: flex;

  width: 37.5vw;
  margin-left: auto;
  margin-right: auto;

  flex-direction: column;

  justify-content: center;

  align-items: center;

  padding: 4rem 1.25rem;

  gap: 2rem;
}

.profile-img {
  width: 7rem;
  height: 7rem;

  border-radius: 50%;

  margin-bottom: 1rem;

  object-fit: cover;
}

p {
  text-align: justify;

  width: 100%;

  animation: fadeIn;
  animation-timeline: view();
}

h1 {
  font-size: 2.5rem;
  font-weight: bold;
  color: var(--text-color);
}

h2 {
  font-size: 1.75rem;
  font-weight: bold;
  color: var(--text-color);

  margin-bottom: 0.75rem;
}

h3 {
  font-size: 1.25rem;
  font-weight: bold;
  color: var(--text-color);

  margin-bottom: 0.5rem;
}

section {
  display: flex;
  flex-direction: column;

  width: 100%;

  margin-top: 1.5rem;
}

.icon {
  color: var(--text-secondary);
  padding: 0.5rem;
  transition: all 0.2s ease;
}

.icon:hover {
  color: var(--text-primary);

  scale: 1.2;
}

.projects:hover .project-card {
  opacity: 0.5;
}

.projects .project-card:hover {
  opacity: 1;
}

.project-card {
  display: flex;
  flex-direction: row;

  gap: 1rem;

  width: 100%;

  padding: 1rem;

  border-radius: 0.5rem;

  border: 1px solid var(--text-secondary);

  transition: all 0.2s ease;
}

.project-card h3 {
  transition: all 0.2s ease;
}

.project-card:hover h3 {
  color: var(--accent-color);
}

.project-card img {
  max-width: 32.5%;
  max-height: 45%;

  border-radius: 0.5rem;

  object-fit: cover;

  border: 1px solid var(--text-secondary);
}

.project-card ul {
  display: flex;
  flex-direction: row;

  gap: 0.5rem;

  margin-top: 0.5rem;

  flex-wrap: wrap;
}

.project-card ul li {
  padding: 0.25rem 0.5rem;

  height: fit-content;

  border-radius: 0.25rem;

  background-color: var(--text-secondary);

  color: var(--bg-secondary);
}
</style>
