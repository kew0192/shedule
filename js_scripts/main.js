const ip_logic = "http://localhost:8081";
const ip_auth = "http://localhost:8080";
const dayKeys = ['monday', 'tuesday', 'wednesday', 'thursday', 'friday'];
const dayNames = ['Понедельник', 'Вторник', 'Среда', 'Четверг', 'Пятница'];

// ===== ПРОСТАЯ ФУНКЦИЯ ДЛЯ ОБНОВЛЕНИЯ ТОКЕНА =====
async function refreshToken() {
    try {
        const refresh_token = localStorage.getItem("refresh_token");
        if (!refresh_token) {
            console.log("❌ Нет refresh токена");
            return false;
        }

        const response = await fetch(`${ip_auth}/refresh`, {
            method: "POST",
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ refresh_token: refresh_token })
        });

        if (!response.ok) {
            console.log("❌ Ошибка обновления токена:", response.status);
            return false;
        }

        const data = await response.json();
        if (data.access_token) {
            localStorage.setItem("access_token", data.access_token);
            console.log("✅ Токен обновлен!");
            return true;
        }
        return false;
    } catch (e) {
        console.error("❌ Ошибка:", e);
        return false;
    }
}

// ===== ПРОСТОЙ ЗАПРОС С ПРОВЕРКОЙ ТОКЕНА =====
async function fetchWithAuth(url, options = {}) {
    let token = localStorage.getItem("access_token");
    
    if (!token) {
        console.log("❌ Нет токена!");
        window.location.href = "/login.html";
        return;
    }
    
    options.headers = {
        ...options.headers,
        'Authorization': `Bearer ${token}`
    };

    let response = await fetch(url, options);

    if (response.status === 401) {
        console.log("🔄 401 ошибка, обновляем токен...");
        const refreshed = await refreshToken();
        
        if (refreshed) {
            const newToken = localStorage.getItem("access_token");
            options.headers['Authorization'] = `Bearer ${newToken}`;
            response = await fetch(url, options);
            console.log("✅ Запрос повторен с новым токеном");
            return response;
        } else {
            console.log("❌ Не удалось обновить токен");
            return response;
        }
    }

    return response;
}

// ===== ПРОВЕРКА РОЛИ =====
function isAdmin() {
    const role = localStorage.getItem("role");
    return role === "Админ" || role === "Admin" || role === "admin";
}

// ===== УПРАВЛЕНИЕ ВИДИМОСТЬЮ КНОПОК =====
function updateButtonsVisibility() {
    const hasAdminRights = isAdmin();
    
    const adminButtons = [
        "AddClass",
        "DeleteClass",
        "OpenTeacherModal",
        "OpenTeacherModal1",
        "AddLesson",
        "ViewUsers",
        "ExportAllClasses"
    ];
    
    adminButtons.forEach(buttonId => {
        const button = document.getElementById(buttonId);
        if (button) {
            button.style.display = hasAdminRights ? "inline-block" : "none";
        }
    });
}

document.addEventListener('DOMContentLoaded', async () => {
    const select = document.getElementById('class');
    const token = localStorage.getItem("access_token");
    if (!token) {
        select.innerHTML = `<option value="">Нет токена</option>`;
        return;
    }

    updateButtonsVisibility();
    await loadClasses();

    select.onchange = async () => {
        const name = select.value;
        if (!name) return;
        const res = await fetchWithAuth(ip_logic + `/class/${encodeURIComponent(name)}`, { 
            headers: { 'Content-Type': 'application/json' } 
        });
        
        if (res && res.ok) {
            const data = await res.json();
            showSchedule(data);
        }
    };
});

