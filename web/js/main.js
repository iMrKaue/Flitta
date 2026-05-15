function formatHour(value) {
    if (!value) {
        return "";
    }

    return value.substring(0, 5);
}

function isValidHour(value) {
    return /^([01]\d|2[0-3]):[0-5]\d$/.test(value);
}

function isStartBeforeEnd(start, end) {
    return start < end;
}

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

async function loadWorkingHours() {
    const hoursSummary = document.getElementById("hoursSummary");
    const hoursStatus = document.getElementById("hoursStatus");

    try {
        const data = await apiFetch("/admin/hours/get");

        if (!data || !data.configured) {
            hoursSummary.innerText = "--";
            hoursStatus.innerText = "Nenhum horário de funcionamento configurado ainda.";
            return;
        }

        const start = formatHour(data.start);
        const end = formatHour(data.end);
        const interval = data.interval;

        document.getElementById("start").value = start;
        document.getElementById("end").value = end;
        document.getElementById("interval").value = interval;

        hoursSummary.innerText = `${start} - ${end}`;
        hoursStatus.innerText = `Funcionamento configurado das ${start} às ${end}, com intervalo de ${interval} minutos.`;
    } catch (err) {
        hoursSummary.innerText = "--";
        hoursStatus.innerText = "Não foi possível carregar os horários configurados.";
    }
}

async function setHours() {
    const startInput = document.getElementById("start");
    const endInput = document.getElementById("end");
    const intervalInput = document.getElementById("interval");
    const hoursStatus = document.getElementById("hoursStatus");

    const start = startInput.value.trim();
    const end = endInput.value.trim();
    const interval = parseInt(intervalInput.value);

    if (!start || !end || !interval) {
        alert("Preencha início, fim e intervalo dos horários.");
        return;
    }

    if (!isValidHour(start) || !isValidHour(end)) {
        alert("Use horários no formato HH:MM. Exemplo: 09:00.");
        return;
    }

    if (!isStartBeforeEnd(start, end)) {
        alert("O horário de início deve ser menor que o horário de fim.");
        return;
    }

    if (interval < 15) {
        alert("O intervalo mínimo deve ser de 15 minutos.");
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
        hoursStatus.innerText = `Funcionamento configurado das ${start} às ${end}, com intervalo de ${interval} minutos.`;

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

function businessTypeLabel(value) {
    const labels = {
        beauty: "Salão / Estética",
        barber: "Barbearia",
        clinic: "Clínica",
        gym: "Academia / Personal",
        petshop: "Petshop",
        other: "Outro"
    };

    return labels[value] || "Outro";
}

async function loadCompanySettings() {
    const companyStatus = document.getElementById("companyStatus");

    try {
        const data = await apiFetch("/admin/company/get");

        if(!data) {
            companyStatus.innerText = "Não foi possível carregar os dados da empresa.";
            return;
        }

        document.getElementById("companyName").value = data.name || "";
        document.getElementById("companyPhone").value = data.phone || "";
        document.getElementById("companyBusinessType").value = data.business_type || "other";

        companyStatus.innerText = `Empresa carregada: ${data.name || "sem nome"} • ${businessTypeLabel(data.business_type)}`;
    } catch (err) {
        companyStatus.innerText = "Não foi possível carregar os dados da empresa.";
    }
} 

async function updateCompanySettings() {
    const nameInput = document.getElementById("companyName");
    const phoneInput = document.getElementById("companyPhone");
    const businessTypeInput = document.getElementById("companyBusinessType");
    const companyStatus = document.getElementById("companyStatus");

    const name = nameInput.value.trim();
    const phone = phoneInput.value.trim();
    const business_type = businessTypeInput.value;

    if(!name) {
        alert("Digite o nome do estabelecimento.");
        return;
    }

    if (!phone) {
        alert("Digite o WhatsApp do estabelecimento.");
        return;
    }

    try {
        const data = await apiFetch("/admin/company/update", {
            method: "PUT",
            body: JSON.stringify({
                name,
                phone,
                business_type
            })
        });

        companyStatus.innerText = `Dados atualizados: ${data.name} • ${businessTypeLabel(data.business_type)}`;

        alert("Dados da empresa salvos com sucesso.");
    } catch (err) {
        alert(err.message || "Não foi possível atualizar os dados da empresa.");
    }
}

loadCompanySettings();
loadServices();
loadAppointments();
loadWorkingHours();