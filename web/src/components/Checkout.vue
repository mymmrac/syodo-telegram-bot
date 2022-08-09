<template>
  <div class="grid grid-cols-1 p-2">
    <div v-for="orderProduct in order.products.values()" :key="orderProduct.product.id"
         class="flex justify-between items-center">
      <p>{{ orderProduct.product.title }}</p>
      <add-remove-buttons class="w-1/3" :amount="orderProduct.amount" :add="() => add(orderProduct)"
                          :remove="() => remove(orderProduct)"/>
    </div>
  </div>
</template>

<script setup lang="ts">
import AddRemoveButtons from "@/components/AddRemoveButtons.vue"

import { storeToRefs } from "pinia"

import { useGlobalStore } from "@/store"
import { OrderProduct } from "@/types"

const store = useGlobalStore()
const { order } = storeToRefs(store)

function add(orderProduct: OrderProduct) {
  store.updateInOrder({
    amount: orderProduct.amount + 1,
    product: orderProduct.product,
  })
}

function remove(orderProduct: OrderProduct) {
  store.updateInOrder({
    amount: orderProduct.amount - 1,
    product: orderProduct.product,
  })
}
</script>
