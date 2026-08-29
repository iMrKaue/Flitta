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
      localStorage.removeItem("token");
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
  const responsibleName = document.getElementById("registerResponsibleName").value;
  const businessName = document.getElementById("registerName").value;
  const phone = document.getElementById("registerPhone").value;
  const email = document.getElementById("registerEmail").value;
  const password = document.getElementById("registerPassword").value;
  const businessType = document.getElementById("businessType").value;

  const errorElement = document.getElementById("registerError");

  if (!responsibleName || !businessName || !phone || !email || !password || !businessType) {
    localStorage.removeItem("token");
    errorElement.innerText = "Preencha todos os campos";
    return;
  }

  try {
    const res = await fetch(`${API_URL}/auth/register`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json"
      },
      body: JSON.stringify({
        responsible_name: responsibleName,
        business_name: businessName,
        phone,
        email,
        password,
        business_type: businessType
      })
    });

    if (!res.ok) {
      localStorage.removeItem("token");
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