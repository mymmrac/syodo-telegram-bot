<template>
  <div class="text-2xl text-tg-text bg-tg-bg">
    Hello World!
  </div>
  <br>
  <div class="bg-tg-bg">tg-bg</div>
  <div class="bg-tg-text">tg-text</div>
  <div class="bg-tg-hint">tg-hint</div>
  <div class="bg-tg-link">tg-link</div>
  <div class="bg-tg-button">tg-button</div>
  <div class="bg-tg-button-text">tg-button-text</div>
  <div class="bg-tg-secondary-bg">tg-secondary-bg</div>

  <button class="m-2 p-2 bg-tg-button text-tg-text rounded">Test {{ count }}</button>

  <br>
  <hr>
  <pre class="bg-tg-bg text-tg-text">{{ tgJson }}</pre>
  <hr>
  <br>
</template>

<script setup lang="ts">
import { ref } from "vue"

const tg = window.Telegram.WebApp
const tgJson = ref(JSON.stringify(tg, null, 4))

function updateTg() {
  tgJson.value = JSON.stringify(window.Telegram.WebApp, null, 4)
}

tg.onEvent("themeChanged", updateTg)
tg.onEvent("viewportChanged", updateTg)

tg.MainButton.setText("+")
tg.MainButton.show()

const count = ref(0)
tg.onEvent("mainButtonClicked", () => {
  count.value++

  if (count.value > 3) {
    tg.MainButton.showProgress(false)
  }
})
</script>

<style scoped lang="scss"></style>
