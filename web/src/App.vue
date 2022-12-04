<template>
  <transition name="m-fade" mode="out-in">
    <div v-show="!checkout">
      <category-list/>
      <div class="w-full px-2 pb-2 my-1">
        <div class="rounded-lg shadow-lg flex gap-2">
          <input type="text" placeholder="Пошук..." @input="updateSearch" :value="search"
                 class="p-2 flex-1 rounded-lg border-none ring-0 focus:ring-0 bg-tg-bg text-tg-text placeholder-tg-text text-sm">
          <button class="rounded-lg px-2" @click="store.clearSearch">
            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" fill="currentColor" class="bi bi-x"
                 viewBox="0 0 16 16">
              <path
                  d="M4.646 4.646a.5.5 0 0 1 .708 0L8 7.293l2.646-2.647a.5.5 0 0 1 .708.708L8.707 8l2.647 2.646a.5.5 0 0 1-.708.708L8 8.707l-2.646 2.647a.5.5 0 0 1-.708-.708L7.293 8 4.646 5.354a.5.5 0 0 1 0-.708z"/>
            </svg>
          </button>
        </div>
      </div>
      <product-list/>
    </div>
  </transition>

  <transition name="m-fade" mode="in-out">
    <checkout v-show="checkout" :order="order"></checkout>
  </transition>

  <div class="h-[96px]"></div>
  <go-to-top-button/>

  <div v-if="showOutOfDate" class="z-50 fixed top-0 bottom-0 left-0 right-0 overflow-y-scroll bg-gray-500/75 p-8"
       @click="showOutOfDate = false">
    <div class="bg-tg-bg rounded shadow p-2 m-card" @click.stop>
      <div class="text-center py-16">
        <p class="text-xl mb-8">Вибачте, але ми зараз не працюємо.</p>
        <p>Ми з радістю приготуємо Ваше замовлення щодня з 10:00 по 21:45</p>
      </div>
      <button class="w-full m-btn-big mt-2" @click="showOutOfDate = false">Закрити</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import GoToTopButton from "@/components/GoToTopButton.vue"
import CategoryList from "@/components/CategoryList.vue"
import ProductList from "@/components/ProductList.vue"
import Checkout from "@/components/Checkout.vue"

import { TelegramWebApps } from "telegram-bots-webapps-types"
import { onMounted, Ref, ref, watch } from "vue"
import { storeToRefs } from "pinia"

import { scrollToTop, showError, tgVersionSupported } from "@/utils"
import { priceToText, Products } from "@/types"
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
tg.MainButton.setParams({ color: "#bb4347", text_color: "#ffffff" })

const store = useGlobalStore()
const { loaded, allProducts, search, order, outOfTime } = storeToRefs(store)

// Loaders
watch(loaded, (isLoaded) => {
  if (!isLoaded) {
    return
  }

  console.log("Loaded")
  tg.ready()
})

const showOutOfDate: Ref<boolean> = ref(false)

function checkOutOfTime() {
  let dateNow = new Date()
  if (dateNow.getHours() < 10 ||
      (dateNow.getHours() >= 22) || (dateNow.getHours() == 21 && dateNow.getMinutes() >= 45)) {
    outOfTime.value = true
  }
}

checkOutOfTime()
if (outOfTime.value) {
  showOutOfDate.value = true
}

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

    checkOutOfTime()

    if (outOfTime.value) {
      tg.MainButton.disable()
      tg.MainButton.setText("На жали ми зараз не працюємо")
    } else {
      tg.MainButton.enable()
      tg.MainButton.setText(`Замовити - ${ priceToText(store.totalOrderPrice) }`)
    }

    tg.enableClosingConfirmation()
    return
  }

  if (!store.isOrderEmpty) {
    tg.MainButton.enable()
    tg.MainButton.setText("Переглянути замовлення")
    tg.MainButton.show()
    tg.enableClosingConfirmation()
  } else {
    tg.MainButton.hide()
    tg.disableClosingConfirmation()
  }
}, { deep: true })

tg.MainButton.onClick(() => {
  checkOutOfTime()

  if (!checkout.value) {
    checkout.value = true
    scrollToTop()

    tg.BackButton.show()
    if (outOfTime.value) {
      tg.MainButton.disable()
      tg.MainButton.setText("На жали ми зараз не працюємо")
    } else {
      tg.MainButton.enable()
      tg.MainButton.setText(`Замовити - ${ priceToText(store.totalOrderPrice) }`)
    }
  } else {
    if (!outOfTime.value) {
      confirmOrder()
    }
  }
})

tg.BackButton.onClick(() => {
  checkout.value = false
  scrollToTop()

  tg.BackButton.hide()
  tg.MainButton.enable()
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
    comment: order.value.comment,
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
      tg.HapticFeedback.notificationOccurred("success")
      tg.close()
      return
    case "cancelled":
      tg.HapticFeedback.notificationOccurred("warning")
      return
    case "failed":
      tg.HapticFeedback.notificationOccurred("error")
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
