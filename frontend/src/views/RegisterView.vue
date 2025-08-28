<template>
  <div>
    <div class="flex flex-col items-center gap-7">
      <p class="text-5xl font-bold p-6">회원가입</p>
      <form class="flex flex-col gap-10">
        <label class="flex flex-col font-semibold text-2xl gap-2">
          이름
          <input class="bg-slate-300 border rounded-lg" v-model="name" />
        </label>
        <label class="flex flex-col font-semibold text-2xl">
          아이디
          <div class="flex items-center">
            <input class="bg-slate-300 border rounded-lg mr-10" v-model="id" required />
            <button
              type="button"
              class="bg-blue-500 transition-colors duration-250 text-lg rounded-lg p-3 text-white"
              @click="isSameId"
            >
              중복 확인
            </button>
          </div>
          <p
            v-if="isSameUserId != null"
            class="text-xl font-semibold"
            :class="isSameUserId ? 'text-red-500' : 'text-green-500'"
          >
            {{ isSameUserId ? "이미 사용중인 아이디입니다." : "사용 가능한 아이디입니다." }}
          </p>
        </label>
        <label class="flex flex-col font-semibold text-2xl gap-2">
          비밀번호
          <input class="bg-slate-300 border rounded-lg" type="password" v-model="password" />
        </label>
      </form>
      <p v-if="message" class="font-bold text-2xl">{{ message }}</p>
      <button
        class="bg-blue-500 text-white text-2xl py-4 px-40 rounded-lg font-semibold"
        @click="signUp"
      >
        회원가입
      </button>
    </div>
  </div>
</template>

<script setup>
import router from "@/router";
import axios from "axios";
import { ref } from "vue";

const name = ref("");
const id = ref("");
const password = ref("");

const isSameUserId = ref(null);
const message = ref("");

const isSameId = async () => {
  if (id.value == "") {
    message.value = "아이디를 입력하세요.";
    return;
  }

  try {
    const res = await axios.get(`https://blog.redeyes.dev/api/user/${id.value}`);

    if (res.data) {
      isSameUserId.value = true;
    } else {
      isSameUserId.value = false;
    }
  } catch (error) {
    //console.error(error);

    if (error.response.status == 404) {
      isSameUserId.value = false;
    }
  }
};

const signUp = async () => {
  if (name.value == "" || id.value == "" || password.value == "") {
    message.value = "모든 항목을 입력하세요.";
    return;
  }

  if (isSameUserId.value == null || isSameUserId.value) {
    message.value = "아이디 중복 확인을 해주세요.";
    return;
  }

  try {
    const res = await axios.post("https://blog.redeyes.dev/api/signUp", {
      name: name.value,
      id: id.value,
      password: password.value,
    });
    console.log(res.data);
    router.push("/");
  } catch (error) {
    console.error(error);
  }
};
</script>

<style scoped></style>
