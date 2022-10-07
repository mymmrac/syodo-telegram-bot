import axios from "axios"

const syodoAPI = axios.create({
    baseURL: __SYODO_API__,
    withCredentials: false,
    headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
        "x-api-key": "yjhlMaWbxb412floOKrhfaJWiAO9OFh21RTq9X9o",
    },
})

export default syodoAPI
