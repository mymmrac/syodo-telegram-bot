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
          <div v-if="linkedProduct && linkedProduct.category_id === noLactoseCategory" class="mt-2">
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
import AddRemoveButtons from "@/components/AddRemoveButtons.vue"

import { computed, ComputedRef, Ref, ref, watch } from "vue"
import { storeToRefs } from "pinia"

import { getImage, getPrice, Product } from "@/types"
import { noLactoseCategory } from "@/definitions"
import { useGlobalStore } from "@/store"

const props = defineProps<{
  product: Product
  linkedProduct: Product | undefined
}>()

const store = useGlobalStore()
const { order } = storeToRefs(store)

const useLinkedProduct: Ref<boolean> = ref(store.isUsedLinkedInOrder(props.product))
const usedProduct: ComputedRef<Product> = computed(() => {
  if (!props.linkedProduct) {
    return props.product
  }
  return useLinkedProduct.value ? props.linkedProduct : props.product
})

const amount: Ref<number> = ref(store.amountInOrder(usedProduct.value.id))
watch(order, () => {
  amount.value = store.amountInOrder(usedProduct.value.id)
}, { deep: true })

const showDetails: Ref<boolean> = ref(false)

function update() {
  if (!props.linkedProduct) {
    store.updateInOrder({
      amount: amount.value,
      product: props.product,
    })
    return
  }

  if (useLinkedProduct.value) {
    store.updateInOrder({
      amount: amount.value,
      product: props.linkedProduct,
    })
    store.removeFromOrder(props.product.id)
  } else {
    store.updateInOrder({
      amount: amount.value,
      product: props.product,
    })
    store.removeFromOrder(props.linkedProduct.id)
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
