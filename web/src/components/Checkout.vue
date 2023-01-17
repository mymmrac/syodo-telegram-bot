<template>
  <div class="px-2">
    <div class="text-center text-xl my-2">Оформити замовлення</div>

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
              <span class="text-sm">{{
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
      <label class="flex justify-start items-center gap-2">
        <input type="checkbox" class="m-checkbox" v-model="order.doNotCall">
        Не телефонуйте мені
      </label>
      <label class="flex justify-start items-center gap-2">
        <input type="checkbox" class="m-checkbox" v-model="order.noNapkins">
        Без серветок
      </label>
      <div class="flex justify-start items-center gap-2">
        <add-remove-buttons fixed-size :amount="order.cutleryCount" :add="() => { order.cutleryCount++ }"
                            :remove="() => { order.cutleryCount-- }"/>
        <div class="flex-1">Звичайні прибори</div>
      </div>
      <div class="flex justify-start items-center gap-2">
        <add-remove-buttons fixed-size :amount="order.trainingCutleryCount"
                            :add="() => { order.trainingCutleryCount++ }"
                            :remove="() => { order.trainingCutleryCount-- }"/>
        <div class="flex-1">Навчальні прибори</div>
      </div>
      <textarea class="m-textarea" placeholder="Коментар до замовлення..." rows="3" maxlength="2048"
                @input="updateComment"></textarea>

      <label class="flex flex-col">
        <span class="ml-1">Ім'я*</span>
        <input type="text" placeholder="..." class="m-input" maxlength="64" v-model.trim="order.name" required/>
      </label>
      <label class="flex flex-col">
        <span class="ml-1">Телефон*</span>
        <input type="text" placeholder="..." class="m-input" maxlength="13" v-model.trim="order.phone" required/>
      </label>

      <div>Спосіб доставки</div>
      <label class="flex justify-start items-center gap-2">
        <input type="radio" value="delivery" checked class="m-radio" v-model="order.deliveryType"/>
        Доставка
      </label>
      <label class="flex justify-start items-center gap-2">
        <input type="radio" value="self_pickup_1" class="m-radio" v-model="order.deliveryType"/>
        Самовивіз (вул. Трускавецька, 2a)
      </label>
      <label class="flex justify-start items-center gap-2">
        <input type="radio" value="self_pickup_2" class="m-radio" v-model="order.deliveryType"/>
        Самовивіз (вул. Mалоголосківська, 28)
      </label>

      <label class="flex flex-col">
        <span class="ml-1">Акція</span>
        <select class="m-select" v-model="order.promotion" required>
          <option value="" selected>Акція Відсутня</option>
          <option value="4+1" :disabled="!promo4Plus1Available">Акція Роли 4+1</option>
          <option value="Самовивіз" :disabled="!promoSelfPickupAvailable">Самовивіз -10%</option>
        </select>
      </label>

      <transition name="m-fade">
        <label class="flex flex-col" v-show="order.deliveryType === 'delivery'">
          <span class="ml-1">Місто*</span>
          <input type="text" placeholder="..." disabled class="m-input" maxlength="128" v-model.trim="order.city"
                 required/>
        </label>
      </transition>
      <transition name="m-fade">
        <label class="flex flex-col" v-show="order.deliveryType === 'delivery'">
          <span class="ml-1">Адреса*</span>
          <input type="text" placeholder="..." class="m-input" maxlength="512" v-model.trim="order.address" required/>
        </label>
      </transition>
      <transition name="m-fade">
        <label class="flex flex-col" v-show="order.deliveryType === 'delivery'">
          <span class="ml-1">Під'їзд</span>
          <input type="text" placeholder="..." class="m-input" maxlength="512" v-model.trim="order.entrance"/>
        </label>
      </transition>
      <transition name="m-fade">
        <label class="flex flex-col" v-show="order.deliveryType === 'delivery'">
          <span class="ml-1">Домофон</span>
          <input type="text" placeholder="..." class="m-input" maxlength="512" v-model.trim="order.eCode"/>
        </label>
      </transition>
      <transition name="m-fade">
        <label class="flex flex-col" v-show="order.deliveryType === 'delivery'">
          <span class="ml-1">Поверх</span>
          <input type="text" placeholder="..." class="m-input" maxlength="512" v-model.trim="order.floor"/>
        </label>
      </transition>
      <transition name="m-fade">
        <label class="flex flex-col" v-show="order.deliveryType === 'delivery'">
          <span class="ml-1">Квартира</span>
          <input type="text" placeholder="..." class="m-input" maxlength="512" v-model.trim="order.apartment"/>
        </label>
      </transition>
    </div>
  </div>
</template>
<script setup lang="ts">
import AddRemoveButtons from "@/components/AddRemoveButtons.vue"

import { storeToRefs } from "pinia"

import { useGlobalStore } from "@/store"
import { getImage, OrderProduct, priceToText, Product } from "@/types"
import { Ref, ref, watch } from "vue"

const store = useGlobalStore()
const { order } = storeToRefs(store)

const promo4Plus1Available: Ref<boolean> = ref(false)
const promoSelfPickupAvailable: Ref<boolean> = ref(false)

watch(() => [ order.value.deliveryType, order.value.products ], () => {
  if (order.value.deliveryType == "delivery") {
    promoSelfPickupAvailable.value = false
    promo4Plus1Available.value = false

    let promoCount = 0
    for (const [ _, product ] of order.value.products) {
      if ([ "7", "14" ].includes(product.product.category_id)) {  // Роли, Без лактози
        promoCount += product.amount
      }

      if (promoCount > 4) {
        promo4Plus1Available.value = true
        break
      }
    }

    if (order.value.promotion == "Самовивіз" || !promo4Plus1Available.value) {
      order.value.promotion = promo4Plus1Available.value ? "4+1" : ""
    }
  } else {
    promoSelfPickupAvailable.value = true
    promo4Plus1Available.value = false

    if (order.value.promotion == "4+1") {
      order.value.promotion = "Самовивіз"
    }
  }
}, { deep: true })

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
