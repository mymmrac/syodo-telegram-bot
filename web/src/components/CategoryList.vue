<template>
  <div class="p-2 grid grid-cols-2 gap-2">
    <div v-for="category in categories" :key="category.id" @click="selectedCategory = category.id"
         class="flex gap-2 cursor-pointer rounded-lg transition duration-200"
         :class="selectedCategory === category.id ? 'bg-tg-button shadow-lg' : ''">
      <img :src="category.icon" :alt="category.title" class="bg-tg-button p-2 w-14 rounded-lg"
           :class="selectedCategory === category.id ? '' : 'shadow-lg'">
      <p class="rounded-lg text-center text-xl w-full p-2 flex items-center transition duration-200"
         :class="selectedCategory === category.id ? 'text-tg-button-text' : ''">
        {{ category.title }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { storeToRefs } from "pinia"

import { categories } from "@/definitions"
import { useGlobalStore } from "@/store"

const store = useGlobalStore()
const { selectedCategory } = storeToRefs(store)
</script>