// ===== ЗАГРУЗКА СПИСКА КЛАССОВ =====
async function loadClasses() {
    const select = document.getElementById('class');
    
    try {
        const res = await fetchWithAuth(ip_logic + "/classes", { 
            headers: { 'Content-Type': 'application/json' } 
        });
        
        if (!res) {
            select.innerHTML = `<option value="">Ошибка подключения</option>`;
            return;
        }
        
        if (!res.ok) {
            select.innerHTML = `<option value="">Ошибка ${res.status}</option>`;
            return;
        }
        
        const data = await res.json();
        const classes = data.data || data || [];
        if (!classes.length) {
            select.innerHTML = `<option value="">Нет классов</option>`;
            return;
        }
        
        select.innerHTML = classes.map(c => `<option value="${c.name}">${c.name}</option>`).join('');
        
        if (classes.length > 0) {
            const firstName = classes[0].name;
            const res = await fetchWithAuth(ip_logic + `/class/${encodeURIComponent(firstName)}`, { 
                headers: { 'Content-Type': 'application/json' } 
            });
            
            if (res && res.ok) {
                const data = await res.json();
                showSchedule(data);
            }
        }
    } catch (e) {
        console.error("Ошибка:", e);
        select.innerHTML = `<option value="">Ошибка</option>`;
    }
}

// ===== СОХРАНЕННАЯ ТАБЛИЦА ДЛЯ ЭКСПОРТА =====
let currentScheduleData = null;

function showSchedule(data) {
    const tbody = document.getElementById('schedule-body');
    if (!tbody) return;
    
    const week = data.week;
    if (!week) {
        tbody.innerHTML = `<tr><td colspan="6">Нет расписания</td></tr>`;
        currentScheduleData = null;
        return;
    }
    
    currentScheduleData = data;
    
    const days = [week.monday, week.tuesday, week.wednesday, week.thursday, week.friday];
    
    let html = '';
    
    for (let lessonNum = 0; lessonNum < 8; lessonNum++) {
        html += `<tr>`;
        html += `<td><strong>${lessonNum + 1}</strong></td>`;
        
        for (let dayIdx = 0; dayIdx < 5; dayIdx++) {
            const day = days[dayIdx];
            const lesson = day && day[lessonNum];
            
            if (lesson) {
                html += `<td>${lesson.name} (${lesson.teacher_name})</td>`;
            } else {
                html += `<td></td>`;
            }
        }
        html += `</tr>`;
    }
    
    tbody.innerHTML = html;
}

