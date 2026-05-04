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

    const contentType = res.headers.get("content-type");

    if (!res.ok) {
        let message = "Erro na requisição";

        if (contentType && contentType.includes("application/json")) {
            const errorData = await res.json();
            message = errorData.error || errorData.message || message;
        } else {
            message = await res.text();
        }

        throw new Error(message);
    }

    if (contentType && contentType.includes("application/json")) {
        return res.json();
    }

    return null;
}