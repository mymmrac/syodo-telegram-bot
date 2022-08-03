<template>
  <div v-if="errors.length > 0" class="text-red-500">
    Виникла помилка: {{ errors[0] }}
  </div>
  <div v-else>
    <div v-for="product in products" :key="product.id">
      {{ product.title }} - {{ product.price }}
      <img :src="product.image_original || product.image" :alt="product.title" width="100">
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ComputedRef, Ref, ref, watch } from "vue"
import { GeneralSettings, Products, SubCategories } from "./definitions"
import syodoAPI from "./syodo"
import { TelegramWebApps } from "telegram-bots-webapps-types"

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
const subCategoriesLoaded: Ref<boolean> = ref(false)
const productsLoaded: Ref<boolean> = ref(false)

const loaded: ComputedRef<boolean> = computed(() => subCategoriesLoaded.value && productsLoaded.value)

watch(loaded, (isLoaded) => {
  if (!isLoaded) {
    return
  }

  console.log("Loaded")
  tg.ready()
})

// SubCategories & Products
const subCategories: Ref<SubCategories> = ref([])
const allProducts: Ref<Products> = ref([])

const products: ComputedRef<Products> = computed(() => {
  return allProducts.value.filter(p => p.category_id !== "14")
})

syodoAPI.get<GeneralSettings>("/generalsettings/subcategories")
    .then(response => {
      if (response.status !== 200) {
        console.error(response)
        errors.value.push("Хмм, не вдалося завантажити категорії")
        return
      }
      subCategories.value = response.data[0].values
    })
    .catch(err => {
      console.error(err)
      errors.value.push("Хмм, не вдалося завантажити категорії")
    })
    .finally(() => subCategoriesLoaded.value = true)

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
    .finally(() => productsLoaded.value = true)
</script>
