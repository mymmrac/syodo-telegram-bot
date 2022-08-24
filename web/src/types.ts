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
    showOnMain: boolean
    description: string
    linkedPosition: string
    hidePosition: boolean
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
    amount: number
    product: Product
}

export type Order = {
    products: Map<string, OrderProduct>
    doNotCall: boolean
    noNapkins: boolean
    cutleryCount: number
    trainingCutleryCount: number
    addComment: boolean
    comment: string
}

export type ProductListItem = Product | SubCategory
export type ProductListItems = ProductListItem[]

export function isProduct(item: ProductListItem): item is Product {
    return (<Product>item).price !== undefined
}
