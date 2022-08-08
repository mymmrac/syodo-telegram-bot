<template>
  <!-- Telegram Colors Demo -->
  <div v-if="false">
    <div class="bg-tg-bg">tg-bg</div>
    <div class="bg-tg-text text-tg-bg">tg-text</div>
    <div class="bg-tg-hint">tg-hint</div>
    <div class="bg-tg-link">tg-link</div>
    <div class="bg-tg-button">tg-button</div>
    <div class="bg-tg-button-text">tg-button-text</div>
    <div class="bg-tg-secondary-bg">tg-secondary-bg</div>
  </div>

  <div v-if="errors.length > 0" class="text-red-500">
    Виникла помилка: {{ errors[0] }}
  </div>
  <template v-else>
    <transition name="m-fade" mode="out-in">
      <div v-show="!checkout">
        <category-list :categories="categories"></category-list>
        <hr class="border-tg-hint">
        <hr class="border-tg-hint">
        <product-list :products="products" @productUpdate="updateOrder"></product-list>
      </div>
    </transition>
    <transition name="m-fade" mode="out-in">
      <div v-show="checkout">
        <p>Total Price: {{ totalPrice }}</p>
        <pre>{{ order }}</pre>
      </div>
    </transition>
  </template>
</template>

<script setup lang="ts">
import { computed, ComputedRef, Ref, ref, watch } from "vue"
import { Order, OrderProduct, priceToText, Products } from "./types"
import syodoAPI from "./syodo"
import { TelegramWebApps } from "telegram-bots-webapps-types"
import ProductList from "@/components/ProductList.vue"
import CategoryList from "@/components/CategoryList.vue"
import { categories } from "@/definitions"

const tg: TelegramWebApps.WebApp = window.Telegram.WebApp

// Errors
const errors: Ref<any[]> = ref([])

function sendError(type: string, data: any) {
  tg.HapticFeedback.notificationOccurred("error")
  console.error(`Error type:${ type }, data: ${ data }`)
  tg.sendData(`${ type }:${ data }`)
}

// Version check
const [ major, minor ] = tg.version.split(".").map(Number)
if (major < 6 || (major == 6 && minor < 1)) {
  sendError("old-version", tg.version)
}

// Loaders
const loaded: Ref<boolean> = ref(false)

watch(loaded, (isLoaded) => {
  if (!isLoaded) {
    return
  }

  console.log("Loaded")
  tg.ready()
})

// Products
const allProducts: Ref<Products> = ref([])

const products: ComputedRef<Products> = computed(() => {
  return allProducts.value.filter(p => p.category_id !== "14")
})

syodoAPI.get<Products>("/products")
    .then(response => {
      if (response.status !== 200) {
        console.error(response)
        errors.value.push("Хмм, не вдалося завантажити меню")
        return
      }
      allProducts.value = response.data
    })
    .catch(err => {
      console.error(err)
      errors.value.push("Хмм, не вдалося завантажити меню")
    })
    .finally(() => loaded.value = true)

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

const prevScroll: Ref<number> = ref(0)

tg.MainButton.onClick(() => {
  if (!checkout.value) {
    prevScroll.value = window.scrollY

    checkout.value = true
    window.scrollTo({ // TODO: Fix scrolling
      top: 0,
      behavior: "auto",
    })

    tg.BackButton.show()
    tg.MainButton.setText(`Замовити - ${ priceToText(totalPrice.value) }`)
  } else {
    alert("OK")
  }
})

tg.BackButton.onClick(() => {
  checkout.value = false
  window.scrollTo({
    top: prevScroll.value,
    behavior: "auto",
  })

  tg.BackButton.hide()
  tg.MainButton.setText("Переглянути замовлення")
})

const checkout: Ref<boolean> = ref(false)

const totalPrice: ComputedRef<number> = computed(() => {
  let price = 0
  order.value.products.forEach(p => price += Number(p.product.price))
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
