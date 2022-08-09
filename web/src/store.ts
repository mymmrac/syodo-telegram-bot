import { defineStore } from "pinia"

import { isProduct, ProductListItems, Products } from "@/types"
import { categories, noLactoseCategory, subCategories } from "@/definitions"
import { insert } from "@/utils"

export const useGlobalStore = defineStore("global", {
    state: () => ({
        loaded: false,

        allProducts: <Products>[],

        selectedCategory: categories[0].id,
        search: "",
    }),

    getters: {
        isSearchEmpty: (state): boolean => state.search === "",

        items: (state): ProductListItems => {
            let items: ProductListItems = state.allProducts
                .filter(p => p.category_id !== noLactoseCategory && !p.hidePosition)
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

            subCategories.forEach(s => {
                const i = items.findIndex(o => {
                    if (!isProduct(o)) {
                        return false
                    }
                    return o.subcategory === s.title
                })
                if (i < 0) {
                    return
                }

                items = insert(items, i, s)
            })

            return items
        },
    },

    actions: {
        clearSearch(): void {
            this.search = ""
        },
    },
})