// ===== ЭКСПОРТ ВСЕХ КЛАССОВ =====
document.getElementById("ExportAllClasses").onclick = async function() {
    if (!isAdmin()) {
        console.log("❌ Нет прав для экспорта всех классов");
        return;
    }
    
    try {
        const res = await fetchWithAuth(ip_logic + "/classes", { 
            headers: { 'Content-Type': 'application/json' } 
        });
        
        if (!res || !res.ok) {
            console.log("❌ Ошибка загрузки списка классов");
            return;
        }
        
        const data = await res.json();
        const classes = data.data || data || [];
        
        if (!classes.length) {
            console.log("❌ Нет классов для экспорта");
            return;
        }
        
        classes.sort((a, b) => {
            const numA = parseInt(a.name.match(/\d+/)?.[0] || 0);
            const numB = parseInt(b.name.match(/\d+/)?.[0] || 0);
            if (numA !== numB) return numA - numB;
            return a.name.localeCompare(b.name);
        });
        
        const allClassesData = [];
        for (const cls of classes) {
            const res = await fetchWithAuth(ip_logic + `/class/${encodeURIComponent(cls.name)}`, { 
                headers: { 'Content-Type': 'application/json' } 
            });
            
            if (res && res.ok) {
                const data = await res.json();
                allClassesData.push({
                    name: cls.name,
                    week: data.week
                });
            }
        }
        
        if (!allClassesData.length) {
            console.log("❌ Нет данных для экспорта");
            return;
        }
        
        let excelHtml = `
            <html xmlns:o="urn:schemas-microsoft-com:office:office" 
                  xmlns:x="urn:schemas-microsoft-com:office:excel" 
                  xmlns="http://www.w3.org/TR/REC-html40">
            <head>
                <meta charset="UTF-8">
                <style>
                    body { font-family: 'Segoe UI', Arial, sans-serif; padding: 20px; }
                    h2 { color: #2d5a2d; font-size: 24px; margin-bottom: 5px; }
                    .subtitle { color: #666; font-size: 14px; margin-bottom: 20px; }
                    table { border-collapse: collapse; width: 100%; font-size: 13px; border: 2px solid #2d5a2d; }
                    th { background: #2d7a2d; color: white; font-weight: bold; padding: 12px 15px; border: 1px solid #1a5a1a; text-align: center; }
                    td { padding: 10px 15px; border: 1px solid #c8dcc8; text-align: left; vertical-align: top; min-height: 60px; }
                    .day-cell { font-weight: bold; background: #e8f0e8; color: #2d5a2d; text-align: center; vertical-align: middle; }
                    .empty-cell { color: #b0c8b0; font-style: italic; text-align: center; }
                    .lesson-item { padding: 3px 0; border-bottom: 1px dashed #dde8dd; }
                    .lesson-item:last-child { border-bottom: none; }
                    .lesson-number { display: inline-block; background: #e8f0e8; color: #1a3a1a; border: 1px solid #c8dcc8; border-radius: 4px; width: 22px; height: 22px; text-align: center; line-height: 20px; font-size: 11px; font-weight: bold; margin-right: 8px; }
                    .lesson-name { color: #1a3a1a; }
                    .lesson-empty { color: #b0c8b0; }
                </style>
            </head>
            <body>
                <h2>📚 Расписание всех классов</h2>
                <p class="subtitle">Создано: ${new Date().toLocaleString()}</p>
                <br>
                <table>
                    <thead>
                        <tr>
                            <th style="min-width: 120px;">День</th>
        `;
        
        for (const classData of allClassesData) {
            excelHtml += `<th style="min-width: 180px;">${classData.name}</th>`;
        }
        excelHtml += `</tr></thead><tbody>`;
        
        for (let dayIdx = 0; dayIdx < 5; dayIdx++) {
            excelHtml += `<tr>`;
            excelHtml += `<td class="day-cell">${dayNames[dayIdx]}</td>`;
            
            for (const classData of allClassesData) {
                const week = classData.week;
                if (!week) {
                    excelHtml += `<td class="empty-cell">Нет данных</td>`;
                    continue;
                }
                
                const days = [week.monday, week.tuesday, week.wednesday, week.thursday, week.friday];
                const day = days[dayIdx];
                
                let lessonsHtml = '';
                for (let lessonNum = 1; lessonNum <= 8; lessonNum++) {
                    const lesson = day ? day[lessonNum - 1] : null;
                    if (lesson) {
                        lessonsHtml += `
                            <div class="lesson-item">
                                <span class="lesson-number">${lessonNum}</span>
                                <span class="lesson-name">${lesson.name}</span>
                            </div>
                        `;
                    } else {
                        lessonsHtml += `
                            <div class="lesson-item">
                                <span class="lesson-number">${lessonNum}</span>
                                <span class="lesson-empty">—</span>
                            </div>
                        `;
                    }
                }
                excelHtml += `<td>${lessonsHtml}</td>`;
            }
            excelHtml += `</tr>`;
        }
        
        excelHtml += `
                    </tbody>
                </table>
                <br>
                <p style="color: #666; font-size: 12px;">
                    <span style="display:inline-block; background:#e8f0e8; border:1px solid #c8dcc8; border-radius:4px; padding:0 6px; font-weight:bold;">1</span> - номер урока
                </p>
            </body>
            </html>
        `;
        
        const blob = new Blob([excelHtml], { 
            type: 'application/vnd.ms-excel;charset=utf-8' 
        });
        
        const link = document.createElement('a');
        link.href = URL.createObjectURL(blob);
        link.download = `Расписание_всех_классов_${new Date().toISOString().slice(0,10)}.xls`;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        URL.revokeObjectURL(link.href);
        
        console.log(`✅ Экспортировано ${allClassesData.length} классов!`);
    } catch (e) {
        console.error("Ошибка:", e);
    }
};

