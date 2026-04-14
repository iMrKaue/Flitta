const API_URL = "http://localhost:8080";

function getToken() {
    return localStorage.getItem("token")
}

async function apiFetch(path, options = {}) {
    const token = getToken();

    const headers = {
        "Content-Type": "application/json",
        "Authorization": `Bearer ${token}`
    };

    const res = await fetch(`${API_URL}${path}`, {
        ...options,
        headers
    });

    if (res.status === 401) {
        //🔒 token inválido → volta pro 
        window.location.href = "login.html";
        return;
    }

    return res.json();
}