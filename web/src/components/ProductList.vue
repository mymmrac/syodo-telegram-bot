<template>
  <div class="p-2 grid grid-cols-2 gap-2">
    <transition-group name="m-fade">
      <the-product v-for="product in products" :key="product.id" v-show="match(product)"
                   :product="product" :linked-product="linkedProduct(product)"
                   @productUpdate="e => $emit('productUpdate', e)"></the-product>
    </transition-group>
  </div>
  <div class="h-[96px]"></div>
</template>

<script setup lang="ts">
import TheProduct from "@/components/TheProduct.vue"
import { OrderProduct, Product, Products } from "@/types"

const props = defineProps<{
  allProducts: Products
  products: Products
  category: string
  search: string
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
  if (product.category_id !== props.category) {
    return false
  }

  if (props.search === "") {
    return true
  }

  const data = `${ product.title.toLowerCase() } ${ product.description.toLowerCase() }`
  const searchWords = props.search.toLowerCase().split(" ")
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
