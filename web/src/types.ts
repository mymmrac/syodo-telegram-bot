export type Product = {
    category_id: string
    category_name: string
    weight: string
    subcategory: string
    image: string
    image_original: string
    mod_date: string
    id: string
    price: string
    title: string
}

export type Products = Product[]

export function getImage(product: Product): string {
    return product.image || product.image_original
}

export function priceToText(price: number): string {
    return (price / 100).toFixed(2) + "грн"
}

export function getPrice(product: Product): string {
    return priceToText(Number(product.price))
}

export type Category = {
    id: string
    title: string
    icon: string
}

export type Categories = Category[]

export type SubCategory = {
    id: number
    title: string
}

export type SubCategories = SubCategory[]

export type OrderProduct = {
    id: string
    amount: number
    product: Product
}

export type Order = {
    products: Map<string, OrderProduct>
}
