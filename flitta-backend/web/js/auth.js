const API_URL = "http://localhost:8080";

async function login() {
  const email = document.getElementById("email").value;
  const password = document.getElementById("password").value;

  try {
    const res = await fetch(`${API_URL}/auth/login`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json"
      },
      body: JSON.stringify({ email, password })
    });

    if (!res.ok) {
      document.getElementById("error").innerText = "Email ou senha inválidos";
      return;
    }

    const data = await res.json();

    // 🔐 salva token
     localStorage.setItem("token", data.token);

    // 👉 vai pro dashboard
    window.location.href = "dashboard.html";

  } catch (err) {
    document.getElementById("error").innerText = "Erro ao conectar com o servidor";
  }
}