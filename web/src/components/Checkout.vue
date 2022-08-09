<template>
  <div class="grid grid-cols-1 p-2">
    <div v-for="orderProduct in order.products.values()" :key="orderProduct.product.id"
         class="flex justify-between items-center">
      <p>{{ orderProduct.product.title }}</p>
      <p>{{ orderProduct.amount }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Order } from "@/types"

defineProps<{
  order: Order
}>()
</script>
