<template>
  <div class="rounded p-2 border border-tg-hint flex flex-col justify-between">
    <div class="aspect-square rounded bg-white grid place-content-center cursor-pointer" @click="showDetails = true">
      <img :src="getImage(product)" :alt="usedProduct.title" class="rounded">
    </div>
    <div>
      <p>{{ usedProduct.title }}</p>
      <div class="flex justify-between">
        <p>{{ product.weight }}</p>
        <p>{{ getPrice(usedProduct) }}</p>
      </div>
      <add-remove-buttons :amount="amount" :add="add" :remove="remove"/>
    </div>
    <transition name="m-card-fade">
      <div v-if="showDetails" class="z-50 fixed top-0 bottom-0 left-0 right-0 overflow-y-scroll bg-gray-500/75 p-8"
           @click="showDetails = false">
        <div class="bg-tg-bg rounded p-2 m-card" @click.stop>
          <div class="aspect-square rounded bg-white grid place-content-center">
            <img :src="getImage(product)" :alt="usedProduct.title" class="rounded">
          </div>
          <p class="text-xl">{{ usedProduct.title }}</p>
          <div class="flex justify-between">
            <p>{{ product.weight }}</p>
            <p>{{ getPrice(usedProduct) }}</p>
          </div>
          <hr>
          <p class="">{{ product.description }}</p>
          <div v-if="linkedProduct && linkedProduct.category_id === '14'" class="mt-2">
            <button class="w-full m-btn" :class="useLinkedProduct ? '' : 'bg-tg-hint'"
                    @click="useLinkedProduct = !useLinkedProduct">
              Без лактози
            </button>
          </div>
          <add-remove-buttons :amount="amount" :add="add" :remove="remove"/>
          <button class="w-full m-btn mt-2" @click="showDetails = false">Закрити</button>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { getImage, getPrice, OrderProduct, Product } from "@/types"
import { computed, ComputedRef, Ref, ref, watch } from "vue"
import { TelegramWebApps } from "telegram-bots-webapps-types"
import AddRemoveButtons from "@/components/AddRemoveButtons.vue"

const props = defineProps<{
  product: Product
  linkedProduct: Product | undefined
}>()

const emit = defineEmits<{
  (e: "productUpdate", product: OrderProduct): void
}>()

const tg: TelegramWebApps.WebApp = window.Telegram.WebApp

const amount: Ref<number> = ref(0)
const showDetails: Ref<boolean> = ref(false)
const useLinkedProduct: Ref<boolean> = ref(false)

const usedProduct: ComputedRef<Product> = computed(() => {
  if (!props.linkedProduct) {
    return props.product
  }
  return useLinkedProduct.value ? props.linkedProduct : props.product
})

function update() {
  tg.HapticFeedback.selectionChanged()

  if (!props.linkedProduct) {
    emit("productUpdate", {
      id: props.product.id,
      amount: amount.value,
      product: props.product,
    })
    return
  }

  if (useLinkedProduct.value) {
    emit("productUpdate", {
      id: props.linkedProduct.id,
      amount: amount.value,
      product: props.linkedProduct,
    })
    emit("productUpdate", {
      id: props.product.id,
      amount: 0,
      product: props.product,
    })
  } else {
    emit("productUpdate", {
      id: props.product.id,
      amount: amount.value,
      product: props.product,
    })
    emit("productUpdate", {
      id: props.linkedProduct.id,
      amount: 0,
      product: props.linkedProduct,
    })
  }
}

watch(useLinkedProduct, () => {
  update()
})

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

.m-card-fade {
  &-enter-active,
  &-leave-active {
    transition: all 0.1s ease;

    .m-card {
      transition: all 0.2s ease;
    }
  }

  &-enter-active .m-card {
    transition-delay: 0.1s;
  }

  &-leave-active {
    transition-delay: 0.2s;
  }

  &-enter-from,
  &-leave-to {
    opacity: 0;

    .m-card {
      opacity: 0;
      transform: scale(0.9);
    }
  }
}
</style>
