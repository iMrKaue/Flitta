const API_URL = "http://localhost:8080";

function getToken() {
    return localStorage.getItem("token");
}

async function apiFetch(path, options = {}) {
    const token = getToken();

    const headers = {
        "Content-Type": "application/json",
        ...(token ? { "Authorization": `Bearer ${token}` } : {})
    };

    const res = await fetch(`${API_URL}${path}`, {
        ...options,
        headers
    });

    if (res.status === 401) {
        localStorage.removeItem("token");
        window.location.href = "login.html";
        return null;
    }

    const text = await res.text();

    if (!res.ok) {
        throw new Error(text || "Erro na requisição");
    }

    if (!text) {
        return null;
    }

    try {
        return JSON.parse(text);
    } catch (err) {
        return text;
    }
}