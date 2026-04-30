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

async function register() {
  const name = document.getElementById("registerName").value;
  const phone = document.getElementById("registerPhone").value;
  const email = document.getElementById("registerEmail").value;
  const password = document.getElementById("registerPassword").value;
  const businessType = document.getElementById("businessType").value;

  const errorElement = document.getElementById("registerError");

  if (!name || !phone || !email || !password || !businessType) {
    errorElement.innerText = "Preencha todos os campos";
  }

  try {
    const res = await fetch(`${API_URL}/auth/register`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json"
      },
      body: JSON.stringify({
        name,
        phone,
        email,
        password,
        business_type: businessType
      })
    });

    if (!res.ok) {
      errorElement.innerText = "Não foi possível cadastrar o estabelecimento";
      return;
    }

    const data = await res.json();

    localStorage.setItem("token", data.token);

    window.location.href = "dashboard.html";
  } catch (err) {
    errorElement.innerText = "Erro ao conectar com o seridor";
  }
}