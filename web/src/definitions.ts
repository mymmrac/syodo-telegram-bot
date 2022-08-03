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

export type SubCategory = {
    id: string
    title: string
}

export type SubCategories = SubCategory[]

export type Setting = {
    id: string
    values: SubCategories
}

export type GeneralSettings = Setting[]
