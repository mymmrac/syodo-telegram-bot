import axios from "axios"

const syodoAPI = axios.create({
    baseURL: __SYODO_API__,
    withCredentials: false,
    headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
        "x-api-key": __SYODO_API_TOKEN__,
    },
})

export default syodoAPI