// ===== ЭКСПОРТ В EXCEL =====
document.getElementById("ExportExcel").onclick = function() {
    if (!currentScheduleData) {
        console.log("❌ Нет данных для экспорта");
        return;
    }
    
    const week = currentScheduleData.week;
    const days = [week.monday, week.tuesday, week.wednesday, week.thursday, week.friday];
    
    let excelHtml = `
        <html xmlns:o="urn:schemas-microsoft-com:office:office" 
              xmlns:x="urn:schemas-microsoft-com:office:excel" 
              xmlns="http://www.w3.org/TR/REC-html40">
        <head>
            <meta charset="UTF-8">
            <style>
                table { border-collapse: collapse; width: 100%; }
                th, td { border: 1px solid #000; padding: 8px; text-align: center; }
                th { background: #4CAF50; color: white; font-weight: bold; }
            </style>
        </head>
        <body>
            <h2>Расписание - ${document.getElementById('class').value}</h2>
            <table>
                <thead>
                    <tr>
                        <th>Урок</th>
                        <th>Понедельник</th>
                        <th>Вторник</th>
                        <th>Среда</th>
                        <th>Четверг</th>
                        <th>Пятница</th>
                    </tr>
                </thead>
                <tbody>
    `;
    
    for (let lessonNum = 0; lessonNum < 8; lessonNum++) {
        excelHtml += `<tr>`;
        excelHtml += `<td><strong>${lessonNum + 1}</strong></td>`;
        
        for (let dayIdx = 0; dayIdx < 5; dayIdx++) {
            const day = days[dayIdx];
            const lesson = day && day[lessonNum];
            
            if (lesson) {
                excelHtml += `<td>${lesson.name}</td>`;
            } else {
                excelHtml += `<td></td>`;
            }
        }
        excelHtml += `</tr>`;
    }
    
    excelHtml += `
                </tbody>
            </table>
        </body>
        </html>
    `;
    
    const blob = new Blob([excelHtml], { 
        type: 'application/vnd.ms-excel;charset=utf-8' 
    });
    
    const link = document.createElement('a');
    link.href = URL.createObjectURL(blob);
    link.download = `Расписание_${document.getElementById('class').value}_${new Date().toISOString().slice(0,10)}.xls`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(link.href);
};

// ===== ПОКАЗАТЬ РАСПИСАНИЕ УЧИТЕЛЯ (СКАЧИВАНИЕ EXCEL) =====
document.getElementById("MySchedule").onclick = async function() {
    const firstName = localStorage.getItem("first_name");
    const lastName = localStorage.getItem("last_name");
    
    if (!firstName || !lastName) {
        console.log("❌ Не найдены данные пользователя");
        return;
    }
    
    const teacherName = firstName + " " + lastName;
    console.log("👨‍🏫 Поиск расписания для учителя:", teacherName);
    
    try {
        const res = await fetchWithAuth(ip_logic + "/classes", { 
            headers: { 'Content-Type': 'application/json' } 
        });
        
        if (!res || !res.ok) {
            console.log("❌ Ошибка загрузки классов");
            return;
        }
        
        const data = await res.json();
        const classes = data.data || data || [];
        
        if (!classes.length) {
            console.log("❌ Нет классов");
            return;
        }
        
        let teacherSchedule = {
            monday: {},
            tuesday: {},
            wednesday: {},
            thursday: {},
            friday: {}
        };
        
        let foundLessons = false;
        
        for (const cls of classes) {
            const res = await fetchWithAuth(ip_logic + `/class/${encodeURIComponent(cls.name)}`, { 
                headers: { 'Content-Type': 'application/json' } 
            });
            
            if (!res || !res.ok) continue;
            
            const classData = await res.json();
            const week = classData.week;
            if (!week) continue;
            
            for (let dayIdx = 0; dayIdx < 5; dayIdx++) {
                const dayKey = dayKeys[dayIdx];
                const day = week[dayKey];
                if (!day) continue;
                
                for (let lessonNum = 0; lessonNum < 8; lessonNum++) {
                    const lesson = day[lessonNum];
                    if (lesson && lesson.teacher_name === teacherName) {
                        foundLessons = true;
                        if (!teacherSchedule[dayKey][lessonNum]) {
                            teacherSchedule[dayKey][lessonNum] = [];
                        }
                        teacherSchedule[dayKey][lessonNum].push({
                            class: cls.name,
                            subject: lesson.name
                        });
                    }
                }
            }
        }
        
        if (!foundLessons) {
            console.log(`❌ Учитель ${teacherName} не найден ни в одном расписании`);
            return;
        }
        
        exportTeacherScheduleToExcel(teacherSchedule, teacherName);
        
    } catch (e) {
        console.error("Ошибка:", e);
    }
};

