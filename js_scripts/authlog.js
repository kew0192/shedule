const ip_auth = "http://localhost:8080"
const login = document.getElementById("login");
const btn = document.getElementById("go_to_main");
btn.addEventListener("click", async () => {
    const response = await fetch(ip_auth + "/login", {
            method: "POST",
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({code: login.value,})
        });
        const result = await response.json();
        if (response.ok){
            let user = result.user
            localStorage.setItem("user_id", user.id);
            localStorage.setItem("first_name", user.first_name);
            localStorage.setItem("last_name", user.last_name);
            localStorage.setItem("role", user.role);
            localStorage.setItem("code", user.code);
            localStorage.setItem("access_token", result.access_token);
            localStorage.setItem("refresh_token", result.refresh_token);

            window.location.href = '/main.html';
        }
        
})