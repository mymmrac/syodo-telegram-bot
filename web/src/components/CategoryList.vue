<template>
  <div class="p-2 grid grid-cols-3 gap-2">
    <div v-for="category in categories" :key="category.id"
         class="grid place-content-center gap-1">
      <img :src="category.icon" :alt="category.title" class="bg-black p-2 rounded">
      <p class="border border-tg-hint rounded text-center">{{ category.title }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Categories } from "@/types"

defineProps<{
  categories: Categories
}>()
</script>
