


const socket = new WebSocket("ws://localhost:8080/api/chat");

// Connection opened
socket.addEventListener("open", (event) => {
  socket.send("Hello Server!");
});

// Listen for messages
socket.addEventListener("message", (event) => {
  console.log("Message from server ", event.data);
});

// Listen for possible errors
socket.addEventListener("error", (event) => {
  console.log("WebSocket error: ", event);
});

// Listen for closing message
socket.addEventListener("close", (event) => {
  console.log("The connection has been closed successfully.");
});