// ===== ЭКСПОРТ РАСПИСАНИЯ УЧИТЕЛЯ В EXCEL =====
function exportTeacherScheduleToExcel(schedule, teacherName) {
    let excelHtml = `
        <html xmlns:o="urn:schemas-microsoft-com:office:office" 
              xmlns:x="urn:schemas-microsoft-com:office:excel" 
              xmlns="http://www.w3.org/TR/REC-html40">
        <head>
            <meta charset="UTF-8">
            <style>
                body { font-family: 'Segoe UI', Arial, sans-serif; padding: 20px; }
                h2 { color: #2d5a2d; font-size: 24px; margin-bottom: 5px; }
                .subtitle { color: #666; font-size: 14px; margin-bottom: 20px; }
                table { border-collapse: collapse; width: 100%; font-size: 13px; border: 2px solid #2d5a2d; }
                th { background: #2d7a2d; color: white; font-weight: bold; padding: 12px 15px; border: 1px solid #1a5a1a; text-align: center; }
                td { padding: 10px 15px; border: 1px solid #c8dcc8; text-align: left; vertical-align: top; min-height: 60px; }
                .lesson-item { padding: 3px 0; border-bottom: 1px dashed #dde8dd; }
                .lesson-item:last-child { border-bottom: none; }
                .lesson-number { display: inline-block; background: #e8f0e8; color: #1a3a1a; border: 1px solid #c8dcc8; border-radius: 4px; width: 22px; height: 22px; text-align: center; line-height: 20px; font-size: 11px; font-weight: bold; margin-right: 8px; }
                .lesson-name { color: #1a3a1a; }
                .lesson-empty { color: #b0c8b0; }
            </style>
        </head>
        <body>
            <h2>📚 Расписание учителя: ${teacherName}</h2>
            <p class="subtitle">Создано: ${new Date().toLocaleString()}</p>
            <br>
            <table>
                <thead>
                    <tr>
                        <th style="min-width: 50px;">Урок</th>
                        <th>Понедельник</th>
                        <th>Вторник</th>
                        <th>Среда</th>
                        <th>Четверг</th>
                        <th>Пятница</th>
                    </tr>
                </thead>
                <tbody>
    `;
    
    for (let lessonNum = 0; lessonNum < 8; lessonNum++) {
        excelHtml += `<tr>`;
        excelHtml += `<td><strong>${lessonNum + 1}</strong></td>`;
        
        for (let dayIdx = 0; dayIdx < 5; dayIdx++) {
            const dayKey = dayKeys[dayIdx];
            const day = schedule[dayKey];
            const lesson = day && day[lessonNum];
            
            if (lesson && lesson.length > 0) {
                const subjects = lesson.map(l => `${l.subject} (${l.class})`).join('<br>');
                excelHtml += `<td>${subjects}</td>`;
            } else {
                excelHtml += `<td class="lesson-empty">—</td>`;
            }
        }
        excelHtml += `</tr>`;
    }
    
    excelHtml += `
                </tbody>
            </table>
        </body>
        </html>
    `;
    
    const blob = new Blob([excelHtml], { 
        type: 'application/vnd.ms-excel;charset=utf-8' 
    });
    
    const link = document.createElement('a');
    link.href = URL.createObjectURL(blob);
    link.download = `Расписание_учителя_${teacherName.replace(' ', '_')}_${new Date().toISOString().slice(0,10)}.xls`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(link.href);
    
    console.log(`✅ Расписание учителя ${teacherName} экспортировано!`);
}

// ===== МОДАЛКА ДЛЯ ПОЛЬЗОВАТЕЛЕЙ =====
const usersModal = document.getElementById("UsersModal");
const usersList = document.getElementById("users-list");

