async function loadAppointments() {
    const data = await apiFetch("/appointments");

    const list = document.getElementById("appointments");
    const appointmentsCount = document.getElementById("appointmentsCount");

    list.innerHTML = "";

    if (!data || data.length === 0) {
        appointmentsCount.innerText = "0";

        const li = document.createElement("li");
        li.className = "empty-state";
        li.innerHTML = `
            <strong>Nenhum agendamento para hoje.</strong>
            <span>Quando houver agendamentos, eles aparecerão aqui.</span>
        `;
        list.appendChild(li);
        return;
    }

    appointmentsCount.innerText = data.length;

    data.forEach(a => {
        const li = document.createElement("li");
        li.className = "appointment-item";

        li.innerHTML = `
            <div>
                <strong>${a.Name}</strong>
                <span>${a.Service}</span>
            </div>
            <small>📅 ${a.Date} às ${a.Time}</small>
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
        alert("Digite o nome do serviço ou atendimento");
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
        document.getElementById("hoursSummary").innerText = `${start} - ${end}`;
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

    alert("Horários de funcionamento salvos com sucesso");
}

async function loadServices() {
    const services = await apiFetch("/admin/service/list");

    const list = document.getElementById("services");
    const servicesCount = document.getElementById("servicesCount");

    list.innerHTML = "";

    if (!services || services.length === 0) {
        servicesCount.innerText = "0";

        const li = document.createElement("li");
        li.className = "empty-state";
        li.innerHTML = `
            <strong>Nenhum serviço cadastrado ainda.</strong>
            <span>Adicione o primeiro serviço ou atendimento abaixo.</span>
        `;
        list.appendChild(li);
        return;
    }

    servicesCount.innerText = services.length;

    services.forEach(s => {
        const li = document.createElement("li");
        li.className = "list-item";

        li.innerHTML = `
            <span>${s.name}</span>
            <button class="danger-button" onclick="deleteService('${s.name}')">Excluir</button>
        `;

        list.appendChild(li);
    });
}

loadServices();
loadAppointments();