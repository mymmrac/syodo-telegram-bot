<template>
  <transition-group tag="div" name="m-buttons-fade" class="relative h-8">
    <button v-if="amount === 0" :class="fixedSize ? 'w-24' : 'w-full'" class="m-btn-big" @click="add">Додати</button>
    <div v-else class="flex justify-around items-center shadow rounded-lg bg-tg-button text-tg-button-text text-lg">
      <button :class="fixedSize ? 'w-8' : 'flex-1'" class="m-btn" @click="removeInternal">-</button>
      <transition tag="template" name="m-text-fade" mode="out-in">
        <div :key="amount" class="text-center w-8 transition duration-200">{{ amount }}</div>
      </transition>
      <button :class="fixedSize ? 'w-8' : 'flex-1'" class="m-btn" @click="addInternal">+</button>
    </div>
  </transition-group>
</template>

<script setup lang="ts">
import { TelegramWebApps } from "telegram-bots-webapps-types"

const props = defineProps<{
  amount: number
  add: (payload: MouseEvent) => void
  remove: (payload: MouseEvent) => void
  fixedSize?: boolean
}>()

const tg: TelegramWebApps.WebApp = window.Telegram.WebApp

function addInternal(payload: MouseEvent): void {
  tg.HapticFeedback.selectionChanged()
  props.add(payload)
}

function removeInternal(payload: MouseEvent): void {
  tg.HapticFeedback.selectionChanged()
  props.remove(payload)
}
</script>

<style scoped lang="scss">
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