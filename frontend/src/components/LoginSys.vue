<template>
  <button
    class="bg-gray-400 p-2.5 rounded-lg text-white hover:bg-gray-600 transition"
    @click="modalOpen"
  >
    login
  </button>

  <!-- modal -->
  <div class="modal-wrap" v-show="isModalOpen" @click="modalOpen">
    <div class="modal-container" @click.prevent="modalOpen">
      <div class="relative flex w-full">
        <button class="X-btn" @click="modalOpen">
          <font-awesome-icon :icon="['fas', 'xmark']" size="2xl" />
        </button>
      </div>
      <span class="font-bold text-3xl mb-7">Login</span>
      <form class="flex flex-col items-center mb-7 gap-7">
        <div>
          <label class="font-bold text-xl rounded-lg" for="user_id">ID :</label>
          <input
            id="user_id"
            type="text"
            placeholder="insert your id"
            class="font-bold text-xl rounded-lg p-3 outline-blue-500"
          />
        </div>
        <div>
          <label class="font-bold text-xl" for="password"> Password : </label>
          <input
            id="password"
            type="password"
            placeholder="insert your password"
            class="font-bold text-xl rounded-lg p-3 outline-blue-500"
          />
        </div>
      </form>
      <div class="flex w-full ml-12 mb-6">
        <span class="text-gray-500 font-semibold">계정이 없나요? </span>
        <RouterLink
          to="/register"
          @click="modalOpen"
          class="hover:underline text-gray-500 font-semibold"
          >회원가입</RouterLink
        >
      </div>
      <button class="modal-btn" @click="modalOpen">로그인</button>
    </div>
  </div>
</template>

<script setup>
import { ref } from "vue";

const isModalOpen = ref(false);

function modalOpen() {
  isModalOpen.value = !isModalOpen.value;
}
</script>

<style scoped>
/* dimmed */
.modal-wrap {
  position: fixed;
  left: 0;
  top: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.4);
  z-index: 10;
}

/* modal or popup */
.modal-container {
  position: relative;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 35rem;
  background: #fff;
  border-radius: 1rem;
  padding: 1rem;
  box-sizing: border-box;
  z-index: 11;

  display: flex;
  flex-direction: column;
  align-items: center;
}

.X-btn {
  position: absolute;

  cursor: pointer;

  right: 0rem;

  transition: all 0.2s ease;
}

.X-btn:hover {
  color: aqua;
}

.modal-btn {
  padding: 0.5rem;
  width: 15rem;
  height: 3.25rem;
  border-radius: 0.5rem;
  color: white;
  font-weight: 700;
  font-size: 1.25rem;
  line-height: 1.25rem;
  background-color: #2663e8;
  transition: background-color 0.2s ease;

  cursor: pointer;
}

.modal-btn:hover {
  background-color: #1a4bb9; /* 호버 시 변경할 색상 */
}
</style>
