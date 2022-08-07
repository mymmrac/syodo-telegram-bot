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
  <div v-else>
    <category-list :categories="categories"></category-list>
    <hr class="border-tg-hint">
    <product-list :products="products"></product-list>
  </div>
</template>

<script setup lang="ts">
import { computed, ComputedRef, Ref, ref, watch } from "vue"
import { Products } from "./types"
import syodoAPI from "./syodo"
import { TelegramWebApps } from "telegram-bots-webapps-types"
import ProductList from "@/components/ProductList.vue"
import CategoryList from "@/components/CategoryList.vue"
import { categories } from "@/definitions"

const tg: TelegramWebApps.WebApp = window.Telegram.WebApp

// Errors
const errors: Ref<any[]> = ref([])

function sendError(type: string, data: any) {
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
</script>
