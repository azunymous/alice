async function registerForm() {
    let formData = $("#registerForm").serialize();
    console.log(formData)
    const response = await postData('/api/register', formData);
    //extract JSON from the http response
    console.log(response)
    if (response["status"] === "SUCCESS") {
        localStorage.setItem('token', response["token"]);
        localStorage.setItem('username', response["username"]);
        window.location.replace("/");
        return response["status"]
    }
    $("#error").text("Error: " + response["error"]);
    return response

}

async function anonRegister() {
    console.log("Anon login")
    const response = await postData('/api/anonregister', "{}");
    //extract JSON from the http response
    console.log(response)
    if (response["status"] === "SUCCESS") {
        localStorage.setItem('token', response["token"]);
        localStorage.setItem('username', response["username"]);
        window.location.replace("/");
        return response["status"]
    }
    $("#error").text("Error: " + response["error"]);
    return response

}

async function loginForm() {
    let formData = $("#loginForm").serialize();
    console.log(formData)
    const response = await postData('/api/login', formData);
    //extract JSON from the http response
    console.log(response)
    if (response["status"] === "SUCCESS") {
        localStorage.setItem('token', response["token"]);
        localStorage.setItem('username', response["username"]);
        window.location.replace("/");
        return response["status"]
    }
    $("#error").text("Error: " + response["error"]);
    return response

}

async function verifyUser() {
    let token = localStorage.getItem('token');

    const response = await postData('/api/verify', "token=" + token);
    //extract JSON from the http response
    console.log(response);
    if (response["status"] === "SUCCESS") {
        localStorage.setItem('token', response["token"]);
        return true
    }
    localStorage.removeItem('token');
    return false

}

function postData(url = '', data = {}) {
    // Default options are marked with *
    return fetch(url, {
        method: 'POST', // *GET, POST, PUT, DELETE, etc.
        mode: 'cors', // no-cors, cors, *same-origin
        cache: 'no-cache', // *default, no-cache, reload, force-cache, only-if-cached
        credentials: 'same-origin', // include, *same-origin, omit
        headers: {
            // 'Content-Type': 'application/json',
            'Content-Type': 'application/x-www-form-urlencoded',
        },
        // redirect: 'follow', // manual, *follow, error
        referrer: 'no-referrer', // no-referrer, *client
        body: data, // body data type must match "Content-Type" header
    })
        .then(response => response.json()); // parses JSON response into native Javascript objects
}

function postJSON(url = '', data = {}) {
    // Default options are marked with *
    return fetch(url, {
        method: 'POST', // *GET, POST, PUT, DELETE, etc.
        mode: 'cors', // no-cors, cors, *same-origin
        cache: 'no-cache', // *default, no-cache, reload, force-cache, only-if-cached
        credentials: 'same-origin', // include, *same-origin, omit
        headers: {
            'Content-Type': 'application/json',
        },
        // redirect: 'follow', // manual, *follow, error
        referrer: 'no-referrer', // no-referrer, *client
        body: data, // body data type must match "Content-Type" header
    })
        .then(response => response.json()); // parses JSON response into native Javascript objects
}

function updateWeather(latitude, longitude) {
    latitude = latitude || 34.0522;
    longitude = longitude || 118.2437;
    console.log(latitude, longitude);
    return async function () {
        let w = await getAPI(latitude, longitude);
        document.getElementById("weather").innerHTML = w.content
    };
}

// let params = (new URL(location)).searchParams;

window.onload = async function () {
    let verified = await verifyUser();
    if (verified) {
        let username = localStorage.getItem("username");
        document.getElementById("greeting").innerHTML = "Welcome " + username + "!"
        document.getElementById("loginFieldSet").setAttribute("hidden", "")
    }
};
