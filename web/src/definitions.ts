import { Category, SubCategory } from "@/types"
import { href } from "@/utils"

export const categories: readonly Category[] = [
    { id: "13", title: "Суші", icon: href("img/sushi.png") },
    { id: "7", title: "Роли", icon: href("img/roles.png") },
    { id: "8", title: "Сети", icon: href("img/sets.png") },
    { id: "9", title: "Напої", icon: href("img/drinks.png") },
    { id: "10", title: "Соуси", icon: href("img/sauces.png") },
    { id: "11", title: "Десерти", icon: href("img/desserts.png") },
]

export const hasSubCategoriesCategory = "7"
export const noLactoseCategory = "14"

export const subCategories: readonly SubCategory[] = [
    { id: 1, title: "Класичні" },
    { id: 2, title: "Фелікси" },
    { id: 3, title: "Макі" },
    { id: 4, title: "Гарячі роли" },
    { id: 5, title: "Авторські" },
]
