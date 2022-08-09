import { defineStore } from "pinia"
import { Products } from "@/types"
import { categories } from "@/definitions"

export const useGlobalStore = defineStore("global", {
    state: () => ({
        loaded: false,
        allProducts: <Products>[],
        selectedCategory: categories[0].id,
    }),
})