document.getElementById("ViewUsers").onclick = async () => {
    if (!isAdmin()) {
        console.log("❌ Нет прав для просмотра пользователей");
        return;
    }
    
    usersModal.showModal();
    usersList.innerHTML = '<p>Загрузка...</p>';
    
    try {
        const res = await fetchWithAuth(`${ip_auth}/admin/users`, {
            headers: { 'Content-Type': 'application/json' }
        });
        
        if (!res) {
            usersList.innerHTML = '<p style="color: red;">Ошибка: нет ответа от сервера</p>';
            return;
        }
        
        if (!res.ok) {
            const err = await res.json();
            usersList.innerHTML = `<p style="color: red;">Ошибка: ${err.error || "Неизвестная ошибка"}</p>`;
            return;
        }
        
        const data = await res.json();
        const users = data.users || [];
        
        if (users.length === 0) {
            usersList.innerHTML = '<p>Нет пользователей</p>';
            return;
        }
        
        let html = '<table border="1" style="width: 100%; border-collapse: collapse;">';
        html += `
            <thead>
                <tr>
                    <th>#</th>
                    <th>Имя</th>
                    <th>Фамилия</th>
                    <th>Код</th>
                    <th>Роль</th>
                </tr>
            </thead>
            <tbody>
        `;
        
        users.forEach((user, index) => {
            html += `
                <tr>
                    <td>${index + 1}</td>
                    <td>${user.First_name || user.first_name || ''}</td>
                    <td>${user.Last_name || user.last_name || ''}</td>
                    <td>${user.Code || user.code || ''}</td>
                    <td>${user.Role || user.role || ''}</td>
                </tr>
            `;
        });
        
        html += '</tbody></table>';
        html += `<p>Всего пользователей: ${users.length}</p>`;
        
        usersList.innerHTML = html;
    } catch (e) {
        console.error("Ошибка:", e);
        usersList.innerHTML = `<p style="color: red;">Ошибка: ${e.message}</p>`;
    }
};

document.getElementById("CloseUsersModal").onclick = () => usersModal.close();
usersModal.onclick = (e) => { if (e.target === usersModal) usersModal.close(); };

// ===== УДАЛЕНИЕ КЛАССА =====
document.getElementById("DeleteClass").onclick = async () => {
    if (!isAdmin()) {
        console.log("❌ Нет прав для удаления классов");
        return;
    }
    
    const select = document.getElementById('class');
    const className = select.value;
    
    if (!className) {
        console.log("❌ Класс не выбран");
        return;
    }
    
    if (!confirm(`Вы уверены, что хотите удалить класс "${className}"?`)) {
        return;
    }
    
    const res = await fetchWithAuth(ip_logic + "/class", {
        method: "DELETE",
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ name: className })
    });
    
    if (res && res.ok) {
        console.log(`✅ Класс "${className}" удален`);
        await loadClasses();
    } else {
        const err = res ? await res.json() : { error: "Ошибка запроса" };
        console.log(`❌ Ошибка: ${err.error}`);
    }
};

// ===== МОДАЛКА ДЛЯ КЛАССА =====
const modalClass = document.getElementById("ClassAdd");
document.getElementById("AddClass").onclick = () => {
    if (!isAdmin()) {
        console.log("❌ Нет прав для создания классов");
        return;
    }
    modalClass.showModal();
};
document.getElementById("CloseClassAdd").onclick = () => modalClass.close();
modalClass.onclick = (e) => { if (e.target === modalClass) modalClass.close(); };

document.getElementById("AddClassBtn").onclick = async () => {
    const value = document.getElementById("addclassinput").value.trim();
    
    const res = await fetchWithAuth(ip_logic + "/class", {
        method: "POST",
        headers: { 
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ name: value })
    });
    if (res && res.ok) { 
        modalClass.close(); 
        await loadClasses();
    }
};

// ===== МОДАЛКА ДЛЯ УЧИТЕЛЯ (ПРОСМОТР) =====
const modalTeacher = document.getElementById("TeacherAdd");
document.getElementById("OpenTeacherModal").onclick = () => {
    if (!isAdmin()) {
        console.log("❌ Нет прав для добавления учителей");
        return;
    }
    modalTeacher.showModal();
};
document.getElementById("CloseTeacherAdd").onclick = () => modalTeacher.close();
modalTeacher.onclick = (e) => { if (e.target === modalTeacher) modalTeacher.close(); };

