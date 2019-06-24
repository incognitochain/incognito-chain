// nodejs
const WebSocket = require('ws')
const url = 'ws://localhost:19334/'

const data = {
    "request":
        {
            "jsonrpc": "1.0",
            "method": "subcribenewshardblock",
            "params": [0],
            "id": 1
        },
    "subcription": "11",
    "type": 0,
}

const socket = new WebSocket(url)

socket.onopen = () => {
    socket.send(JSON.stringify(data))
}
socket.onerror = (err) => {
    console.log(`websocket error: ${err}`, err)
}
socket.onmessage = e => {
    console.log(`Received data: ${e.data}`, e)
}
socket.onclose = () => {
    console.log("Closed Stream")
}
