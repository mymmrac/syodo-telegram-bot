import { defineStore } from "pinia"
import { Products } from "@/types"

export const useGlobalStore = defineStore("global", {
    state: () => ({
        loaded: false,
        allProducts: <Products>[],
    }),
})