document.getElementById("AddTeacherBtn").onclick = async () => {
    const firstName = document.getElementById("f").value.trim();
    const lastName = document.getElementById("l").value.trim();
    const subject = document.getElementById("teacher_subject").value.trim();

    if (!firstName || !lastName || !subject) {
        console.log("❌ Заполните все поля");
        return;
    }

    const resAuth = await fetchWithAuth("http://localhost:8080/admin/teacher", {
        method: "POST",
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ 
            first_name: firstName, 
            last_name: lastName,
            code: firstName.toLowerCase() + lastName.toLowerCase() + Math.floor(Math.random() * 1000)
        })
    });

    if (!resAuth || !resAuth.ok) {
        const err = resAuth ? await resAuth.json() : { error: "Ошибка запроса" };
        console.log(`❌ Ошибка: ${err.error}`);
        return;
    }

    modalTeacher.close();
    console.log(`✅ Учитель ${firstName} ${lastName} создан!`);
    location.reload();
};

// ===== МОДАЛКА ДЛЯ УЧИТЕЛЯ (РАСПИСАНИЕ) =====
const modalTeacher1 = document.getElementById("TeacherAdd1");
document.getElementById("OpenTeacherModal1").onclick = () => {
    if (!isAdmin()) {
        console.log("❌ Нет прав для добавления учителей");
        return;
    }
    modalTeacher1.showModal();
};
document.getElementById("CloseTeacherAdd1").onclick = () => modalTeacher1.close();
modalTeacher1.onclick = (e) => { if (e.target === modalTeacher1) modalTeacher1.close(); };

document.getElementById("AddTeacherBtn1").onclick = async () => {
    const firstName = document.getElementById("f1").value.trim();
    const lastName = document.getElementById("l1").value.trim();
    const subject = document.getElementById("teacher_subject1").value.trim();

    if (!firstName || !lastName || !subject) {
        console.log("❌ Заполните все поля");
        return;
    }

    const resLogic = await fetchWithAuth("http://localhost:8081/teacher", {
        method: "POST",
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            name: firstName + " " + lastName,
            lesson: {
                name: subject
            }
        })
    });

    if (!resLogic || !resLogic.ok) {
        const err = resLogic ? await resLogic.json() : { error: "Ошибка запроса" };
        console.log(`❌ Ошибка: ${err.error}`);
        return;
    }

    modalTeacher1.close();
    console.log(`✅ Учитель ${firstName} ${lastName} создан!`);
    location.reload();
};

// ===== МОДАЛКА ДЛЯ УРОКОВ =====
const modalLesson = document.getElementById("LessonAdd");

document.getElementById("AddLesson").onclick = () => {
    if (!isAdmin()) {
        console.log("❌ Нет прав для добавления уроков");
        return;
    }
    modalLesson.showModal();
};
document.getElementById("CloseLessonAdd").onclick = () => modalLesson.close();
modalLesson.onclick = (e) => { if (e.target === modalLesson) modalLesson.close(); };

document.getElementById("AddLessonBtn").onclick = async () => {
    const class_name = document.getElementById("lesson_class").value.trim();
    const teacher_name = document.getElementById("lesson_teacher").value.trim();
    const subject = document.getElementById("lesson_subject").value.trim();
    const hours = parseInt(document.getElementById("lesson_hours").value);

    if (!class_name || !teacher_name || !subject || !hours) {
        console.log("❌ Заполните все поля");
        return;
    }

    const res = await fetchWithAuth(ip_logic + "/lessons/add", {
        method: "POST",
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ class_name, teacher_name, subject, hours })
    });

    if (res && res.ok) {
        modalLesson.close();
        console.log(`✅ Уроки добавлены!`);
        location.reload();
    } else {
        const err = res ? await res.json() : { error: "Ошибка запроса" };
        console.log(`❌ Ошибка: ${err.error}`);
    }
};