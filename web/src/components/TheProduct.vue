<template>
  <div class="rounded p-2 border border-tg-hint flex flex-col justify-between">
    <div class="aspect-square rounded bg-white grid place-content-center">
      <img :src="getImage(product)" :alt="product.title" class="rounded">
    </div>
    <div>
      <p>{{ product.title }}</p>
      <div class="flex justify-between">
        <p>{{ product.weight }}</p>
        <p>{{ getPrice(product) }}</p>
      </div>
      <transition-group tag="div" name="m-buttons-fade" class="mt-2 relative">
        <button v-if="amount === 0" class="w-full m-btn" @click="add">Додати</button>
        <div v-else class="flex justify-around items-center">
          <button class="w-full m-btn" @click="remove">-</button>
          <transition tag="template" name="m-text-fade" mode="out-in">
            <p :key="amount" class="px-3 transition duration-200">{{ amount }}</p>
          </transition>
          <button class="w-full m-btn" @click="add">+</button>
        </div>
      </transition-group>
    </div>
  </div>
</template>

<script setup lang="ts">
import { getImage, getPrice, OrderProduct, Product } from "@/types"
import { Ref, ref } from "vue"
import { TelegramWebApps } from "telegram-bots-webapps-types"

const props = defineProps<{
  product: Product
}>()

const emit = defineEmits<{
  (e: "productUpdate", product: OrderProduct): void
}>()

const tg: TelegramWebApps.WebApp = window.Telegram.WebApp

const amount: Ref<number> = ref(0)

function update() {
  tg.HapticFeedback.selectionChanged()
  emit("productUpdate", {
    id: props.product.id,
    amount: amount.value,
    product: props.product,
  })
}

function add() {
  amount.value++
  update()
}

function remove() {
  amount.value--
  update()
}
</script>

<style scoped lang="scss">
.m-btn {
  @apply py-1 px-2 rounded;
}

.m-buttons-fade {
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

.m-text-fade {
  &-enter-active,
  &-leave-active {
    transition: all 0.12s ease;
  }

  &-enter-from,
  &-leave-to {
    opacity: 20;
    transform: scale(0.8);
  }
}
</style>
