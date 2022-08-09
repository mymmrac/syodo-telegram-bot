<template>
  <transition name="m-fade" mode="out-in">
    <div v-show="!checkout">
      <category-list/>
      <div class="w-full px-2 pb-2 flex gap-2">
        <input type="text" placeholder="Пошук..." v-model.trim="search"
               class="p-2 flex-1 rounded border-none ring-0 focus:ring-0 bg-tg-button text-tg-button-text placeholder-tg-button-text">
        <button class="rounded px-2" @click="store.clearSearch">
          <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" fill="currentColor"
               class="bi bi-backspace-fill" viewBox="0 0 16 16">
            <path
                d="M15.683 3a2 2 0 0 0-2-2h-7.08a2 2 0 0 0-1.519.698L.241 7.35a1 1 0 0 0 0 1.302l4.843 5.65A2 2 0 0 0 6.603 15h7.08a2 2 0 0 0 2-2V3zM5.829 5.854a.5.5 0 1 1 .707-.708l2.147 2.147 2.146-2.147a.5.5 0 1 1 .707.708L9.39 8l2.146 2.146a.5.5 0 0 1-.707.708L8.683 8.707l-2.147 2.147a.5.5 0 0 1-.707-.708L7.976 8 5.829 5.854z"/>
          </svg>
        </button>
      </div>
      <hr>
      <product-list @productUpdate="updateOrder"/>
      <go-to-top-button/>
    </div>
  </transition>
  <transition name="m-fade" mode="in-out">
    <checkout v-show="checkout" :order="order"></checkout>
  </transition>
</template>

<script setup lang="ts">
import ProductList from "@/components/ProductList.vue"
import CategoryList from "@/components/CategoryList.vue"
import Checkout from "@/components/Checkout.vue"
import GoToTopButton from "@/components/GoToTopButton.vue"

import { TelegramWebApps } from "telegram-bots-webapps-types"
import { computed, ComputedRef, Ref, ref, watch } from "vue"
import { storeToRefs } from "pinia"

import { Order, OrderProduct, priceToText, Products } from "@/types"
import { scrollToTop, sendError } from "@/utils"
import { useGlobalStore } from "@/store"
import syodoAPI from "@/syodo-api"

const tg: TelegramWebApps.WebApp = window.Telegram.WebApp

// Version check
const [ major, minor ] = tg.version.split(".").map(Number)
if (major < 6 || (major == 6 && minor < 1)) {
  sendError("old-version", tg.version)
}

const store = useGlobalStore()
const { loaded, allProducts, search } = storeToRefs(store)

// Loaders
watch(loaded, (isLoaded) => {
  if (!isLoaded) {
    return
  }

  console.log("Loaded")
  tg.ready()
})

// Products
syodoAPI.get<Products>("/products")
    .then(response => {
      if (response.status !== 200) {
        console.error(response)
        sendError("load-products", "Хмм, не вдалося завантажити меню")
        return
      }
      allProducts.value = response.data
    })
    .catch(err => {
      console.error(err)
      sendError("load-products", "Хмм, не вдалося завантажити меню")
    })
    .finally(() => loaded.value = true)

// Order
const order: Ref<Order> = ref(<Order>{
  products: new Map<string, OrderProduct>(),
})

function updateOrder(product: OrderProduct) {
  if (product.amount == 0) {
    order.value.products.delete(product.id)
  } else {
    order.value.products.set(product.id, product)
  }

  if (order.value.products.size !== 0) {
    tg.MainButton.setText("Переглянути замовлення")
    tg.MainButton.show()
  } else {
    tg.MainButton.hide()
  }
}

tg.MainButton.onClick(() => {
  if (!checkout.value) {
    checkout.value = true
    scrollToTop()

    tg.BackButton.show()
    tg.MainButton.setText(`Замовити - ${ priceToText(totalPrice.value) }`)
  } else {
    alert("OK")
  }
})

tg.BackButton.onClick(() => {
  checkout.value = false
  scrollToTop()

  tg.BackButton.hide()
  tg.MainButton.setText("Переглянути замовлення")
})

// Checkout
const checkout: Ref<boolean> = ref(false)

const totalPrice: ComputedRef<number> = computed(() => {
  let price = 0
  order.value.products.forEach(p => price += Number(p.product.price) * p.amount)
  return price
})
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

  &-leave-active {
    position: absolute;
    left: 0;
    right: 0;
  }
}
</style>
