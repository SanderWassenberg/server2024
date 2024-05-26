import * as util from "./util.js";

const form    = document.querySelector("#contact_form");
const spinner = form.querySelector(".spinner");
const notif   = form.querySelector("#notif");
const submit  = form.querySelector("#submit");

function start_loading() {
	for (let i = 0; i < form.elements.length; i++) { form.elements[i].disabled = true; }
	util.hide(spinner, false);
	set_notif("Waiting for server response...", "black");
}

function stop_loading() {
	for (let i = 0; i < form.elements.length; i++) { form.elements[i].disabled = false; }
	util.hide(spinner, true);
	util.hide(notif);
}

function set_notif(text, color) {
	util.hide(notif, false);
	notif.innerText = text;
	notif.style.color = color;
	if (color === "red") {
		notif.classList.remove("wiggle");
		void notif.offsetWidth; // See: https://css-tricks.com/restart-css-animation/
		notif.classList.add("wiggle");
	}
}

// returns an error string or nothing.
function verify(data) {
	// JS works with utf16 strings, so lengths are not actually representative of the amount of characters. "😃" is counted as length 2.
	// On the sever we count the actual utf8 characters, we don't do that here because the maxlength attribute on <input> elements also uses utf16,
	// so we keep this check the same as that one. The html tag is quite easy to remove with the inspector though, which is why also do the checks here. 3 layers of defense!
	{	const len_limit = 200;
		const str = data.get("subject");
		if (str.length > len_limit) return `Subject length is ${str.length}, may at most be ${len_limit}`;
	}

	{	const len_limit = 600;
		const str = data.get("message");
		if (str.length > len_limit) return `Message length is ${str.length}, may at most be ${len_limit}`;
	}

	{	const len_limit = 100;
		const str = data.get("email");
		if (str.length > len_limit) return `Message length is ${str.length}, may at most be ${len_limit}`;

		if (!util.is_valid_email(str)) {
			return "Invalid email address.";
		}
	}
}

async function send_form(e) {
	e.preventDefault();

	let data = new FormData(form);

	{	const err = verify(data);
		if (err) {
			set_notif(err, "red");
			return;
		}
	}

	start_loading();
	// NOTE: This also disables form elements. FormData ignores disabled elements,
	//       so it's important to get the formdata before calling this.

	console.log("sending", data);
	const response = await util.api_post('/contact', data);
	console.log("response:", response);

	stop_loading();

	if (response instanceof Error) {
		set_notif(`JS Exception: ${response}`, "red");
		return;
	}

	if (!response.ok) {
		const type = response.type != "basic" ? ` (type ${response.type})` : "";
		set_notif(`${response.status} (${response.statusText})${type}: ${await response.text()}`, "red");
		return;
	}

	// Clear textboxes
	for (let i = 0; i < form.elements.length; i++) { form.elements[i].value = ""; }

	const type = response.type != "basic" ? ` (type ${response.type})` : ""
	set_notif(`${response.status} (${response.statusText})${type}: ${await response.text()}`, "green");

}


form.addEventListener("submit", send_form);

