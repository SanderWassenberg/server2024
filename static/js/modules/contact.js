import * as util from "./util.js";
import { Captcha } from "./captcha.js"; // though unused, this orders initialization of modules such that the Captcha class is initialized before this code is, which is required for the onsubmit callback tobe properly set

const form    = document.querySelector("#contact_form");
const spinner = form.querySelector(".spinner");
const subject = form.querySelector("#subject");
const email   = form.querySelector("#email");
const message = form.querySelector("#message");
const notif   = form.querySelector("#notif");
const submit  = form.querySelector("#submit_btn");
const captcha = form.querySelector("my-captcha");

captcha.onsubmit = () => submit.click();
// form.submit() doesn't work, gives an exception titled NS_ERROR_FAILURE, no other information. Can't find online what exactly is happening. This works.

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
	notif.classList.remove("wiggle");
	util.hide(notif, false);
	notif.innerText = text;
	notif.style.color = color;
	if (color === "red") {
		void notif.offsetWidth; // See: https://css-tricks.com/restart-css-animation/
		notif.classList.add("wiggle");
	}
}

async function send_form(e) {
	e.preventDefault();

	let data = new FormData(form);

	// JS works with utf16 strings, so lengths are not actually representative of the amount of characters. "😃" is counted as length 2.
	// On the severside we count the actual utf8 characters, we don't do that here because the maxlength attribute on <input> elements also uses utf16,
	// and it is more intuitive to the user to keep this check the same as that one. The maxlength html tag is quite easy to remove with the inspector though,
	// which is why we still do the checks here in js. 3 layers of defense!
	function validate_length(elem, name, len_limit) {
		const str = data.get(name);
		if (str.length <= len_limit) return true;
		set_notif(`Subject length is ${str.length}, may at most be ${len_limit}`, "red");
		elem.focus();
	}

	if (!validate_length(subject, "subject", 200)) return;
	if (!validate_length(email,   "email",   100)) return;
	if (!validate_length(message, "message", 600)) return;

	if (!util.is_valid_email(data.get("email"))) {
		set_notif("Invalid email address", "red");
		email.focus();
		return;
	}

	if (!captcha.verify(err => {
		set_notif(err, "red");
		captcha.focus();
	})) return;

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

