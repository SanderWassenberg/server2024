import { api_post, hide } from "./util.js";
import * as gdpr from "./gdpr-popup.js";

const form       = document.querySelector("#signup_form");
const btn        = document.querySelector("#signup_btn");
const spinner    = document.querySelector(".spinner");
const notif      = document.querySelector("#notif");
const gdpr_popup = document.querySelector("gdpr-popup");


form.addEventListener("submit", async e => {

	let data = new FormData(form);

	hide(notif, true);
	hide(spinner, false); // start loading

	console.log("sending", data);
	const response = await api_post('/signup', data);
	console.log("response:", response);

	hide(spinner, true); // stop loading


	if (response instanceof Error) {
		set_notif(`JS Exception: ${response}`, "red");
		return;
	}
	if (!response.ok) {
		const type = response.type != "basic" ? ` (type ${response.type})` : "";
		set_notif(`${response.status} (${response.statusText})${type}: ${await response.text()}`, "red");
		return;
	}

	location.href = "/login"
});

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
