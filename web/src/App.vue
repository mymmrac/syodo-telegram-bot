<template>
  <transition name="m-fade" mode="out-in">
    <div v-show="!checkout">
      <category-list/>
      <div class="w-full px-2 pb-2 flex gap-2">
        <input type="text" placeholder="Пошук..." @input="updateSearch" :value="search"
               class="p-2 flex-1 rounded border-none ring-0 focus:ring-0 bg-tg-button text-tg-button-text placeholder-tg-button-text shadow">
        <button class="rounded px-2 shadow" @click="store.clearSearch">
          <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" fill="currentColor"
               class="bi bi-backspace-fill" viewBox="0 0 16 16">
            <path
                d="M15.683 3a2 2 0 0 0-2-2h-7.08a2 2 0 0 0-1.519.698L.241 7.35a1 1 0 0 0 0 1.302l4.843 5.65A2 2 0 0 0 6.603 15h7.08a2 2 0 0 0 2-2V3zM5.829 5.854a.5.5 0 1 1 .707-.708l2.147 2.147 2.146-2.147a.5.5 0 1 1 .707.708L9.39 8l2.146 2.146a.5.5 0 0 1-.707.708L8.683 8.707l-2.147 2.147a.5.5 0 0 1-.707-.708L7.976 8 5.829 5.854z"/>
          </svg>
        </button>
      </div>
      <hr>
      <product-list/>
    </div>
  </transition>

  <transition name="m-fade" mode="in-out">
    <checkout v-show="checkout" :order="order"></checkout>
  </transition>

  <div class="h-[96px]"></div>
  <go-to-top-button/>
</template>

<script setup lang="ts">
import ProductList from "@/components/ProductList.vue"
import CategoryList from "@/components/CategoryList.vue"
import Checkout from "@/components/Checkout.vue"
import GoToTopButton from "@/components/GoToTopButton.vue"

import { TelegramWebApps } from "telegram-bots-webapps-types"
import { onMounted, Ref, ref, watch } from "vue"
import { storeToRefs } from "pinia"

import { priceToText, Products } from "@/types"
import { scrollToTop, showError, tgVersionSupported } from "@/utils"
import { useGlobalStore } from "@/store"
import syodoAPI from "@/syodo-api"
import botAPI from "@/bot-api"

const tg: TelegramWebApps.WebApp = window.Telegram.WebApp

// Version check
if (!tgVersionSupported("6.1")) {
  showError("old-version", `Застаріла версія Telegram (v${ tg.version }), очікується щонайменше v6.1`)
  tg.close()
}

// Color scheme
function updateColorScheme() {
  document.documentElement.className = tg.colorScheme
}

tg.onEvent("themeChanged", updateColorScheme)
onMounted(() => {
  updateColorScheme()
})

const store = useGlobalStore()
const { loaded, allProducts, search, order } = storeToRefs(store)

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
        showError("load-products", "Хмм, не вдалося завантажити меню", response.statusText)
        return
      }
      allProducts.value = response.data
    })
    .catch(err => {
      console.error(err)
      showError("load-products", "Хмм, не вдалося завантажити меню", err)
    })
    .finally(() => loaded.value = true)

function updateSearch(e: Event) {
  const target = e.target as HTMLInputElement
  search.value = target.value.trim()
}

// Order
watch(order, () => {
  if (checkout.value) {
    if (store.isOrderEmpty) {
      checkout.value = false
      tg.MainButton.hide()
      tg.BackButton.hide()
      tg.disableClosingConfirmation()
      return
    }

    tg.MainButton.setText(`Замовити - ${ priceToText(store.totalOrderPrice) }`)
    tg.enableClosingConfirmation()
    return
  }

  if (!store.isOrderEmpty) {
    tg.MainButton.setText("Переглянути замовлення")
    tg.MainButton.show()
    tg.enableClosingConfirmation()
  } else {
    tg.MainButton.hide()
    tg.disableClosingConfirmation()
  }
}, { deep: true })

tg.MainButton.onClick(() => {
  if (!checkout.value) {
    checkout.value = true
    scrollToTop()

    tg.BackButton.show()
    tg.MainButton.setText(`Замовити - ${ priceToText(store.totalOrderPrice) }`)
  } else {
    confirmOrder()
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

// Order processing
function confirmOrder() {
  tg.MainButton.showProgress(false)

  if (order.value.products.size === 0) {
    tg.MainButton.hideProgress()
    showError("empty-order", "Хмм, у Вас пуста корзина")
    return
  }

  if (!tg.initDataUnsafe.hash || !tg.initDataUnsafe.user?.id || !tg.initDataUnsafe.query_id) {
    tg.MainButton.hideProgress()
    showError("empty-hash", "Хмм, щось не так з Вашими даними")
    return
  }

  const finalOrder: {
    appData: string
    products: {
      id: string
      title: string
      price: number
      amount: number
      categoryID: string
    }[],
    doNotCall: boolean
    noNapkins: boolean
    cutleryCount: number
    trainingCutleryCount: number
    comment: string
  } = {
    appData: tg.initData,
    products: Array.from(order.value.products.values()).map(op => {
      return {
        id: op.product.id,
        title: op.product.title,
        price: Number(op.product.price),
        amount: op.amount,
        categoryID: op.product.category_id,
      }
    }),
    doNotCall: order.value.doNotCall,
    noNapkins: order.value.noNapkins,
    cutleryCount: order.value.cutleryCount,
    trainingCutleryCount: order.value.trainingCutleryCount,
    comment: order.value.addComment ? order.value.comment : "",
  }

  botAPI.post("/order", finalOrder)
      .then(response => {
        if (response.status !== 200) {
          showError("order-status", "Хмм, не вдалося щось не так з замовленням", response.statusText)
          return
        }

        const invoiceURL: string = response.data
        tg.openInvoice(invoiceURL, invoiceResult)
      })
      .catch(err => {
        showError("order", "Хмм, не вдалося опрацювати замовлення", err)
      })
      .finally(() => {
        tg.MainButton.hideProgress()
      })
}

function invoiceResult(result: string) {
  switch (result) {
    case "paid":
      tg.close()
      return
    case "failed":
      return
  }

  showError("invoice", "Хмм, щось не так з оплатою", `Статус: ${ result }`)
}

tg.onEvent("invoiceClosed", () => {
  tg.MainButton.hideProgress()
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
