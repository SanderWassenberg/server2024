import { api_post } from "./util.js"


const search_params = get_search_params();

if (search_params.with === undefined) {
	alert("Invalid url. Must be 'chat/?with=other_user'");
	location.href = "/search";
}

const other_msg_template = document.querySelector("#other_msg");
const own_msg_template   = document.querySelector("#own_msg");
const messages           = document.querySelector("#messages");
const textbox            = document.querySelector("#textbox");
const username_elem      = document.querySelector("#username");

username_elem.innerText = search_params.with;

(async () => {
	let response = await api_post("/chat_history", search_params.with);
	if (!await check_response(response)) return;

	let history = await response.json();
	for (let i = 0; i < history.message.length; i++) {
		let frag = (history.side[i] === "1" ? own_msg_template : other_msg_template).content.cloneNode(true);
		let msg  = frag.querySelector(".msg");
		msg.innerText = history.message[i];
		messages.append(msg);
	}
	messages.scrollTop = messages.scrollTopMax;
})();

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
let socket_err = false;
socket.addEventListener("error", (e) => {
	alert("websocket error");
	socket_err = true;
});

// Listen for closing message
socket.addEventListener("close", (e) => {
	if (!socket_err) alert("server closed websocket"); // dont alert again if err-popup was already shown.
});



function add_msg(text, from_self) {
	let frag = (from_self ? own_msg_template : other_msg_template).content.cloneNode(true);
	let msg  = frag.querySelector(".msg");
	msg.innerText = text;

	const scroll_down = messages.scrollTop === messages.scrollTopMax || from_self; // already at bottom, or from self
	/* must be AFTER scroll_down definition */ messages.append(msg);

    if (scroll_down) messages.scrollTop = messages.scrollTopMax;
}

function send_msg(text) {
	socket.send(text);
}

function get_search_params() {
  const key_value_pairs = location.search.substring(1).split("&").map(kv_str => kv_str.split("="));
  const obj = {};
  for (const pair of key_value_pairs) {
    obj[pair[0]] = pair[1];
  }
  return obj;
}

async function check_response(response) {
	if (response instanceof Error) {
		set_notif(`JS Exception: ${response}`, "red");
		return;
	}
	if (!response.ok) {
		const type = response.type != "basic" ? ` (type ${response.type})` : "";
		alert(`${response.status} (${response.statusText})${type}: ${await response.text()}`, "red");
		location.href = "/login/";
		return;
	}

	return true;
}