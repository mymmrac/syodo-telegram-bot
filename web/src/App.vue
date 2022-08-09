<template>
  <transition name="m-fade" mode="out-in">
    <div v-show="!checkout">
      <category-list :categories="categories" :selected-category="selectedCategory"
                     @categorySelected="categorySelected"></category-list>
      <div class="w-full px-2 pb-2 flex gap-2">
        <input type="text" placeholder="Пошук..." v-model.trim="search"
               class="p-2 flex-1 rounded border-none ring-0 focus:ring-0 bg-tg-button text-tg-button-text placeholder-tg-button-text">
        <button class="rounded px-2" @click="search = ''">
          <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" fill="currentColor"
               class="bi bi-backspace-fill" viewBox="0 0 16 16">
            <path
                d="M15.683 3a2 2 0 0 0-2-2h-7.08a2 2 0 0 0-1.519.698L.241 7.35a1 1 0 0 0 0 1.302l4.843 5.65A2 2 0 0 0 6.603 15h7.08a2 2 0 0 0 2-2V3zM5.829 5.854a.5.5 0 1 1 .707-.708l2.147 2.147 2.146-2.147a.5.5 0 1 1 .707.708L9.39 8l2.146 2.146a.5.5 0 0 1-.707.708L8.683 8.707l-2.147 2.147a.5.5 0 0 1-.707-.708L7.976 8 5.829 5.854z"/>
          </svg>
        </button>
      </div>
      <hr>
      <product-list :products="products" :all-products="allProducts" :category="selectedCategory" :search="search"
                    @productUpdate="updateOrder"></product-list>

      <button :class="scrollPos < 256 ? 'hidden' : ''" @click="scrollToTop('smooth')"
              class="fixed z-20 bottom-8 right-2 text-tg-button rounded-full bg-tg-button-text shadow-xl">
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
import syodoAPI from "./syodo-api"
import { TelegramWebApps } from "telegram-bots-webapps-types"
import ProductList from "@/components/ProductList.vue"
import CategoryList from "@/components/CategoryList.vue"
import { categories, subCategories } from "@/definitions"

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
  products.value.forEach(p => console.log(p.subcategory || "---"))

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

const search: Ref<string> = ref("")

function categorySelected(id: string) {
  selectedCategory.value = id
}

const products: ComputedRef<Products> = computed(() => {
  return allProducts.value
      .filter(p => p.category_id !== "14" && !p.hidePosition)
      .sort((p1, p2) => {
        if (p1.subcategory && p2.subcategory) {
          const s1 = subCategories.find(s => s.title === p1.subcategory)
          const s2 = subCategories.find(s => s.title === p2.subcategory)

          if (s1 && s2) {
            return s1.id - s2.id
          }
        } else if (p1.subcategory) {
          return 1
        } else if (p2.subcategory) {
          return -1
        }

        return 0
      })
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
