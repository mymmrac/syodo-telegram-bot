<template>
  <button :class="scrollPos < 256 ? 'hidden' : ''" @click="scrollToTop('smooth')"
          class="fixed z-20 bottom-8 right-2 text-tg-button rounded-full bg-tg-button-text shadow-xl">
    <svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" fill="currentColor" viewBox="0 0 16 16">
      <path
          d="M16 8A8 8 0 1 0 0 8a8 8 0 0 0 16 0zm-7.5 3.5a.5.5 0 0 1-1 0V5.707L5.354 7.854a.5.5 0 1 1-.708-.708l3-3a.5.5 0 0 1 .708 0l3 3a.5.5 0 0 1-.708.708L8.5 5.707V11.5z"/>
    </svg>
  </button>
</template>

<script setup lang="ts">
import { onMounted, ref, Ref } from "vue"
import { scrollToTop } from "@/utils"

const scrollPos: Ref<number> = ref(window.scrollY)
onMounted(() => {
  window.addEventListener("scroll", () => {
    scrollPos.value = window.scrollY
  })
})
</script>
