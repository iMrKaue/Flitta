async function loadServices() {
    const services = await apiFetch("/admin/service/list");

    const list = document.getElementById("services");
    list.innerHTML = "";

    services.forEach(s => {
        const li = document.createElement("li");

        li.innerHTML = `
            ${s.name}
            <button onclick="deleteService('${s.name}')">X</button>
        `;

        list.appendChild(li);
    });
}

function logout() {
    localStorage.removeItem("token");
    window.location.href = "login.html";
}

async function createService() {
    const name = document.getElementById("newService").value;

    if (!name) {
        alert("Digite um nome");
        return;
    }

    await apiFetch("/admin/service/create", {
        method: "POST",
        body: JSON.stringify({ name })
    });

    document.getElementById("newService").value = "";
    loadServices();
}

async function deleteService(name) {
    await apiFetch(`/admin/service/delete?name=${name}`, {
        method: "DELETE"
    });

    loadServices();
}

async function setHours() {
    const start = document.getElementById("start").value;
    const end = document.getElementById("end").value;
    const interval = parseInt(document.getElementById("interval").value);

    if (!start || !end || !interval) {
        alert("Preencha tudo");
        return;
    }

    await apiFetch("/admin/hours/set", {
        method: "POST",
        body: JSON.stringify({
            start,
            end,
            interval,
        })
    });

    alert("Horários salvos com sucesso");
}

async function loadAppointments() {
    const data = await apiFetch("/appointments");

    console.log("APPOINTMENTS:", data); // 👈 ADICIONA ISSO

    const list = document.getElementById("appointments");
    list.innerHTML = "";

    data.forEach(a => {
        const li = document.createElement("li");
    
        li.innerHTML = `
            <div style="margin-bottom:10px">
                <strong>${a.Name}</strong> - ${a.Service}<br>
                <small>📅 ${a.Date} às ${a.Time}</small>
            </div>
        `;
    
        list.appendChild(li);
    });
}

loadServices();
loadAppointments();