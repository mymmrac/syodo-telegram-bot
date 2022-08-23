<template>
  <div class="grid grid-cols-1 px-2 divide-y divide-tg-hint">
    <div v-for="orderProduct in order.products.values()" :key="orderProduct.product.id" class="py-2">
      <div class="flex flex-nowrap gap-2 items-center justify-between ">
        <div class="aspect-square w-16 shadow rounded bg-white grid place-content-center cursor-pointer">
          <img :src="getImage(originalProduct(orderProduct.product))" :alt="orderProduct.product.title"
               class="rounded">
        </div>
        <div class="flex-1 truncate" :title="orderProduct.product.title">{{ orderProduct.product.title }}</div>
        <add-remove-buttons class="w-24 mt-0" :amount="orderProduct.amount" :add="() => add(orderProduct)"
                            :remove="() => remove(orderProduct)"/>
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import AddRemoveButtons from "@/components/AddRemoveButtons.vue"

import { storeToRefs } from "pinia"

import { useGlobalStore } from "@/store"
import { getImage, OrderProduct, Product } from "@/types"

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

function originalProduct(product: Product): Product {
  if (product.linkedPosition) {
    return product
  }

  const p = store.linkedFromProduct(product)
  return p ? p : product
}
</script>
