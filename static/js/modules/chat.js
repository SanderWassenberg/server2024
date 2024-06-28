
const other_msg_template = document.querySelector("#other_msg");
const own_msg_template   = document.querySelector("#own_msg");
const messages           = document.querySelector("#messages");
const textbox            = document.querySelector("#textbox");

textbox.addEventListener("keydown", e => {
	if (e.code !== "Enter" || e.shiftKey) return;

	const msg = textbox.value;
	if (msg === "") {
		e.preventDefault();
		return;
	}

	textbox.value = "";
	e.preventDefault();


	add_msg(msg, true);
	send_msg(msg);
});

const socket = new WebSocket(`${location.protocol.replace("http", "ws")}//${location.host}/api/chat${location.search}`);

// Listen for messages
socket.addEventListener("message", (e) => { add_msg(e.data, false); });

// Listen for possible errors
socket.addEventListener("error", (e) => {
	alert("websocket error");
});

// Listen for closing message
socket.addEventListener("close", (e) => {
	alert("websocket closed");
});



function add_msg(text, from_self) {
	let frag = (from_self ? own_msg_template : other_msg_template).content.cloneNode(true);
	let msg  = frag.querySelector(".msg");
	msg.innerText = text;

	const scroll_down = messages.scrollTop === messages.scrollTopMax || from_self; // already at bottom, or from self
	messages.append(msg);

    if (scroll_down) messages.scrollTop = messages.scrollTopMax;
}

function send_msg(text) {
	socket.send(text);
}