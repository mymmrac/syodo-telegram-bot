import axios from "axios"

const syodoAPI = axios.create({
    baseURL: "https://e0uf7jciif.execute-api.eu-central-1.amazonaws.com/production",
    withCredentials: false,
    headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
    },
})

export default syodoAPI
