<template>
  <div class="px-2">
    <div class="grid grid-cols-1">
      <transition-group name="m-fade">
        <div v-for="orderProduct in order.products.values()" :key="orderProduct.product.id"
             class="shadow-lg rounded-lg p-2">
          <div class="flex flex-nowrap gap-2 items-center justify-between ">
            <div class="aspect-square w-16 shadow rounded-lg bg-white grid place-content-center">
              <img :src="getImage(originalProduct(orderProduct.product))" :alt="orderProduct.product.title"
                   class="rounded">
            </div>
            <div class="flex-1 flex flex-col truncate" :title="orderProduct.product.title">
              <span class="truncate">{{ orderProduct.product.title }}</span>
              <span class="text-sm -mb-1">{{
                  priceToText(Number(orderProduct.product.price) * orderProduct.amount)
                }}</span>
            </div>
            <add-remove-buttons fixed-size :amount="orderProduct.amount" :add="() => addProduct(orderProduct)"
                                :remove="() => removeProduct(orderProduct)"/>
          </div>
        </div>
      </transition-group>
    </div>
    <div class="grid grid-cols-1 gap-2 pt-8">
      <label class="flex justify-start gap-2">
        <input type="checkbox" class="m-checkbox" v-model="order.doNotCall">
        Не телефонуйте мені
      </label>
      <label class="flex justify-start gap-2">
        <input type="checkbox" class="m-checkbox" v-model="order.noNapkins">
        Без серветок
      </label>
      <div class="flex justify-start gap-2">
        <add-remove-buttons fixed-size :amount="order.cutleryCount" :add="() => { order.cutleryCount++ }"
                            :remove="() => { order.cutleryCount-- }"/>
        <div class="flex-1">Звичайні прибори</div>
      </div>
      <div class="flex justify-start gap-2">
        <add-remove-buttons fixed-size :amount="order.trainingCutleryCount"
                            :add="() => { order.trainingCutleryCount++ }"
                            :remove="() => { order.trainingCutleryCount-- }"/>
        <div class="flex-1">Навчальні прибори</div>
      </div>
      <textarea
          class="form-textarea rounded-lg bg-tg-bg text-tg-text placeholder-tg-text focus:ring-0 border-0 shadow-lg resize-none"
          placeholder="Коментар до замовлення..." rows="3"
          @input="updateComment"></textarea>
    </div>
  </div>
</template>
<script setup lang="ts">
import AddRemoveButtons from "@/components/AddRemoveButtons.vue"

import { storeToRefs } from "pinia"

import { useGlobalStore } from "@/store"
import { getImage, priceToText, OrderProduct, Product } from "@/types"

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
.m-fade {
  &-enter-active,
  &-leave-active {
    transition: all 0.4s ease;
  }

  &-enter-from,
  &-leave-to {
    opacity: 0;
    transform: scale(0.9);
  }
}
</style>
