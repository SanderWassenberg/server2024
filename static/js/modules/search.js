import { api_post, hide } from "./util.js";

const notif         = document.querySelector("#notif");
const interest_form = document.querySelector("#interest_form");
const your_interest = document.querySelector("#your_interest");
const username      = document.querySelector("#username");

at_page_start();

async function at_page_start() {
	let response = await api_post("/user_info", "");
	if (!await check_response(response)) return;

	let info = await response.json();
	document.documentElement.dataset["role"] = info.role;

	your_interest.value = info.interest;
	username.innerText = info.name;
}

interest_form.addEventListener("submit", async (e) => {
	const response = await api_post("/set_interest", your_interest.value);
	if (!await check_response(response)) return;
	set_notif(`interest updated to ${your_interest.value}`, "green");
});



async function check_response(response) {
	if (response instanceof Error) {
		set_notif(`JS Exception: ${response}`, "red");
		return;
	}
	if (!response.ok) {
		const type = response.type != "basic" ? ` (type ${response.type})` : "";
		set_notif(`${response.status} (${response.statusText})${type}: ${await response.text()}`, "red");

		if (response.status === 401 /* unauthorized */) {
			setTimeout(() => {
				location.href = "/login/";
			}, 2000);
		}
		return;
	}

	return true;
}

function set_notif(text, color) {
	notif.classList.remove("wiggle");
	hide(notif, false);
	notif.innerText = text;
	notif.style.color = color;
	if (color === "red") {
		void notif.offsetWidth; // See: https://css-tricks.com/restart-css-animation/
		notif.classList.add("wiggle");
	}
}
