import { api_post, hide } from "./util.js";
import * as gdpr from "./gdpr-popup.js";

const form       = document.querySelector("#login_form");
const btn        = document.querySelector("#login_btn");
const spinner    = document.querySelector(".spinner");
const notif      = document.querySelector("#notif");
const gdpr_popup = document.querySelector("gdpr-popup");


form.addEventListener("submit", async e => {

	// login werkt alleen met noodzakelijke cookies
	if (gdpr.consent != "yes" && gdpr.consent != "necessary") {
		hide(gdpr_popup, false);
		set_notif("De loginfunctie onthoudt dat je ingelogd bent d.m.v. cookies. Zonder cookies kan je dus niet inloggen.", "red");
		return;
	}

	let data = new FormData(form);


	hide(notif, true);
	hide(spinner, false); // start loading

	console.log("sending", data);
	const response = await api_post('/login', data);
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

	const session_token = await response.text();
	if (session_token.length < 24) {
		set_notif(`Malformed session token.`, "red");
		return;
	}

	document.cookie = `session_token=${session_token};`;

	location.href = "/search"

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
