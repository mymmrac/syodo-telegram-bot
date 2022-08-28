<template>
  <div class="px-2">
    <div class="grid grid-cols-1 divide-y divide-tg-hint">
      <div v-for="orderProduct in order.products.values()" :key="orderProduct.product.id" class="py-2">
        <div class="flex flex-nowrap gap-2 items-center justify-between ">
          <div class="aspect-square w-16 shadow rounded bg-white grid place-content-center cursor-pointer">
            <img :src="getImage(originalProduct(orderProduct.product))" :alt="orderProduct.product.title"
                 class="rounded">
          </div>
          <div class="flex-1 truncate" :title="orderProduct.product.title">{{ orderProduct.product.title }}</div>
          <add-remove-buttons fixed-size :amount="orderProduct.amount" :add="() => addProduct(orderProduct)"
                              :remove="() => removeProduct(orderProduct)"/>
        </div>
      </div>
    </div>
    <div class="grid grid-cols-1 gap-2 pt-8">
      <label class="flex justify-between gap-2">
        Не телефонуйте мені
        <input type="checkbox" class="m-checkbox" v-model="order.doNotCall">
      </label>
      <label class="flex justify-between gap-2">
        Без серветок
        <input type="checkbox" class="m-checkbox" v-model="order.noNapkins">
      </label>
      <div class="flex justify-between gap-2">
        <div class="flex-1">Кількість приборів</div>
        <add-remove-buttons fixed-size :amount="order.cutleryCount" :add="() => { order.cutleryCount++ }"
                            :remove="() => { order.cutleryCount-- }"/>
      </div>
      <div class="flex justify-between gap-2">
        <div class="flex-1">Кількість навчальних приборів</div>
        <add-remove-buttons fixed-size :amount="order.trainingCutleryCount"
                            :add="() => { order.trainingCutleryCount++ }"
                            :remove="() => { order.trainingCutleryCount-- }"/>
      </div>
      <label class="flex justify-between gap-2">
        Додати коментар до замовлення
        <input type="checkbox" class="m-checkbox" v-model="order.addComment">
      </label>
      <textarea
          class="form-textarea rounded bg-tg-button text-tg-button-text placeholder-tg-button-text focus:ring-0 border-0 shadow resize-none shadow"
          placeholder="Коментар до замовлення..." rows="3" v-show="order.addComment" @input="updateComment"></textarea>
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

function addProduct(orderProduct: OrderProduct) {
  store.updateInOrder({
    amount: orderProduct.amount + 1,
    product: orderProduct.product,
  })
}

function removeProduct(orderProduct: OrderProduct) {
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

function updateComment(e: Event) {
  const target = e.target as HTMLInputElement
  order.value.comment = target.value.trim()
}
</script>

<style scoped lang="scss">
.m-checkbox {
  @apply form-checkbox rounded focus:ring-0 focus:ring-offset-0 text-tg-button w-8 h-8 border-0 shadow;
}
</style>
