<template>
  <button :class="scrollPos < 256 ? 'hidden' : ''" @click="scrollToTop('smooth')"
          class="fixed z-20 bottom-8 right-2 text-tg-button-text rounded-xl bg-tg-button p-2 shadow">
    <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" fill="currentColor" viewBox="0 0 16 16">
      <path fill-rule="evenodd"
            d="M8 15a.5.5 0 0 0 .5-.5V2.707l3.146 3.147a.5.5 0 0 0 .708-.708l-4-4a.5.5 0 0 0-.708 0l-4 4a.5.5 0 1 0 .708.708L7.5 2.707V14.5a.5.5 0 0 0 .5.5z"/>
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
