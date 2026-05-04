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
    const nameInput = document.getElementById("newService");
    const durationInput = document.getElementById("newServiceDuration");

    const name = nameInput.value.trim();
    const duration = parseInt(durationInput.value);

    if (!name) {
        alert("Digite o nome do serviço ou atendimento.");
        return;
    }

    if (!duration || duration <= 0) {
        alert("Digite a duração do serviço em minutos.");
        return;
    }

    try {
        await apiFetch("/admin/service/create", {
            method: "POST",
            body: JSON.stringify({
                name,
                duration
            })
        });

        nameInput.value = "";
        durationInput.value = "";

        await loadServices();
    } catch (err) {
        alert(err.message || "Não foi possível criar o serviço.");
    }
}

async function deleteService(name) {
    const confirmDelete = confirm(`Deseja excluir o serviço "${name}"?`);

    if (!confirmDelete) {
        return;
    }

    try {
        await apiFetch(`/admin/service/delete?name=${encodeURIComponent(name)}`, {
            method: "DELETE"
        });

        await loadServices();
    } catch (err) {
        alert(err.message || "Não foi possível excluir o serviço.");
    }
}

async function setHours() {
    const start = document.getElementById("start").value.trim();
    const end = document.getElementById("end").value.trim();
    const interval = parseInt(document.getElementById("interval").value);

    if (!start || !end || !interval) {
        alert("Preencha início, fim e intervalo dos horários.");
        return;
    }

    try {
        await apiFetch("/admin/hours/set", {
            method: "POST",
            body: JSON.stringify({
                start,
                end,
                interval,
            })
        });

        document.getElementById("hoursSummary").innerText = `${start} - ${end}`;
        alert("Horários de funcionamento salvos com sucesso.");
    } catch (err) {
        alert(err.message || "Não foi possível salvar os horários.");
    }
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

        const duration = s.duration || 30;

        li.innerHTML = `
            <div class="service-info">
                <strong>${s.name}</strong>
                <small>${duration} minutos de duração</small>
            </div>
            <button class="danger-button" onclick="deleteService('${s.name}')">Excluir</button>
        `;

        list.appendChild(li);
    });
}

loadServices();
loadAppointments();