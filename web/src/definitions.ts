import { Categories, SubCategories } from "@/types"

const categories: Categories = [
    { id: "13", title: "Суші", icon: "/img/sushi.png" },
    { id: "7", title: "Роли", icon: "/img/roles.png" },
    { id: "8", title: "Сети", icon: "/img/sets.png" },
    { id: "9", title: "Напої", icon: "/img/drinks.png" },
    { id: "10", title: "Соуси", icon: "/img/sauces.png" },
    { id: "11", title: "Десерти", icon: "/img/desserts.png" },
]

const subCategories: SubCategories = [
    { id: 1, title: "Класичні" },
    { id: 2, title: "Фелікси" },
    { id: 3, title: "Макі" },
    { id: 4, title: "Гарячі роли" },
    { id: 5, title: "Авторські" },
]

export {
    categories,
    subCategories,
}
