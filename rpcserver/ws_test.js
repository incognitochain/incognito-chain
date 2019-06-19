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
const data1 = {
    "request":
        {
            "jsonrpc": "1.0",
            "method": "subcribenewshardblock",
            "params": [0],
            "id": 1
        },
    "subcription": "13",
    "type": 1,
}

const socket = new WebSocket(url)

socket.onopen = () => {
    socket.send(JSON.stringify(data))
    // setTimeout(()=>socket.send(JSON.stringify(data1)), 10000)
    // setTimeout(()=>socket.send(JSON.stringify(data1)), 13000)
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
