import axios from "axios"

const botAPI = axios.create({
    baseURL: __BOT_API__,
    withCredentials: false,
    headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
    },
})

export default botAPI
