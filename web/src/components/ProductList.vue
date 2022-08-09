<template>
  <div class="p-2 grid grid-cols-2 gap-2">
    <transition name="m-fade">
      <div v-show="selectedCategory === hasSubCategoriesCategory && store.isSearchEmpty"
           class="grid grid-cols-3 gap-2 col-span-2">
        <div v-for="subCategory in subCategories" :key="subCategory.id"
             @click="scrollToID(`sub-category-${subCategory.id}`)"
             class="cursor-pointer bg-tg-button text-tg-button-text rounded text-center">
          {{ subCategory.title }}
        </div>
      </div>
    </transition>
    <transition-group name="m-fade">
      <template v-for="object in objects" :key="object.id">
        <template v-if="isProduct(object)">
          <the-product v-show="match(object)"
                       :product="object" :linked-product="linkedProduct(object)"
                       @productUpdate="e => $emit('productUpdate', e)"/>
        </template>
        <div v-else-if="selectedCategory === hasSubCategoriesCategory && store.isSearchEmpty" class="rounded p-2 col-span-2"
             :id="`sub-category-${object.id}`">
          <p class="border-b-2 border-tg-hint pt-2 pb-1 text-xl">{{ object.title }}</p>
        </div>
      </template>
    </transition-group>
  </div>
  <div class="h-[96px]"></div>
</template>

<script setup lang="ts">
import TheProduct from "@/components/TheProduct.vue"

import { storeToRefs } from "pinia"

import { isProduct, Objects, OrderProduct, Product, Products } from "@/types"
import { hasSubCategoriesCategory, subCategories } from "@/definitions"
import { useGlobalStore } from "@/store"
import { scrollToID } from "@/utils"

const store = useGlobalStore()
const { selectedCategory, search } = storeToRefs(store)

const props = defineProps<{
  allProducts: Products
  objects: Objects
}>()

defineEmits<{
  (e: "productUpdate", product: OrderProduct): void
}>()

function linkedProduct(product: Product): Product | undefined {
  if (!product.linkedPosition) {
    return undefined
  }

  return props.allProducts.find(p => p.id == product.linkedPosition)
}

function match(product: Product): boolean {
  if (product.category_id !== selectedCategory.value) {
    return false
  }

  if (store.isSearchEmpty) {
    return true
  }

  const data = `${ product.title.toLowerCase() } ${ product.description.toLowerCase() }`
  const searchWords = search.value.toLowerCase().split(" ")
  for (const searchWord of searchWords) {
    let s = searchWord
    let p = true

    if (searchWord.startsWith("-")) {
      s = searchWord.substring(1)
      p = false
    }

    if ((p && !data.includes(s)) || (!p && data.includes(s))) {
      return false
    }
  }

  return true
}
</script>

<style scoped lang="scss">
.m-fade {
  &-enter-active,
  &-leave-active {
    transition: all 0.4s ease;
  }

  &-enter-from,
  &-leave-to {
    opacity: 0.2;
    transform: scale(0.9);
  }
}
</style>
