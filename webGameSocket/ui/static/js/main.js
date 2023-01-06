var id = "";
const socket = new WebSocket("ws://127.0.0.1:4000/ws");
document.addEventListener("DOMContentLoaded", function () {

    console.log("Attempting Connection...");
    socket.onopen = () => {
        console.log("Successfully Connected");
    };
    socket.onerror = error => {
        let data = {"event": "error", "payload": error}
        socket.send(JSON.stringify(data));
    };
    socket.onclose = event => {
        console.log("Socket Closed Connection: ", event);
        let data = {"event": "closed", "payload": event}
        socket.send(JSON.stringify(data));
    };
    socket.onmessage = (msg) => {
        console.log(msg);
        var el = document.querySelector('#data');
        let data = JSON.parse(msg.data);
        console.log(data);
        id = data.socket_id;
        el.innerHTML = data.template;

        const btns = document.querySelectorAll(".btn-ready");
        btns.forEach(btn => {
            if (btn.id !== id) {
                btn.style.display = "none";
            }
        });
        const p = document.querySelectorAll(".p-ready");
        p.forEach(p => {
            if (p.id === id) {
                p.style.display = "none";
            }
        });
    };
});


function valider() {
    var pseudo = document.querySelector('#pseudo').value
    let data = {"event": "login", "payload": pseudo}
    socket.send(JSON.stringify(data));
}


function readySet() {

    let data = {"event": "ready", "payload": ""};
    socket.send(JSON.stringify(data));


}

function sendGuess(even) {
    var x = event.charCode;
    if (x == 13) {
        var guessWord = document.querySelector('#guessWord').value
        let data = {"event": "guess", "payload": guessWord};
        socket.send(JSON.stringify(data));
    }
}


