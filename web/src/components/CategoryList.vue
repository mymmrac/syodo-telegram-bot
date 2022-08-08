<template>
  <div class="p-2 grid grid-cols-3 gap-2">
    <div v-for="category in categories" :key="category.id" @click="$emit('categorySelected', category.id)"
         class="grid place-content-center gap-1 cursor-pointer">
      <img :src="category.icon" :alt="category.title" class="bg-black p-2 rounded">
      <p class="rounded text-center"
         :class="selectedCategory === category.id ? 'bg-tg-button text-tg-button-text' : ''">
        {{ category.title }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Categories } from "@/types"

defineProps<{
  selectedCategory: string
  categories: Categories
}>()

defineEmits<{
  (e: "categorySelected", id: string): void
}>()
</script>
