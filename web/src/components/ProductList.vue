<template>
  <div class="p-2 grid grid-cols-2 gap-2">
    <the-product v-for="product in products" :key="product.id" :product="product"
                 @productUpdate="e => $emit('productUpdate', e)"></the-product>
  </div>
</template>

<script setup lang="ts">
import TheProduct from "@/components/TheProduct.vue"
import { OrderProduct, Products } from "@/types"

defineProps<{
  products: Products
}>()

defineEmits<{
  (e: "productUpdate", product: OrderProduct): void
}>()
</script>
