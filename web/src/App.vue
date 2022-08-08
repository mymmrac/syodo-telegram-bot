<template>
  <transition name="m-fade" mode="out-in">
    <div v-show="!checkout">
      <category-list :categories="categories" @categorySelected="categorySelected"></category-list>
      <hr class="border-tg-hint">
      <product-list :products="products" :category="selectedCategory" @productUpdate="updateOrder"></product-list>

      <button :class="scrollPos < 256 ? 'hidden' : ''" @click="scrollToTop('smooth')"
              class="fixed bottom-8 right-2 text-tg-link rounded-full bg-tg-bg">
        <svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" fill="currentColor" viewBox="0 0 16 16">
          <path
              d="M16 8A8 8 0 1 0 0 8a8 8 0 0 0 16 0zm-7.5 3.5a.5.5 0 0 1-1 0V5.707L5.354 7.854a.5.5 0 1 1-.708-.708l3-3a.5.5 0 0 1 .708 0l3 3a.5.5 0 0 1-.708.708L8.5 5.707V11.5z"/>
        </svg>
      </button>
    </div>
  </transition>
  <transition name="m-fade" mode="in-out">
    <div v-show="checkout">
      <p>Total Price: {{ totalPrice }}</p>
      <pre>{{ order }}</pre>
    </div>
  </transition>
</template>

<script setup lang="ts">
import { computed, ComputedRef, onMounted, Ref, ref, watch } from "vue"
import { Order, OrderProduct, priceToText, Products } from "./types"
import syodoAPI from "./syodo"
import { TelegramWebApps } from "telegram-bots-webapps-types"
import ProductList from "@/components/ProductList.vue"
import CategoryList from "@/components/CategoryList.vue"
import { categories } from "@/definitions"

const tg: TelegramWebApps.WebApp = window.Telegram.WebApp

// Errors
function sendError(type: string, data: any) {
  console.error(`Error type:${ type }, data: ${ data }`)
  tg.HapticFeedback.notificationOccurred("error")
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

const scrollPos: Ref<number> = ref(window.scrollY)

onMounted(() => {
  window.addEventListener("scroll", () => {
    scrollPos.value = window.scrollY
  })
})

// Products
const allProducts: Ref<Products> = ref([])

const selectedCategory: Ref<string> = ref(categories[0].id)

function categorySelected(id: string) {
  selectedCategory.value = id
}

const products: ComputedRef<Products> = computed(() => {
  return allProducts.value.filter(p => p.category_id !== "14" && p.showOnMain)
})

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

function scrollToTop(behavior: ScrollBehavior = "auto") {
  window.scrollTo({ top: 0, behavior: behavior })
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
