<template>
  <div class="p-2 grid grid-cols-2 gap-2">
    <transition-group name="m-fade">
      <the-product v-for="product in products" :key="product.id" v-show="product.category_id === category"
                   :product="product" @productUpdate="e => $emit('productUpdate', e)"></the-product>
    </transition-group>
  </div>
  <div class="h-[96px]"></div>
</template>

<script setup lang="ts">
import TheProduct from "@/components/TheProduct.vue"
import { OrderProduct, Products } from "@/types"

defineProps<{
  products: Products
  category: string
}>()

defineEmits<{
  (e: "productUpdate", product: OrderProduct): void
}>()
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
